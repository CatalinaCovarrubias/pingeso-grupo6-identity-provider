// Paquete tokens firma y valida los JWT del servicio de identidad (RS256) y publica
// la llave pública en formato JWKS. No conoce HTTP ni la base de datos.
package tokens

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
)

// bitsMinimos es el tamaño mínimo aceptado para la llave RSA de firma.
const bitsMinimos = 2048

// Llave es la llave privada RSA con la que se firman los tokens, junto con su kid.
type Llave struct {
	privada *rsa.PrivateKey
	kid     string
}

// NuevaLlave valida el tamaño de la llave y calcula su kid.
func NuevaLlave(privada *rsa.PrivateKey) (*Llave, error) {
	if privada == nil {
		return nil, errors.New("la llave privada es nula")
	}
	if privada.N.BitLen() < bitsMinimos {
		return nil, fmt.Errorf("la llave RSA debe tener al menos %d bits, tiene %d", bitsMinimos, privada.N.BitLen())
	}
	if err := privada.Validate(); err != nil {
		return nil, fmt.Errorf("la llave RSA no es válida: %w", err)
	}
	return &Llave{privada: privada, kid: huella(&privada.PublicKey)}, nil
}

// GenerarLlave crea una llave RSA de 2048 bits en memoria. Solo se usa en desarrollo:
// los tokens firmados con ella dejan de valer cuando el proceso se reinicia.
func GenerarLlave() (*Llave, error) {
	privada, err := rsa.GenerateKey(rand.Reader, bitsMinimos)
	if err != nil {
		return nil, fmt.Errorf("no se pudo generar la llave RSA: %w", err)
	}
	return NuevaLlave(privada)
}

// CargarLlave lee una llave RSA desde un archivo PEM, en formato PKCS#8 ("PRIVATE KEY")
// o PKCS#1 ("RSA PRIVATE KEY"). El error nunca incluye el contenido del archivo.
func CargarLlave(ruta string) (*Llave, error) {
	contenido, err := os.ReadFile(ruta)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer la llave JWT en %s: %w", ruta, err)
	}
	bloque, _ := pem.Decode(contenido)
	if bloque == nil {
		return nil, fmt.Errorf("el archivo %s no contiene un bloque PEM", ruta)
	}

	var privada *rsa.PrivateKey
	switch bloque.Type {
	case "PRIVATE KEY":
		llave, err := x509.ParsePKCS8PrivateKey(bloque.Bytes)
		if err != nil {
			return nil, fmt.Errorf("la llave PKCS#8 en %s no es válida", ruta)
		}
		rsaLlave, ok := llave.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("la llave en %s no es RSA", ruta)
		}
		privada = rsaLlave
	case "RSA PRIVATE KEY":
		privada, err = x509.ParsePKCS1PrivateKey(bloque.Bytes)
		if err != nil {
			return nil, fmt.Errorf("la llave PKCS#1 en %s no es válida", ruta)
		}
	default:
		return nil, fmt.Errorf("tipo de bloque PEM no soportado en %s: %q", ruta, bloque.Type)
	}
	return NuevaLlave(privada)
}

// Kid devuelve el identificador de la llave que va en el header de cada token y en el JWKS.
func (l *Llave) Kid() string {
	return l.kid
}

// huella calcula el thumbprint RFC 7638 de la llave pública, en base64url.
// Así el kid es estable para una misma llave y cambia al rotarla.
func huella(publica *rsa.PublicKey) string {
	miembros := fmt.Sprintf(`{"e":"%s","kty":"RSA","n":"%s"}`, exponente(publica), modulo(publica))
	suma := sha256.Sum256([]byte(miembros))
	return base64.RawURLEncoding.EncodeToString(suma[:])
}

func modulo(publica *rsa.PublicKey) string {
	return base64.RawURLEncoding.EncodeToString(publica.N.Bytes())
}

func exponente(publica *rsa.PublicKey) string {
	return base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publica.E)).Bytes())
}
