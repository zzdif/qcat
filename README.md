# QCat

QCat is a modern alternative to netcat with QUIC, TCP, and UDP support. The primary purpose is to learn about network programming in Go and analyze network packets via tools like Wireshark.

## Features

- Support for QUIC protocol (default)
- TCP protocol support
- UDP protocol support
- Listen and connect modes
- Simple, easy-to-understand codebase for educational purposes

## Installation

### Prerequisites

- Go 1.18 or higher installed (https://golang.org/dl/)
- Git (for building from source)

### Install via go install

```shell
go install github.com/zzdif/qcat@latest
```

Ensure `$GOPATH/bin` or `$HOME/go/bin` is in your `PATH` to use the installed binary.

### Build from source

```shell
git clone https://github.com/zzdif/qcat.git
cd qcat
go build -o qcat
```

Optionally, move the binary to a directory in your `PATH`:

```shell
mv qcat $HOME/go/bin/
```

## Usage

### Listen mode (server)

```shell
# Listen on port 8000 with QUIC protocol (default)
./qcat listen -l :8000

# Listen using TCP
./qcat listen -l :8000 -p tcp

# Listen using UDP
./qcat listen -l :8000 -p udp
```

### Idle Timeout (QUIC)
Use the `--idle-timeout` flag to specify a custom QUIC idle timeout (duration).
```shell
# Listen with custom QUIC idle timeout (e.g., 2 minutes)
./qcat listen -l :8000 --idle-timeout 2m
```

### Connect mode (client)

```shell
# Connect to a QUIC server
./qcat connect -c example.com:8000

# Connect using TCP
./qcat connect -c example.com:8000 -p tcp

# Connect using UDP
./qcat connect -c example.com:8000 -p udp
```

### Verbose Mode

Add `-v` flag for verbose output:

```shell
./qcat listen -l :8000 -v
```

## Learning Network Protocols

This tool is primarily designed for learning purposes. You can use Wireshark to capture and analyze packets while QCat is running to learn about:

1. QUIC protocol handshake and packet structure
2. Differences between TCP and UDP packets
3. Network packet analysis
4. Connection establishment and termination

## License

MIT
