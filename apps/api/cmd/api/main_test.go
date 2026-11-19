package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestImprimirContrato(t *testing.T) {
	var salida bytes.Buffer
	if err := ejecutar(context.Background(), []string{"openapi"}, &salida); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	contrato := salida.String()
	for _, esperado := range []string{"openapi: 3.1", "/healthz", "/readyz"} {
		if !strings.Contains(contrato, esperado) {
			t.Fatalf("el contrato no contiene %q", esperado)
		}
	}
}

func TestEscribirContratoEnArchivo(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "openapi.yaml")
	if err := ejecutar(context.Background(), []string{"openapi", ruta}, &bytes.Buffer{}); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	contenido, err := os.ReadFile(ruta)
	if err != nil {
		t.Fatalf("no se pudo leer el contrato: %v", err)
	}
	if !strings.Contains(string(contenido), "openapi: 3.1") {
		t.Fatal("el archivo no contiene un contrato OpenAPI 3.1")
	}
}

func TestLevantarServidorFallaSinConfiguracion(t *testing.T) {
	for _, variable := range []string{"API_PORT", "API_ENV", "DATABASE_URL", "FRONTEND_URL", "TZ"} {
		t.Setenv(variable, "")
	}
	if err := ejecutar(context.Background(), nil, &bytes.Buffer{}); err == nil {
		t.Fatal("se esperaba error por configuración faltante")
	}
}

func TestLevantarServidorFallaConDatabaseURLInvalida(t *testing.T) {
	t.Setenv("API_PORT", "0")
	t.Setenv("API_ENV", "test")
	t.Setenv("DATABASE_URL", "esto-no-es-una-url://")
	t.Setenv("FRONTEND_URL", "http://localhost:5160")
	t.Setenv("TZ", "America/Santiago")

	if err := ejecutar(context.Background(), nil, &bytes.Buffer{}); err == nil {
		t.Fatal("se esperaba error por DATABASE_URL inválida")
	}
}

func TestLevantarServidorSeCierraAlCancelar(t *testing.T) {
	t.Setenv("API_PORT", "0") // puerto 0: el sistema elige uno libre
	t.Setenv("API_ENV", "test")
	t.Setenv("DATABASE_URL", "postgres://usuario:clave@localhost:1/base")
	t.Setenv("FRONTEND_URL", "http://localhost:5160")
	t.Setenv("TZ", "America/Santiago")

	ctx, cancelar := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancelar()
	if err := ejecutar(ctx, nil, &bytes.Buffer{}); err != nil {
		t.Fatalf("el servidor debía cerrarse sin error, llegó: %v", err)
	}
}
