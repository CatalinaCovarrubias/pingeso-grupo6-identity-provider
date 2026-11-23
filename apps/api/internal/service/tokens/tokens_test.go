package tokens

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	llaveCompartida *Llave
	unaVez          sync.Once
)

// llaveDePrueba genera una sola llave para todo el paquete: crear llaves RSA es lento.
func llaveDePrueba(t *testing.T) *Llave {
	t.Helper()
	unaVez.Do(func() {
		var err error
		llaveCompartida, err = GenerarLlave()
		if err != nil {
			t.Fatalf("no se pudo generar la llave: %v", err)
		}
	})
	return llaveCompartida
}

var instante = time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)

func emisorDePrueba(t *testing.T) *Emisor {
	t.Helper()
	return NuevoEmisor(llaveDePrueba(t), "pingeso-identidad", "pingeso", 15*time.Minute).
		ConReloj(func() time.Time { return instante })
}

func TestEmitirYValidar(t *testing.T) {
	e := emisorDePrueba(t)
	token, emitidos, err := e.Emitir("0b6f1c2e-6c1a-4f0e-9a51-3d2b7e8f9a01", []string{"administrador"}, []string{"usuarios:administrar"})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	claims, err := e.Validar(token)
	if err != nil {
		t.Fatalf("el token propio debe validar: %v", err)
	}
	if claims.Subject != emitidos.Subject || claims.ID == "" || claims.Roles[0] != "administrador" {
		t.Fatalf("claims inesperados: %+v", claims)
	}
	if !claims.ExpiresAt.Equal(instante.Add(15 * time.Minute)) {
		t.Fatalf("exp inesperado: %v", claims.ExpiresAt)
	}

	cabecera := decodificarSegmento(t, token, 0)
	if cabecera["alg"] != "RS256" || cabecera["kid"] != e.llave.Kid() {
		t.Fatalf("header inesperado: %v", cabecera)
	}
}

func TestEmitirSinRolesUsaListasVacias(t *testing.T) {
	e := emisorDePrueba(t)
	token, _, err := e.Emitir("sujeto", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	cuerpo := decodificarSegmento(t, token, 1)
	if _, ok := cuerpo["roles"].([]any); !ok {
		t.Fatalf("roles debe ser una lista, llegó %v", cuerpo["roles"])
	}
}

func TestRechazaTokensInvalidos(t *testing.T) {
	e := emisorDePrueba(t)
	valido, _, err := e.Emitir("sujeto", []string{"usuario"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	claimsBase := Claims{RegisteredClaims: jwt.RegisteredClaims{
		Issuer: "pingeso-identidad", Subject: "sujeto", Audience: jwt.ClaimStrings{"pingeso"},
		ExpiresAt: jwt.NewNumericDate(instante.Add(time.Hour)), IssuedAt: jwt.NewNumericDate(instante),
	}}

	sinFirma := strings.Join(strings.Split(valido, ".")[:2], ".") + "."

	ninguno := jwt.NewWithClaims(jwt.SigningMethodNone, claimsBase)
	ninguno.Header["kid"] = e.llave.Kid()
	tokenNone, err := ninguno.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}

	// Confusión de algoritmo: HS256 usando la llave pública como secreto compartido.
	publicaDER, _ := x509.MarshalPKIXPublicKey(&e.llave.privada.PublicKey)
	hs := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsBase)
	hs.Header["kid"] = e.llave.Kid()
	tokenHS, err := hs.SignedString(publicaDER)
	if err != nil {
		t.Fatal(err)
	}

	partes := strings.Split(valido, ".")
	cuerpo := decodificarSegmento(t, valido, 1)
	cuerpo["roles"] = []string{"administrador"}
	alterado, _ := json.Marshal(cuerpo)
	tokenAlterado := partes[0] + "." + base64.RawURLEncoding.EncodeToString(alterado) + "." + partes[2]

	otraLlave, err := GenerarLlave()
	if err != nil {
		t.Fatal(err)
	}
	ajeno, _, _ := NuevoEmisor(otraLlave, "pingeso-identidad", "pingeso", 15*time.Minute).
		ConReloj(func() time.Time { return instante }).Emitir("sujeto", nil, nil)

	// Otra llave que se presenta con el kid propio: la firma no coincide.
	suplantado := jwt.NewWithClaims(jwt.SigningMethodRS256, claimsBase)
	suplantado.Header["kid"] = e.llave.Kid()
	tokenSuplantado, _ := suplantado.SignedString(otraLlave.privada)

	otraAudiencia, _, _ := NuevoEmisor(e.llave, "pingeso-identidad", "otro-curso", 15*time.Minute).
		ConReloj(func() time.Time { return instante }).Emitir("sujeto", nil, nil)
	otroEmisor, _, _ := NuevoEmisor(e.llave, "impostor", "pingeso", 15*time.Minute).
		ConReloj(func() time.Time { return instante }).Emitir("sujeto", nil, nil)
	vencido, _, _ := NuevoEmisor(e.llave, "pingeso-identidad", "pingeso", 15*time.Minute).
		ConReloj(func() time.Time { return instante.Add(-16 * time.Minute) }).Emitir("sujeto", nil, nil)
	futuro, _, _ := NuevoEmisor(e.llave, "pingeso-identidad", "pingeso", 15*time.Minute).
		ConReloj(func() time.Time { return instante.Add(time.Minute) }).Emitir("sujeto", nil, nil)
	sinSujeto, _, _ := e.Emitir("", nil, nil)

	casos := map[string]string{
		"vacío":                 "",
		"basura":                "esto.no.es-un-jwt",
		"sin firma":             sinFirma,
		"alg none":              tokenNone,
		"HS256 con llave":       tokenHS,
		"payload alterado":      tokenAlterado,
		"otra llave y otro kid": ajeno,
		"otra llave, mismo kid": tokenSuplantado,
		"otra audiencia":        otraAudiencia,
		"otro emisor":           otroEmisor,
		"vencido":               vencido,
		"emitido en el futuro":  futuro,
		"sin sub":               sinSujeto,
	}
	for nombre, token := range casos {
		t.Run(nombre, func(t *testing.T) {
			if _, err := e.Validar(token); !errors.Is(err, ErrTokenInvalido) {
				t.Fatalf("se esperaba ErrTokenInvalido, llegó %v", err)
			}
		})
	}
}

func TestAceptaDesfaseDeReloj(t *testing.T) {
	e := emisorDePrueba(t)
	// Vencido hace 20 s: dentro del desfase de 30 s.
	token, _, _ := NuevoEmisor(e.llave, "pingeso-identidad", "pingeso", 15*time.Minute).
		ConReloj(func() time.Time { return instante.Add(-15*time.Minute - 20*time.Second) }).Emitir("sujeto", nil, nil)
	if _, err := e.Validar(token); err != nil {
		t.Fatalf("un desfase de 20 s debe aceptarse: %v", err)
	}
	// Emitido 20 s en el futuro (reloj del emisor adelantado).
	adelantado, _, _ := NuevoEmisor(e.llave, "pingeso-identidad", "pingeso", 15*time.Minute).
		ConReloj(func() time.Time { return instante.Add(20 * time.Second) }).Emitir("sujeto", nil, nil)
	if _, err := e.Validar(adelantado); err != nil {
		t.Fatalf("un reloj adelantado 20 s debe aceptarse: %v", err)
	}
}

func TestJWKSPublicaElKidDeFirma(t *testing.T) {
	e := emisorDePrueba(t)
	llaves := e.JWKS()
	if len(llaves) != 1 {
		t.Fatalf("se esperaba una llave, llegaron %d", len(llaves))
	}
	jwk := llaves[0]
	if jwk.Kid != e.llave.Kid() || jwk.Alg != "RS256" || jwk.Kty != "RSA" || jwk.Use != "sig" || jwk.E != "AQAB" {
		t.Fatalf("JWK inesperado: %+v", jwk)
	}
	if strings.Contains(jwk.N, "=") {
		t.Fatal("n debe ir en base64url sin relleno")
	}
}

func TestKidEsEstableParaLaMismaLlave(t *testing.T) {
	l := llaveDePrueba(t)
	otra, err := NuevaLlave(l.privada)
	if err != nil {
		t.Fatal(err)
	}
	if otra.Kid() != l.Kid() || len(l.Kid()) != 43 {
		t.Fatalf("kid inestable o con largo inesperado: %q", l.Kid())
	}
}

func TestCargarLlave(t *testing.T) {
	l := llaveDePrueba(t)
	pkcs8, _ := x509.MarshalPKCS8PrivateKey(l.privada)
	pkcs1 := x509.MarshalPKCS1PrivateKey(l.privada)

	for nombre, bloque := range map[string]*pem.Block{
		"PKCS#8": {Type: "PRIVATE KEY", Bytes: pkcs8},
		"PKCS#1": {Type: "RSA PRIVATE KEY", Bytes: pkcs1},
	} {
		t.Run(nombre, func(t *testing.T) {
			cargada, err := CargarLlave(escribir(t, pem.EncodeToMemory(bloque)))
			if err != nil {
				t.Fatalf("no se esperaba error: %v", err)
			}
			if cargada.Kid() != l.Kid() {
				t.Fatal("el kid debe coincidir con el de la llave original")
			}
		})
	}
}

func TestCargarLlaveFalla(t *testing.T) {
	chica, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	casos := map[string]string{
		"no existe":     filepath.Join(t.TempDir(), "no-existe.pem"),
		"sin PEM":       escribir(t, []byte("no es un pem")),
		"tipo extraño":  escribir(t, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte{1}})),
		"PKCS#8 roto":   escribir(t, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte{1, 2}})),
		"PKCS#1 roto":   escribir(t, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: []byte{1, 2}})),
		"menos de 2048": escribir(t, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(chica)})),
	}
	for nombre, ruta := range casos {
		t.Run(nombre, func(t *testing.T) {
			if _, err := CargarLlave(ruta); err == nil {
				t.Fatal("se esperaba error")
			}
		})
	}
	if _, err := NuevaLlave(nil); err == nil {
		t.Fatal("una llave nula debe fallar")
	}
}

func escribir(t *testing.T, contenido []byte) string {
	t.Helper()
	ruta := filepath.Join(t.TempDir(), "llave.pem")
	if err := os.WriteFile(ruta, contenido, 0o600); err != nil {
		t.Fatal(err)
	}
	return ruta
}

func decodificarSegmento(t *testing.T, token string, indice int) map[string]any {
	t.Helper()
	crudo, err := base64.RawURLEncoding.DecodeString(strings.Split(token, ".")[indice])
	if err != nil {
		t.Fatal(err)
	}
	var datos map[string]any
	if err := json.Unmarshal(crudo, &datos); err != nil {
		t.Fatal(err)
	}
	return datos
}
