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
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer conn.CloseWithError(0, "")

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		return fmt.Errorf("failed to open stream: %v", err)
	}
	defer stream.Close()

	return c.handleConnection(stream)
}

func (c *Client) connectTCP(ctx context.Context) error {
	if c.config.Verbose {
		log.Printf("Connecting to %s using TCP", c.config.Address)
	}

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", c.config.Address)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
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
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer conn.Close()

	return c.handleConnection(conn)
}

func (c *Client) handleConnection(conn io.ReadWriteCloser) error {
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
