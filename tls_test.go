package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateSelfSignedCert(t *testing.T) {
	certPEM, keyPEM, err := generateSelfSignedCert()
	if err != nil {
		t.Fatalf("generateSelfSignedCert failed: %v", err)
	}

	if len(certPEM) == 0 {
		t.Error("certPEM is empty")
	}
	if len(keyPEM) == 0 {
		t.Error("keyPEM is empty")
	}

	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		t.Error("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}

	if cert.Subject.CommonName != "localhost" {
		t.Errorf("expected CN=localhost, got %s", cert.Subject.CommonName)
	}

	foundLocalhost := false
	for _, name := range cert.DNSNames {
		if name == "localhost" {
			foundLocalhost = true
			break
		}
	}
	if !foundLocalhost {
		t.Error("localhost not found in DNS names")
	}
}

func TestCertCanBeLoadedByTLS(t *testing.T) {
	certPEM, keyPEM, err := generateSelfSignedCert()
	if err != nil {
		t.Fatalf("generateSelfSignedCert failed: %v", err)
	}

	_, err = tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("certificate/key pair invalid: %v", err)
	}
}

func TestSaveCert(t *testing.T) {
	tmpDir := t.TempDir()
	origConfigDir := configDir
	configDir = func() string { return tmpDir }
	t.Cleanup(func() { configDir = origConfigDir })

	certPEM, keyPEM, err := generateSelfSignedCert()
	if err != nil {
		t.Fatalf("generateSelfSignedCert failed: %v", err)
	}

	if err := saveCert(certPEM, keyPEM); err != nil {
		t.Fatalf("saveCert failed: %v", err)
	}

	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	if _, err := os.Stat(certPath); err != nil {
		t.Errorf("cert.pem not created: %v", err)
	}
	if _, err := os.Stat(keyPath); err != nil {
		t.Errorf("key.pem not created: %v", err)
	}

	keyInfo, _ := os.Stat(keyPath)
	if keyInfo.Mode().Perm() != 0600 {
		t.Errorf("key.pem has wrong permissions: %v", keyInfo.Mode().Perm())
	}
}

func TestCertExists(t *testing.T) {
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	if certExists(certPath, keyPath) {
		t.Error("certExists returned true for non-existent files")
	}

	os.WriteFile(certPath, []byte("cert"), 0644)
	if certExists(certPath, keyPath) {
		t.Error("certExists returned true when only cert exists")
	}

	os.WriteFile(keyPath, []byte("key"), 0600)
	if !certExists(certPath, keyPath) {
		t.Error("certExists returned false when both files exist")
	}
}

func TestGetLocalIPs(t *testing.T) {
	ips := getLocalIPs()
	for _, ip := range ips {
		if ip.IsLoopback() {
			t.Error("getLocalIPs returned loopback address")
		}
		if ip.To4() == nil {
			t.Error("getLocalIPs returned non-IPv4 address")
		}
	}
}

func TestLoadOrGenerateCert(t *testing.T) {
	tmpDir := t.TempDir()
	origConfigDir := configDir
	configDir = func() string { return tmpDir }
	t.Cleanup(func() { configDir = origConfigDir })

	t.Run("generates new cert when none exists", func(t *testing.T) {
		certPath, keyPath, err := loadOrGenerateCert()
		if err != nil {
			t.Fatalf("loadOrGenerateCert failed: %v", err)
		}

		if certPath == "" || keyPath == "" {
			t.Error("expected non-empty cert and key paths")
		}

		if _, err := os.Stat(certPath); err != nil {
			t.Errorf("cert file not created: %v", err)
		}
		if _, err := os.Stat(keyPath); err != nil {
			t.Errorf("key file not created: %v", err)
		}
	})

	t.Run("reuses existing cert", func(t *testing.T) {
		certPath1, keyPath1, _ := loadOrGenerateCert()
		certPath2, keyPath2, err := loadOrGenerateCert()
		if err != nil {
			t.Fatalf("loadOrGenerateCert failed: %v", err)
		}

		if certPath1 != certPath2 || keyPath1 != keyPath2 {
			t.Error("expected same paths for existing cert")
		}
	})
}

func TestCertPaths(t *testing.T) {
	tmpDir := t.TempDir()
	origConfigDir := configDir
	configDir = func() string { return tmpDir }
	t.Cleanup(func() { configDir = origConfigDir })

	certPath, keyPath := certPaths()
	if certPath != filepath.Join(tmpDir, "cert.pem") {
		t.Errorf("unexpected cert path: %s", certPath)
	}
	if keyPath != filepath.Join(tmpDir, "key.pem") {
		t.Errorf("unexpected key path: %s", keyPath)
	}
}
