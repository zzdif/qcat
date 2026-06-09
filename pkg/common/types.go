package common

import (
	"fmt"
	"net"
	"time"
)

// Protocol represents network protocol type
type Protocol string

const (
	// QUIC protocol
	QUIC Protocol = "quic"
	// TCP protocol
	TCP Protocol = "tcp"
	// UDP protocol
	UDP Protocol = "udp"
)

// ValidProtocols returns the set of supported protocol values.
func ValidProtocols() []Protocol {
	return []Protocol{QUIC, TCP, UDP}
}

// IsValid returns whether the protocol is a supported value.
func (p Protocol) IsValid() bool {
	for _, valid := range ValidProtocols() {
		if p == valid {
			return true
		}
	}
	return false
}

// ValidateAddress checks that an address string is a valid host:port.
func ValidateAddress(addr string) error {
	_, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid address %q: %w", addr, err)
	}
	return nil
}

// Config holds common configuration options
type Config struct {
	// Protocol to use
	Protocol Protocol

	// Address to connect to or listen on
	Address string

	// Verbose output
	Verbose bool

	// IdleTimeout is the QUIC idle timeout (duration). Zero means use default.
	IdleTimeout time.Duration

	// TLSCert is the path to a PEM-encoded TLS certificate file (server mode).
	TLSCert string

	// TLSKey is the path to a PEM-encoded TLS private key file (server mode).
	TLSKey string

	// TLSCA is the path to a PEM-encoded CA certificate for server verification (client mode).
	TLSCA string

	// TLSInsecure skips TLS certificate verification (client mode).
	TLSInsecure bool
}