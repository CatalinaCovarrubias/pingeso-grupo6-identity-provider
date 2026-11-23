package config

import (
	"strings"
	"testing"
	"time"
)

// definirEntorno deja todas las variables obligatorias con valores válidos.
func definirEntorno(t *testing.T) {
	t.Helper()
	t.Setenv("API_PORT", "8080")
	t.Setenv("API_ENV", "dev")
	t.Setenv("DATABASE_URL", "postgres://usuario:clave@localhost:5432/base")
	t.Setenv("FRONTEND_URL", "http://localhost:5160")
	t.Setenv("TZ", "America/Santiago")
	t.Setenv("JWT_EMISOR", "pingeso-identidad")
	t.Setenv("JWT_AUDIENCIA", "pingeso")
	t.Setenv("JWT_TTL", "15m")
	t.Setenv("JWT_LLAVE_PRIVADA_ARCHIVO", "")
}

func TestCargarConTodasLasVariables(t *testing.T) {
	definirEntorno(t)

	cfg, err := Cargar()
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if cfg.Puerto != "8080" || cfg.Ambiente != "dev" {
		t.Fatalf("configuración inesperada: %+v", cfg)
	}
	if cfg.JWTTTL != 15*time.Minute || cfg.JWTEmisor != "pingeso-identidad" || cfg.JWTAudiencia != "pingeso" {
		t.Fatalf("configuración JWT inesperada: %+v", cfg)
	}
	if cfg.EsProduccion() || !cfg.EsDesarrollo() {
		t.Fatal("dev debe considerarse desarrollo y no producción")
	}
}

func TestCargarFallaSiFaltanVariables(t *testing.T) {
	definirEntorno(t)
	t.Setenv("DATABASE_URL", "")
	t.Setenv("TZ", "")
	t.Setenv("JWT_EMISOR", "")

	_, err := Cargar()
	if err == nil {
		t.Fatal("se esperaba error por variables faltantes")
	}
	for _, nombre := range []string{"DATABASE_URL", "TZ", "JWT_EMISOR"} {
		if !strings.Contains(err.Error(), nombre) {
			t.Fatalf("el error debe nombrar %s: %v", nombre, err)
		}
	}
}

func TestCargarFallaConAmbienteInvalido(t *testing.T) {
	definirEntorno(t)
	t.Setenv("API_ENV", "produccion")
	t.Setenv("JWT_LLAVE_PRIVADA_ARCHIVO", "/tmp/llave.pem")

	if _, err := Cargar(); err == nil {
		t.Fatal("se esperaba error por API_ENV inválido")
	}
}

func TestEsProduccion(t *testing.T) {
	definirEntorno(t)
	t.Setenv("API_ENV", "prod")
	t.Setenv("JWT_LLAVE_PRIVADA_ARCHIVO", "/run/secrets/jwt.pem")

	cfg, err := Cargar()
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if !cfg.EsProduccion() || cfg.EsDesarrollo() {
		t.Fatal("prod debe considerarse producción")
	}
}

func TestLlaveObligatoriaFueraDeDesarrollo(t *testing.T) {
	for _, ambiente := range []string{"test", "prod"} {
		t.Run(ambiente, func(t *testing.T) {
			definirEntorno(t)
			t.Setenv("API_ENV", ambiente)

			_, err := Cargar()
			if err == nil || !strings.Contains(err.Error(), "JWT_LLAVE_PRIVADA_ARCHIVO") {
				t.Fatalf("en %s la llave es obligatoria, llegó: %v", ambiente, err)
			}
		})
	}
}

func TestLlaveOpcionalEnDesarrollo(t *testing.T) {
	definirEntorno(t)

	cfg, err := Cargar()
	if err != nil {
		t.Fatalf("en dev la llave es opcional: %v", err)
	}
	if cfg.JWTLlavePrivadaArchivo != "" {
		t.Fatal("no se esperaba ruta de llave")
	}
}

func TestCargarFallaConTTLInvalido(t *testing.T) {
	for _, valor := range []string{"quince", "30s", "48h"} {
		t.Run(valor, func(t *testing.T) {
			definirEntorno(t)
			t.Setenv("JWT_TTL", valor)

			if _, err := Cargar(); err == nil || !strings.Contains(err.Error(), "JWT_TTL") {
				t.Fatalf("se esperaba error por JWT_TTL=%s, llegó: %v", valor, err)
			}
		})
	}
}
