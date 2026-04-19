package client

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

// Client represents a qcat client
type Client struct {
	config common.Config
}

// New creates a new client instance
func New(config common.Config) *Client {
	return &Client{
		config: config,
	}
}

// Connect connects to a server
func (c *Client) Connect(ctx context.Context) error {
	switch c.config.Protocol {
	case common.QUIC:
		return c.connectQUIC(ctx)
	case common.TCP:
		return c.connectTCP(ctx)
	case common.UDP:
		return c.connectUDP(ctx)
	default:
		return fmt.Errorf("unsupported protocol: %s", c.config.Protocol)
	}
}

func (c *Client) connectQUIC(ctx context.Context) error {
	if c.config.Verbose {
		log.Printf("Connecting to %s using QUIC", c.config.Address)
	}

	tlsConfig, err := common.LoadClientTLSConfig(c.config.TLSCA, c.config.TLSInsecure)
	if err != nil {
		return fmt.Errorf("failed to configure TLS: %v", err)
	}

	conn, err := quic.DialAddr(ctx, c.config.Address, tlsConfig, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to %s via QUIC: %v", c.config.Address, err)
	}
	defer conn.CloseWithError(0, "")

	if c.config.Verbose {
		log.Printf("Connected to QUIC server at %s", conn.RemoteAddr())
	}

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		return fmt.Errorf("failed to open stream: %v", err)
	}
	defer stream.Close()

	if c.config.Verbose {
		log.Printf("Opened QUIC stream %d", stream.StreamID())
	}

	return c.handleConnection(stream, &quicStreamHalfCloser{stream: stream})
}

func (c *Client) connectTCP(ctx context.Context) error {
	if c.config.Verbose {
		log.Printf("Connecting to %s using TCP", c.config.Address)
	}

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", c.config.Address)
	if err != nil {
		return fmt.Errorf("failed to connect to %s via TCP: %v", c.config.Address, err)
	}
	defer conn.Close()

	if c.config.Verbose {
		log.Printf("Connected to TCP server at %s", conn.RemoteAddr())
	}

	// *net.TCPConn implements CloseWrite for half-close on stdin EOF
	var hc halfCloser
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		hc = tcpConn
	}
	return c.handleConnection(conn, hc)
}

func (c *Client) connectUDP(ctx context.Context) error {
	if c.config.Verbose {
		log.Printf("Connecting to %s using UDP", c.config.Address)
	}

	addr, err := net.ResolveUDPAddr("udp", c.config.Address)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("failed to connect to %s via UDP: %v", c.config.Address, err)
	}
	defer conn.Close()

	if c.config.Verbose {
		log.Printf("Connected to UDP server at %s", conn.RemoteAddr())
	}

	return c.handleConnection(conn, nil)
}

// halfCloser is implemented by connections that support half-close (e.g., *net.TCPConn).
type halfCloser interface {
	CloseWrite() error
}

// quicStreamHalfCloser wraps a quic stream to support half-close via Stream.Close().
type quicStreamHalfCloser struct {
	stream quic.Stream
}

func (q *quicStreamHalfCloser) CloseWrite() error {
	return q.stream.Close()
}

// handleConnection manages bidirectional data flow between stdin and the network connection.
// halfClose is called on stdin EOF to signal the remote side.
func (c *Client) handleConnection(conn io.ReadWriteCloser, halfCloser halfCloser) error {
	if c.config.Verbose {
		conn = &common.VerboseConn{
			Conn:    conn,
			Verbose: true,
			Role:    "client",
		}
	}
	defer conn.Close()

	// Copy stdin to connection; on EOF, half-close the write side.
	go func() {
		io.Copy(conn, os.Stdin)
		// Half-close: signal EOF to remote side when stdin is done.
		if halfCloser != nil {
			_ = halfCloser.CloseWrite()
		}
	}()

	// Copy connection to stdout
	_, err := io.Copy(os.Stdout, conn)
	if err != nil {
		if c.config.Verbose {
			log.Printf("Error reading from connection: %v", err)
		}
	}

	return nil
}