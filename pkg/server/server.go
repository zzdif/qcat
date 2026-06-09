package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"github.com/quic-go/quic-go"
	"qcat/pkg/common"
)

// Server represents a qcat server
type Server struct {
	config common.Config
}

// New creates a new server instance
func New(config common.Config) *Server {
	return &Server{
		config: config,
	}
}

// udpConnWrapper wraps a UDP connection to implement io.ReadWriteCloser for interactive sessions.
// It records the last client address seen on Read and writes to that address on Write.
// Note: responses always go to whichever client sent the most recent datagram,
// matching netcat-udp behavior.
type udpConnWrapper struct {
	conn       *net.UDPConn
	clientAddr *net.UDPAddr
	mu         sync.RWMutex
}

// Read reads a datagram, stores the client address, and returns the payload.
func (u *udpConnWrapper) Read(p []byte) (int, error) {
	n, addr, err := u.conn.ReadFromUDP(p)
	if err != nil {
		return n, err
	}
	u.mu.Lock()
	u.clientAddr = addr
	u.mu.Unlock()
	return n, nil
}

// Write sends data to the last known client address.
func (u *udpConnWrapper) Write(p []byte) (int, error) {
	u.mu.RLock()
	addr := u.clientAddr
	u.mu.RUnlock()
	if addr == nil {
		return 0, fmt.Errorf("no UDP client address to write to")
	}
	return u.conn.WriteToUDP(p, addr)
}

// Close closes the underlying UDP connection.
func (u *udpConnWrapper) Close() error {
	return u.conn.Close()
}

// halfCloser is implemented by connections that support half-close (e.g., *net.TCPConn).
type halfCloser interface {
	CloseWrite() error
}

// quicStreamHalfCloser wraps a quic stream to support half-close.
type quicStreamHalfCloser struct {
	stream quic.Stream
}

func (q *quicStreamHalfCloser) CloseWrite() error {
	return q.stream.Close()
}

// Start starts the server
func (s *Server) Start(ctx context.Context) error {
	switch s.config.Protocol {
	case common.QUIC:
		return s.startQUIC(ctx)
	case common.TCP:
		return s.startTCP(ctx)
	case common.UDP:
		return s.startUDP(ctx)
	default:
		return fmt.Errorf("unsupported protocol: %s", s.config.Protocol)
	}
}

func (s *Server) startQUIC(ctx context.Context) error {
	if s.config.Verbose {
		log.Printf("Starting QUIC server on %s", s.config.Address)
	}

	// Generate or load TLS configuration
	var tlsCfg *tls.Config
	var err error
	if s.config.TLSCert != "" && s.config.TLSKey != "" {
		tlsCfg, err = common.LoadOrGenerateTLSConfig(s.config.TLSCert, s.config.TLSKey)
	} else {
		tlsCfg, err = common.GenerateTLSConfig()
	}
	if err != nil {
		return fmt.Errorf("failed to configure TLS: %v", err)
	}

	// Configure QUIC idle timeout if set
	var quicCfg *quic.Config
	if s.config.IdleTimeout > 0 {
		quicCfg = &quic.Config{MaxIdleTimeout: s.config.IdleTimeout}
	}
	listener, err := quic.ListenAddr(s.config.Address, tlsCfg, quicCfg)
	if err != nil {
		return fmt.Errorf("failed to start QUIC server: %v", err)
	}
	defer listener.Close()

	// Accept a single QUIC connection
	conn, err := listener.Accept(ctx)
	if err != nil {
		return fmt.Errorf("failed to accept connection: %v", err)
	}
	defer conn.CloseWithError(0, "")
	if s.config.Verbose {
		log.Printf("Accepted QUIC connection from %s", conn.RemoteAddr())
	}

	// Accept a bidirectional stream
	stream, err := conn.AcceptStream(ctx)
	if err != nil {
		return fmt.Errorf("failed to accept stream: %v", err)
	}
	if s.config.Verbose {
		log.Printf("Accepted QUIC stream %d", stream.StreamID())
	}

	// Handle the stream until stdin EOF or remote close
	s.handleConnection(stream, &quicStreamHalfCloser{stream: stream})

	return nil
}

func (s *Server) startTCP(ctx context.Context) error {
	if s.config.Verbose {
		log.Printf("Starting TCP server on %s", s.config.Address)
	}

	listener, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		return fmt.Errorf("failed to start TCP server: %v", err)
	}
	defer listener.Close()

	// Accept a single TCP connection
	conn, err := listener.Accept()
	if err != nil {
		return fmt.Errorf("failed to accept connection: %v", err)
	}
	defer conn.Close()
	if s.config.Verbose {
		log.Printf("Accepted TCP connection from %s", conn.RemoteAddr())
	}

	// Handle connection until stdin EOF or remote close
	// *net.TCPConn implements CloseWrite for half-close on stdin EOF
	var hc halfCloser
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		hc = tcpConn
	}
	s.handleConnection(conn, hc)

	return nil
}

func (s *Server) startUDP(ctx context.Context) error {
	if s.config.Verbose {
		log.Printf("Starting UDP server on %s", s.config.Address)
	}

	addr, err := net.ResolveUDPAddr("udp", s.config.Address)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to start UDP server: %v", err)
	}
	// Wrap UDPConn for interactive session; handleConnection will close when done
	wrapper := &udpConnWrapper{conn: conn}
	s.handleConnection(wrapper, nil)

	return nil
}

// handleConnection manages bidirectional data flow between stdin and the network connection.
// halfClose is called on stdin EOF to signal the remote side (e.g., TCP FIN or QUIC stream FIN).
func (s *Server) handleConnection(conn io.ReadWriteCloser, halfClose halfCloser) {
	if s.config.Verbose {
		conn = &common.VerboseConn{
			Conn:    conn,
			Verbose: true,
			Role:    "server",
		}
	}
	defer conn.Close()

	// Copy received data to stdout
	go func() {
		if _, err := io.Copy(os.Stdout, conn); err != nil {
			if s.config.Verbose {
				log.Printf("Error reading from connection: %v", err)
			}
		}
	}()

	// Copy stdin to connection
	if _, err := io.Copy(conn, os.Stdin); err != nil {
		if s.config.Verbose {
			log.Printf("Error writing to connection: %v", err)
		}
	}

	// Half-close: signal EOF to remote side when stdin is done
	if halfClose != nil {
		_ = halfClose.CloseWrite()
	}
}