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

// variablesDelServicio son todas las variables que lee config.Cargar.
var variablesDelServicio = []string{
	"API_PORT", "API_ENV", "DATABASE_URL", "FRONTEND_URL", "TZ",
	"JWT_EMISOR", "JWT_AUDIENCIA", "JWT_TTL", "JWT_LLAVE_PRIVADA_ARCHIVO",
}

// definirEntornoValido deja una configuración completa de pruebas, con API_ENV=test.
func definirEntornoValido(t *testing.T) {
	t.Helper()
	t.Setenv("API_PORT", "0") // puerto 0: el sistema elige uno libre
	t.Setenv("API_ENV", "test")
	t.Setenv("DATABASE_URL", "postgres://usuario:clave@localhost:1/base")
	t.Setenv("FRONTEND_URL", "http://localhost:5160")
	t.Setenv("TZ", "America/Santiago")
	t.Setenv("JWT_EMISOR", "pingeso-identidad")
	t.Setenv("JWT_AUDIENCIA", "pingeso")
	t.Setenv("JWT_TTL", "15m")
	t.Setenv("JWT_LLAVE_PRIVADA_ARCHIVO", escribirLlaveDePrueba(t))
}

func TestLevantarServidorFallaSinConfiguracion(t *testing.T) {
	for _, variable := range variablesDelServicio {
		t.Setenv(variable, "")
	}
	if err := ejecutar(context.Background(), nil, &bytes.Buffer{}); err == nil {
		t.Fatal("se esperaba error por configuración faltante")
	}
}

func TestLevantarServidorFallaConDatabaseURLInvalida(t *testing.T) {
	definirEntornoValido(t)
	t.Setenv("DATABASE_URL", "esto-no-es-una-url://")

	if err := ejecutar(context.Background(), nil, &bytes.Buffer{}); err == nil {
		t.Fatal("se esperaba error por DATABASE_URL inválida")
	}
}

func TestLevantarServidorFallaConLlaveInvalida(t *testing.T) {
	definirEntornoValido(t)
	ruta := filepath.Join(t.TempDir(), "rota.pem")
	if err := os.WriteFile(ruta, []byte("no es una llave"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("JWT_LLAVE_PRIVADA_ARCHIVO", ruta)

	if err := ejecutar(context.Background(), nil, &bytes.Buffer{}); err == nil {
		t.Fatal("se esperaba error por llave inválida")
	}
}

func TestLevantarServidorFallaSinLlaveEnProduccion(t *testing.T) {
	definirEntornoValido(t)
	t.Setenv("API_ENV", "prod")
	t.Setenv("JWT_LLAVE_PRIVADA_ARCHIVO", "")

	err := ejecutar(context.Background(), nil, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "JWT_LLAVE_PRIVADA_ARCHIVO") {
		t.Fatalf("en prod la llave es obligatoria, llegó: %v", err)
	}
}

func TestLevantarServidorEnDesarrolloGeneraLlave(t *testing.T) {
	definirEntornoValido(t)
	t.Setenv("API_ENV", "dev")
	t.Setenv("JWT_LLAVE_PRIVADA_ARCHIVO", "")

	ctx, cancelar := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancelar()
	if err := ejecutar(ctx, nil, &bytes.Buffer{}); err != nil {
		t.Fatalf("en dev debe arrancar sin llave: %v", err)
	}
}

func TestLevantarServidorSeCierraAlCancelar(t *testing.T) {
	definirEntornoValido(t)

	ctx, cancelar := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancelar()
	if err := ejecutar(ctx, nil, &bytes.Buffer{}); err != nil {
		t.Fatalf("el servidor debía cerrarse sin error, llegó: %v", err)
	}
}
