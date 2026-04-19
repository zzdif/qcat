package common

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"time"
)

// GenerateTLSConfig creates a TLS configuration with a self-signed certificate.
func GenerateTLSConfig() (*tls.Config, error) {
	cert, err := GenerateSelfSignedCert()
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"qcat"},
	}, nil
}

// LoadOrGenerateTLSConfig loads a TLS certificate from certFile/keyFile if they exist,
// or generates a new self-signed certificate and saves it to those paths.
func LoadOrGenerateTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err == nil {
		return &tls.Config{
			MinVersion:   tls.VersionTLS13,
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"qcat"},
		}, nil
	}

	// Generate new certificate
	tlsCert, err := GenerateSelfSignedCert()
	if err != nil {
		return nil, err
	}

	// Save certificate to file
	certDER := tlsCert.Certificate[0]
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err := os.WriteFile(certFile, certPEM, 0644); err != nil {
		return nil, err
	}

	// Save private key to file
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(tlsCert.PrivateKey.(*rsa.PrivateKey))})
	if err := os.WriteFile(keyFile, keyPEM, 0600); err != nil {
		return nil, err
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"qcat"},
	}, nil
}

// LoadClientTLSConfig creates a TLS configuration for client connections.
// If caFile is provided, the CA certificate is loaded for server verification.
// If insecure is true, certificate verification is skipped (for testing only).
func LoadClientTLSConfig(caFile string, insecure bool) (*tls.Config, error) {
	if insecure {
		return &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS13,
			NextProtos:         []string{"qcat"},
		}, nil
	}

	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS13,
		NextProtos: []string{"qcat"},
	}

	if caFile != "" {
		caPEM, err := os.ReadFile(caFile)
		if err != nil {
			return nil, err
		}
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caPEM) {
			return nil, err
		}
		tlsCfg.RootCAs = certPool
	} else {
		// Use system cert pool
		systemPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		tlsCfg.RootCAs = systemPool
	}

	return tlsCfg, nil
}

// GenerateSelfSignedCert generates a self-signed TLS certificate.
func GenerateSelfSignedCert() (tls.Certificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Random serial number per RFC 5280 §4.1.2.2
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}

	now := time.Now()

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"QCat"},
		},
		NotBefore:             now,
		NotAfter:              now.Add(time.Hour * 24 * 180), // Valid for 180 days
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:              []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	cert := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  key,
	}

	return cert, nil
}