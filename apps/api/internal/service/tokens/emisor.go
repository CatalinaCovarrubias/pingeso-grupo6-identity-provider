package tokens

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Algoritmo es el único algoritmo de firma que emite y acepta el servicio.
const Algoritmo = "RS256"

// Desfase es la tolerancia de reloj que se acepta al validar exp, nbf e iat.
const Desfase = 30 * time.Second

// ErrTokenInvalido envuelve cualquier motivo de rechazo de un token.
var ErrTokenInvalido = errors.New("token inválido")

// Claims es el contenido de cada token emitido.
type Claims struct {
	jwt.RegisteredClaims
	Roles    []string `json:"roles"`
	Permisos []string `json:"permisos"`
}

// Emisor firma y valida tokens con una sola llave.
type Emisor struct {
	llave     *Llave
	emisor    string
	audiencia string
	ttl       time.Duration
	ahora     func() time.Time
}

// NuevoEmisor arma un emisor con el iss, el aud y la duración indicados.
func NuevoEmisor(llave *Llave, emisor, audiencia string, ttl time.Duration) *Emisor {
	return &Emisor{llave: llave, emisor: emisor, audiencia: audiencia, ttl: ttl, ahora: time.Now}
}

// ConReloj reemplaza la hora actual; solo se usa en pruebas.
func (e *Emisor) ConReloj(ahora func() time.Time) *Emisor {
	copia := *e
	copia.ahora = ahora
	return &copia
}

// TTL es la duración de cada token.
func (e *Emisor) TTL() time.Duration {
	return e.ttl
}

// Emitir firma un token para el sujeto con los roles y permisos indicados.
func (e *Emisor) Emitir(sujeto string, roles, permisos []string) (string, Claims, error) {
	ahora := e.ahora()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    e.emisor,
			Subject:   sujeto,
			Audience:  jwt.ClaimStrings{e.audiencia},
			ExpiresAt: jwt.NewNumericDate(ahora.Add(e.ttl)),
			IssuedAt:  jwt.NewNumericDate(ahora),
			NotBefore: jwt.NewNumericDate(ahora),
			ID:        uuid.NewString(),
		},
		Roles:    noNulo(roles),
		Permisos: noNulo(permisos),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = e.llave.kid
	firmado, err := token.SignedString(e.llave.privada)
	if err != nil {
		return "", Claims{}, fmt.Errorf("no se pudo firmar el token: %w", err)
	}
	return firmado, claims, nil
}

// Validar comprueba firma, algoritmo, kid, iss, aud, exp, nbf e iat, con un desfase
// de reloj de 30 segundos. Rechaza alg "none", HS256 y cualquier otro algoritmo.
func (e *Emisor) Validar(token string) (Claims, error) {
	var claims Claims
	_, err := jwt.ParseWithClaims(token, &claims, e.llavePublica,
		jwt.WithValidMethods([]string{Algoritmo}),
		jwt.WithIssuer(e.emisor),
		jwt.WithAudience(e.audiencia),
		jwt.WithLeeway(Desfase),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithTimeFunc(e.ahora),
	)
	if err != nil {
		return Claims{}, fmt.Errorf("%w: %w", ErrTokenInvalido, err)
	}
	if claims.Subject == "" {
		return Claims{}, fmt.Errorf("%w: falta sub", ErrTokenInvalido)
	}
	return claims, nil
}

// llavePublica entrega la llave para verificar la firma solo si el kid coincide.
func (e *Emisor) llavePublica(token *jwt.Token) (any, error) {
	kid, _ := token.Header["kid"].(string)
	if kid != e.llave.kid {
		return nil, errors.New("kid desconocido")
	}
	return &e.llave.privada.PublicKey, nil
}

// JWK es una llave pública en formato JSON Web Key (RFC 7517).
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS devuelve las llaves públicas vigentes. Hoy es una; al rotar habrá dos.
func (e *Emisor) JWKS() []JWK {
	publica := &e.llave.privada.PublicKey
	return []JWK{{
		Kty: "RSA",
		Use: "sig",
		Alg: Algoritmo,
		Kid: e.llave.kid,
		N:   modulo(publica),
		E:   exponente(publica),
	}}
}

func noNulo(valores []string) []string {
	if valores == nil {
		return []string{}
	}
	return valores
}
