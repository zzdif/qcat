package common

import (
	"io"
	"log"
)

type verboseConn struct {
	conn    io.ReadWriteCloser
	verbose bool
	role    string // "client" or "server"
}

func (v *verboseConn) Read(p []byte) (n int, err error) {
	n, err = v.conn.Read(p)
	if n > 0 && v.verbose {
		log.Printf("[%s] Received %d bytes", v.role, n)
		maxDump := 32
		if n < maxDump {
			maxDump = n
		}
		log.Printf("[%s] Data (hex): % x", v.role, p[:maxDump])
		// TODO: parse TCP/UDP headers here and log relevant header fields
	}
	return
}

func (v *verboseConn) Write(p []byte) (n int, err error) {
	n, err = v.conn.Write(p)
	if v.verbose {
		log.Printf("[%s] Sent %d bytes", v.role, n)
	}
	return n, err
}

func (v *verboseConn) Close() error {
	return v.conn.Close()
}
