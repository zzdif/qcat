package common

import (
	"io"
	"log"
)

type VerboseConn struct {
	Conn    io.ReadWriteCloser
	Verbose bool
	Role    string // "client" or "server"
}

func (v *VerboseConn) Read(p []byte) (n int, err error) {
	n, err = v.Conn.Read(p)
	if n > 0 && v.Verbose {
		log.Printf("[%s] Received %d bytes", v.Role, n)
	}
	return
}

func (v *VerboseConn) Write(p []byte) (n int, err error) {
	n, err = v.Conn.Write(p)
	if v.Verbose {
		log.Printf("[%s] Sent %d bytes", v.Role, n)
	}
	return n, err
}

func (v *VerboseConn) Close() error {
	return v.Conn.Close()
}