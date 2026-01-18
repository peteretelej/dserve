package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

var configDir = func() string {
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "dserve")
}

func certPaths() (string, string) {
	dir := configDir()
	return filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem")
}

func loadOrGenerateCert() (string, string, error) {
	certPath, keyPath := certPaths()

	if certExists(certPath, keyPath) {
		return certPath, keyPath, nil
	}

	certPEM, keyPEM, err := generateSelfSignedCert()
	if err != nil {
		return "", "", err
	}

	if err := saveCert(certPEM, keyPEM); err != nil {
		return "", "", err
	}

	return certPath, keyPath, nil
}

func certExists(certPath, keyPath string) bool {
	_, err1 := os.Stat(certPath)
	_, err2 := os.Stat(keyPath)
	return err1 == nil && err2 == nil
}

func generateSelfSignedCert() ([]byte, []byte, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	template.IPAddresses = append(template.IPAddresses, getLocalIPs()...)

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return certPEM, keyPEM, nil
}

func saveCert(certPEM, keyPEM []byte) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	certPath, keyPath := certPaths()

	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		return err
	}
	return os.WriteFile(keyPath, keyPEM, 0600)
}

func getLocalIPs() []net.IP {
	var ips []net.IP
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP)
			}
		}
	}
	return ips
}
