# QCat

QCat is a modern alternative to netcat with QUIC, TCP, and UDP support. The primary purpose is to learn about network programming in Go and analyze network packets via tools like Wireshark.

## Features

- Support for QUIC protocol (default)
- TCP protocol support
- UDP protocol support
- Listen and connect modes
- Simple, easy-to-understand codebase for educational purposes

## Installation

```shell
go get -u github.com/yourusername/qcat
```

Or clone and build from source:

```shell
git clone https://github.com/yourusername/qcat.git
cd qcat
go build -o qcat
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

## Contributing

Contributions are welcome! Feel free to submit a Pull Request.
