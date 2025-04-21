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

	listener, err := quic.ListenAddr(s.config.Address, generateTLSConfig(), nil)
	if err != nil {
		return fmt.Errorf("failed to start QUIC server: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept(ctx)
		if err != nil {
			return fmt.Errorf("failed to accept connection: %v", err)
		}

		stream, err := conn.AcceptStream(ctx)
		if err != nil {
			log.Printf("Failed to accept stream: %v", err)
			continue
		}

		go s.handleConnection(stream)
	}
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

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept connection: %v", err)
		}

		go s.handleConnection(conn)
	}
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
	defer conn.Close()

	// Simple handling for UDP (note: UDP is connectionless)
	buffer := make([]byte, 1024)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			return fmt.Errorf("error reading from UDP: %v", err)
		}

		if s.config.Verbose {
			log.Printf("Received %d bytes from %s", n, clientAddr)
		}

		os.Stdout.Write(buffer[:n])
		
		// Echo back what was received
		_, err = conn.WriteToUDP(buffer[:n], clientAddr)
		if err != nil {
			log.Printf("Failed to write to UDP: %v", err)
		}
	}
}

func (s *Server) handleConnection(conn io.ReadWriteCloser) {
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

// generateTLSConfig generates a TLS config for QUIC
func generateTLSConfig() *quic.Config {
	return &quic.Config{}
}
