package client

import (
	"context"
	"crypto/tls"
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

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"qcat"},
	}

	conn, err := quic.DialAddr(ctx, c.config.Address, tlsConfig, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to %s via QUIC: %v", c.config.Address, err)
	}
	if c.config.Verbose {
		log.Printf("Connected to QUIC server at %s", conn.RemoteAddr())
	}

	// Close QUIC connection on context cancellation
	go func() {
		<-ctx.Done()
		if c.config.Verbose {
			log.Printf("Context canceled, closing QUIC connection")
		}
		_ = conn.CloseWithError(0, "context canceled")
	}()

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		return fmt.Errorf("failed to open stream: %v", err)
	}
	if c.config.Verbose {
		log.Printf("Opened QUIC stream %d", stream.StreamID())
	}
	defer stream.Close()
	defer conn.CloseWithError(0, "")

	return c.handleConnection(stream)
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
	if c.config.Verbose {
		log.Printf("Connected to TCP server at %s", conn.RemoteAddr())
	}

	// Close connection on context cancellation
	go func() {
		<-ctx.Done()
		if c.config.Verbose {
			log.Printf("Context canceled, closing TCP connection")
		}
		_ = conn.Close()
	}()
	defer conn.Close()

	return c.handleConnection(conn)
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
	if c.config.Verbose {
		log.Printf("Connected to UDP server at %s", conn.RemoteAddr())
	}

	// Close connection on context cancellation
	go func() {
		<-ctx.Done()
		if c.config.Verbose {
			log.Printf("Context canceled, closing UDP connection")
		}
		_ = conn.Close()
	}()
	defer conn.Close()

	return c.handleConnection(conn)
}

func (c *Client) handleConnection(conn io.ReadWriteCloser) error {
	// verbose connection case
	if c.config.Verbose {
		conn = &verboseConn{
			conn:    conn,
			verbose: true,
			role:    "client",
		}
	}
	// Copy stdin to connection
	go func() {
		if _, err := io.Copy(conn, os.Stdin); err != nil {
			if c.config.Verbose {
				log.Printf("Error writing to connection: %v", err)
			}
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
