package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

// escribirLlaveDePrueba genera una llave RSA de 2048 bits, la guarda como PEM PKCS#8
// en un directorio temporal y devuelve su ruta.
func escribirLlaveDePrueba(t *testing.T) string {
	t.Helper()
	llave, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("no se pudo generar la llave: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(llave)
	if err != nil {
		t.Fatalf("no se pudo serializar la llave: %v", err)
	}
	ruta := filepath.Join(t.TempDir(), "jwt.pem")
	contenido := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	if err := os.WriteFile(ruta, contenido, 0o600); err != nil {
		t.Fatalf("no se pudo escribir la llave: %v", err)
	}
	return ruta
}
