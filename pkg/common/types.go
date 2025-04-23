package common

import "time"

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
}
