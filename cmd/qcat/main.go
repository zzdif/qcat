package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "qcat",
		Short: "qcat - netcat alternative with QUIC, TCP and UDP support",
		Long: `qcat is a networking utility that reads and writes data across network connections.
It supports QUIC (default), TCP, and UDP protocols and can be used for learning and
analyzing network traffic.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Show help if no commands provided
			cmd.Help()
		},
	}

	// Set up flags
	var (
		proto       string
		listenAddr  string
		connectAddr string
		verbose     bool
	)

	// Common flags
	rootCmd.PersistentFlags().StringVarP(&proto, "protocol", "p", "quic", "Protocol to use: quic, tcp, or udp")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Listen command
	listenCmd := &cobra.Command{
		Use:   "listen",
		Short: "Start qcat in listen (server) mode",
		Long:  `Start qcat in listen mode, waiting for incoming connections.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Listening on %s using %s protocol\n", listenAddr, proto)
			// Call into server package here
		},
	}
	listenCmd.Flags().StringVarP(&listenAddr, "listen", "l", ":8000", "Address to listen on (format: [host]:port)")

	// Connect command
	connectCmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to a remote host",
		Long:  `Connect to a remote host using specified protocol.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Connecting to %s using %s protocol\n", connectAddr, proto)
			// Call into client package here
		},
	}
	connectCmd.Flags().StringVarP(&connectAddr, "connect", "c", "localhost:8000", "Address to connect to (format: host:port)")

	// Add commands
	rootCmd.AddCommand(listenCmd, connectCmd)

	// Execute the CLI
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
