package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"

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
type udpConnWrapper struct {
	conn       *net.UDPConn
	clientAddr *net.UDPAddr
}

// Read reads a datagram, stores the client address, and returns the payload.
func (u *udpConnWrapper) Read(p []byte) (int, error) {
	n, addr, err := u.conn.ReadFromUDP(p)
	if err != nil {
		return n, err
	}
	u.clientAddr = addr
	return n, nil
}

// Write sends data to the last known client address.
func (u *udpConnWrapper) Write(p []byte) (int, error) {
	if u.clientAddr == nil {
		return 0, fmt.Errorf("no UDP client address to write to")
	}
	return u.conn.WriteToUDP(p, u.clientAddr)
}

// Close closes the underlying UDP connection.
func (u *udpConnWrapper) Close() error {
	return u.conn.Close()
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

	// Generate TLS configuration
	tlsCfg, err := common.GenerateTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to generate TLS config: %v", err)
	}
	// Configure QUIC idle timeout if set
	var quicCfg *quic.Config
	if s.config.IdleTimeout > 0 {
		// Set maximum idle timeout after handshake
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
	s.handleConnection(stream)

	// Close QUIC connection gracefully
	_ = conn.CloseWithError(0, "")
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
	if s.config.Verbose {
		log.Printf("Accepted TCP connection from %s", conn.RemoteAddr())
	}

	// Handle connection until stdin EOF or remote close
	s.handleConnection(conn)
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
	s.handleConnection(wrapper)
	return nil
}

func (s *Server) handleConnection(conn io.ReadWriteCloser) {
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
}
