package main

import (
   "context"
   "fmt"
   "os"

   "github.com/spf13/cobra"

   "qcat/pkg/common"
   "qcat/pkg/server"
   "qcat/pkg/client"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "qcat",
		Short: "qcat - netcat alternative with QUIC, TCP and UDP support",
		Long: `qcat is a networking utility that reads and writes data across network connections.
It supports QUIC (default), TCP, and UDP protocols and can be used for learning and
analyzing network traffic.

Examples:
  qcat listen -l :8000
  qcat connect -c example.com:8000 -p tcp`,
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
       Long:  `Start qcat in listen mode, waiting for incoming connections.

Examples:
  qcat listen -l :8000
  qcat listen -l :8000 -p tcp`,
       Run: func(cmd *cobra.Command, args []string) {
           // Configure and start the server
           config := common.Config{
               Protocol: common.Protocol(proto),
               Address:  listenAddr,
               Verbose:  verbose,
           }
           srv := server.New(config)
           if err := srv.Start(context.Background()); err != nil {
               fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
               os.Exit(1)
           }
       },
   }
	listenCmd.Flags().StringVarP(&listenAddr, "listen", "l", ":8000", "Address to listen on (format: [host]:port)")

   // Connect command
   connectCmd := &cobra.Command{
       Use:   "connect",
       Short: "Connect to a remote host",
       Long:  `Connect to a remote host using specified protocol.

Examples:
  qcat connect -c example.com:8000
  qcat connect -c example.com:8000 -p udp`,
       Run: func(cmd *cobra.Command, args []string) {
           // Configure and run the client
           config := common.Config{
               Protocol: common.Protocol(proto),
               Address:  connectAddr,
               Verbose:  verbose,
           }
           cli := client.New(config)
           if err := cli.Connect(context.Background()); err != nil {
               fmt.Fprintf(os.Stderr, "Error connecting to server: %v\n", err)
               os.Exit(1)
           }
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
