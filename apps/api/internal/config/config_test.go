package config

import (
	"strings"
	"testing"
)

// definirEntorno deja todas las variables obligatorias con valores válidos.
func definirEntorno(t *testing.T) {
	t.Setenv("API_PORT", "8080")
	t.Setenv("API_ENV", "dev")
	t.Setenv("DATABASE_URL", "postgres://usuario:clave@localhost:5432/base")
	t.Setenv("FRONTEND_URL", "http://localhost:5160")
	t.Setenv("TZ", "America/Santiago")
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
	if cfg.EsProduccion() {
		t.Fatal("dev no debe considerarse producción")
	}
}

func TestCargarFallaSiFaltanVariables(t *testing.T) {
	definirEntorno(t)
	t.Setenv("DATABASE_URL", "")
	t.Setenv("TZ", "")

	_, err := Cargar()
	if err == nil {
		t.Fatal("se esperaba error por variables faltantes")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") || !strings.Contains(err.Error(), "TZ") {
		t.Fatalf("el error debe nombrar las variables faltantes: %v", err)
	}
}

func TestCargarFallaConAmbienteInvalido(t *testing.T) {
	definirEntorno(t)
	t.Setenv("API_ENV", "produccion")

	if _, err := Cargar(); err == nil {
		t.Fatal("se esperaba error por API_ENV inválido")
	}
}

func TestEsProduccion(t *testing.T) {
	definirEntorno(t)
	t.Setenv("API_ENV", "prod")

	cfg, err := Cargar()
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if !cfg.EsProduccion() {
		t.Fatal("prod debe considerarse producción")
	}
}
