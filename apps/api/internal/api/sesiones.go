package api

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

// --- Structs del contrato ---

// SolicitudSesion son las credenciales del login.
type SolicitudSesion struct {
	Correo     string `json:"correo" format:"email" maxLength:"254" example:"admin@dev.local"`
	Contrasena string `json:"contrasena" minLength:"1" maxLength:"72" example:"admin-dev-local"`
}

type entradaIniciarSesion struct {
	Body SolicitudSesion
}

type salidaSesion struct {
	SetCookie http.Cookie `header:"Set-Cookie" doc:"Cookie sesion con el JWT: HttpOnly, Secure, SameSite=Lax, Path=/"`
	Body      UsuarioActual
}

type salidaUsuarioActual struct {
	Body UsuarioActual
}

type salidaCerrarSesion struct {
	SetCookie http.Cookie `header:"Set-Cookie" doc:"Borra la cookie sesion (Max-Age=0)"`
}

// Renovacion informa cuándo vence el token renovado.
type Renovacion struct {
	ExpiraEn time.Time `json:"expira_en" doc:"Momento en que vence el token nuevo"`
}

type salidaRenovacion struct {
	SetCookie http.Cookie `header:"Set-Cookie" doc:"Cookie sesion con el JWT renovado"`
	Body      Renovacion
}

// SolicitudCambioContrasena pide cambiar la contraseña propia.
type SolicitudCambioContrasena struct {
	ContrasenaActual string `json:"contrasena_actual" minLength:"1" maxLength:"72" example:"admin-dev-local"`
	ContrasenaNueva  string `json:"contrasena_nueva" minLength:"8" maxLength:"72" example:"una-clave-nueva-segura"`
}

type entradaCambioContrasena struct {
	Body SolicitudCambioContrasena
}

// JWK es una llave pública RSA (RFC 7517).
type JWK struct {
	Kty string `json:"kty" enum:"RSA"`
	Use string `json:"use" enum:"sig"`
	Alg string `json:"alg" enum:"RS256"`
	Kid string `json:"kid" doc:"Identificador de la llave: coincide con el header kid de cada token"`
	N   string `json:"n" doc:"Módulo RSA en base64url"`
	E   string `json:"e" example:"AQAB" doc:"Exponente RSA en base64url"`
}

// JWKS es el conjunto de llaves públicas (RFC 7517).
type JWKS struct {
	Keys []JWK `json:"keys"`
}

type salidaJWKS struct {
	CacheControl string `header:"Cache-Control"`
	Body         JWKS
}

// --- Rutas ---

func registrarSesiones(api huma.API, s ServicioSesiones) {
	huma.Register(api, publica(huma.Operation{
		OperationID: "iniciarSesion",
		Method:      http.MethodPost,
		Path:        "/api/sesiones",
		Summary:     "Inicia sesión con correo y contraseña",
		Description: "Fija la cookie `sesion` con el JWT y devuelve el usuario. El token no va en el body. " +
			"Un correo inexistente y una contraseña incorrecta dan el mismo 401. El 403 (cuenta no activa) " +
			"solo aparece con la contraseña correcta.",
		Tags:   []string{etiquetaSesiones},
		Errors: []int{http.StatusUnauthorized, http.StatusForbidden},
	}), func(ctx context.Context, in *entradaIniciarSesion) (*salidaSesion, error) {
		sesion, err := s.Iniciar(ctx, in.Body.Correo, in.Body.Contrasena)
		if err != nil {
			return nil, traducirError(err)
		}
		return &salidaSesion{
			SetCookie: cookieSesion(sesion.Token, segundosHasta(sesion.Usuario.ExpiraEn)),
			Body:      aUsuarioActual(sesion.Usuario),
		}, nil
	})

	huma.Register(api, autenticada(huma.Operation{
		OperationID: "obtenerSesionActual",
		Method:      http.MethodGet,
		Path:        "/api/sesiones/actual",
		Summary:     "Devuelve el usuario de la sesión actual, con sus permisos",
		Tags:        []string{etiquetaSesiones},
	}), func(ctx context.Context, _ *struct{}) (*salidaUsuarioActual, error) {
		actual, err := s.Actual(ctx)
		if err != nil {
			return nil, traducirError(err)
		}
		return &salidaUsuarioActual{Body: aUsuarioActual(actual)}, nil
	})

	huma.Register(api, autenticada(huma.Operation{
		OperationID:   "cerrarSesion",
		Method:        http.MethodDelete,
		Path:          "/api/sesiones/actual",
		Summary:       "Cierra la sesión borrando la cookie",
		Description:   "El JWT ya emitido sigue siendo válido hasta su `exp` (15 minutos como máximo).",
		Tags:          []string{etiquetaSesiones},
		DefaultStatus: http.StatusNoContent,
	}), func(context.Context, *struct{}) (*salidaCerrarSesion, error) {
		return &salidaCerrarSesion{SetCookie: cookieSesion("", -1)}, nil
	})

	huma.Register(api, autenticada(huma.Operation{
		OperationID: "renovarSesion",
		Method:      http.MethodPost,
		Path:        "/api/sesiones/actual/renovacion",
		Summary:     "Renueva el token de la sesión (simulado)",
		Description: "Simulado en el Hito 1: emite un token nuevo con los mismos roles y permisos.",
		Tags:        []string{etiquetaSesiones},
	}), func(ctx context.Context, _ *struct{}) (*salidaRenovacion, error) {
		renovado, err := s.Renovar(ctx)
		if err != nil {
			return nil, traducirError(err)
		}
		salida := &salidaRenovacion{SetCookie: cookieSesion(renovado.Token, segundosHasta(renovado.ExpiraEn))}
		salida.Body.ExpiraEn = renovado.ExpiraEn
		return salida, nil
	})

	huma.Register(api, autenticada(huma.Operation{
		OperationID:   "cambiarContrasena",
		Method:        http.MethodPut,
		Path:          "/api/sesiones/actual/contrasena",
		Summary:       "Cambia la contraseña del usuario de la sesión (simulado)",
		Description:   "Simulado en el Hito 1: valida el formato y no cambia la contraseña.",
		Tags:          []string{etiquetaSesiones},
		DefaultStatus: http.StatusNoContent,
	}), func(ctx context.Context, in *entradaCambioContrasena) (*struct{}, error) {
		if err := s.CambiarContrasena(ctx, in.Body.ContrasenaActual, in.Body.ContrasenaNueva); err != nil {
			return nil, traducirError(err)
		}
		return nil, nil
	})

	huma.Register(api, publica(huma.Operation{
		OperationID: "obtenerJWKS",
		Method:      http.MethodGet,
		Path:        "/.well-known/jwks.json",
		Summary:     "Publica las llaves públicas para validar los JWT",
		Description: "Es la URL que los demás servicios configuran en `AUTH_JWKS_URL`.",
		Tags:        []string{etiquetaSesiones},
	}), func(context.Context, *struct{}) (*salidaJWKS, error) {
		salida := &salidaJWKS{CacheControl: "public, max-age=300"}
		salida.Body.Keys = []JWK{}
		for _, l := range s.JWKS() {
			salida.Body.Keys = append(salida.Body.Keys, JWK{Kty: l.Kty, Use: l.Use, Alg: l.Alg, Kid: l.Kid, N: l.N, E: l.E})
		}
		return salida, nil
	})
}

// segundosHasta calcula el Max-Age de la cookie para que venza junto con el token.
func segundosHasta(t time.Time) int {
	return max(int(time.Until(t).Seconds()), 1)
}
