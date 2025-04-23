package main

import (
   "context"
   "fmt"
   "os"
   "time"

   "github.com/spf13/cobra"

   "qcat/pkg/client"
   "qcat/pkg/common"
   "qcat/pkg/server"
)

func main() {
   rootCmd := &cobra.Command{
       Use:   "qcat",
       Short: "netcat alternative with QUIC, TCP and UDP support",
       Long: `qcat is a networking utility that reads and writes data across network connections.
It supports QUIC (default), TCP, and UDP protocols and can be used for learning and
analyzing network traffic.

Examples:
  qcat listen -l :8000
  qcat connect -c example.com:8000 -p tcp`,
       Run: func(cmd *cobra.Command, args []string) {
           _ = cmd.Help()
       },
   }

   var (
       proto       string
       listenAddr  string
       connectAddr string
       verbose     bool
       // Idle timeout for QUIC connections; zero means use default
       idleTimeout time.Duration
   )

   rootCmd.PersistentFlags().
       StringVarP(&proto, "protocol", "p", "quic", "Protocol to use: quic, tcp, or udp")
   rootCmd.PersistentFlags().
       BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
   // QUIC idle timeout configuration
   rootCmd.PersistentFlags().
       DurationVar(&idleTimeout, "idle-timeout", 0, "Idle timeout for QUIC connections (e.g. 2m). Zero means default.")

   listenCmd := &cobra.Command{
       Use:   "listen",
       Short: "Start qcat in listen (server) mode",
       Long:  `Start qcat in listen mode, waiting for incoming connections.

Examples:
  qcat listen -l :8000
  qcat listen -l :8000 -p tcp`,
       Run: func(cmd *cobra.Command, args []string) {
           // --idle-timeout is only valid when using QUIC protocol
           if cmd.Flags().Changed("idle-timeout") && common.Protocol(proto) != common.QUIC {
               fmt.Fprintln(os.Stderr, "Error: --idle-timeout flag is only valid with QUIC protocol")
               os.Exit(1)
           }
           config := common.Config{
               Protocol:    common.Protocol(proto),
               Address:     listenAddr,
               Verbose:     verbose,
               IdleTimeout: idleTimeout,
           }
           srv := server.New(config)
           if err := srv.Start(context.Background()); err != nil {
               fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
               os.Exit(1)
           }
       },
   }
   listenCmd.Flags().
       StringVarP(&listenAddr, "listen", "l", ":8000", "Address to listen on (format: [host]:port)")

   connectCmd := &cobra.Command{
       Use:   "connect",
       Short: "Connect to a remote host",
       Long:  `Connect to a remote host using specified protocol.

Examples:
  qcat connect -c example.com:8000
  qcat connect -c example.com:8000 -p udp`,
       Run: func(cmd *cobra.Command, args []string) {
           // idle-timeout is only valid in server (listen) mode
           if cmd.Flags().Changed("idle-timeout") {
               fmt.Fprintln(os.Stderr, "Error: --idle-timeout flag is only valid in listen (server) mode")
               os.Exit(1)
           }
           config := common.Config{
               Protocol:    common.Protocol(proto),
               Address:     connectAddr,
               Verbose:     verbose,
           }
           cli := client.New(config)
           if err := cli.Connect(context.Background()); err != nil {
               fmt.Fprintf(os.Stderr, "Error connecting to server: %v\n", err)
               os.Exit(1)
           }
       },
   }
   connectCmd.Flags().
       StringVarP(&connectAddr, "connect", "c", "localhost:8000", "Address to connect to (format: host:port)")

   rootCmd.AddCommand(listenCmd, connectCmd)

   if err := rootCmd.Execute(); err != nil {
       fmt.Fprintln(os.Stderr, err)
       os.Exit(1)
   }
}
