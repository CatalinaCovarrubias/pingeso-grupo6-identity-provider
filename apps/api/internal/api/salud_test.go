package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// bdFalsa simula PostgreSQL: responde bien o con error según el campo err.
type bdFalsa struct{ err error }

func (b bdFalsa) Ping(context.Context) error { return b.err }

// pedir hace una petición GET al servidor y devuelve la respuesta grabada.
func pedir(t *testing.T, dep Dependencias, ruta string) *httptest.ResponseRecorder {
	t.Helper()
	servidor, _ := NuevoServidor(dep)
	rec := httptest.NewRecorder()
	servidor.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, ruta, nil))
	return rec
}

func TestHealthzRespondeOK(t *testing.T) {
	rec := pedir(t, Dependencias{}, "/healthz")
	if rec.Code != http.StatusOK {
		t.Fatalf("se esperaba 200, llegó %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"estado":"ok"`) {
		t.Fatalf("cuerpo inesperado: %s", rec.Body.String())
	}
}

func TestReadyzConBaseDisponible(t *testing.T) {
	rec := pedir(t, Dependencias{BD: bdFalsa{}}, "/readyz")
	if rec.Code != http.StatusOK {
		t.Fatalf("se esperaba 200, llegó %d", rec.Code)
	}
}

func TestReadyzConBaseCaida(t *testing.T) {
	rec := pedir(t, Dependencias{BD: bdFalsa{err: errors.New("sin conexión")}}, "/readyz")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("se esperaba 503, llegó %d", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "application/problem+json") {
		t.Fatalf("el error debe ser problem+json, llegó %q", rec.Header().Get("Content-Type"))
	}
}

func TestReadyzSinBaseConfigurada(t *testing.T) {
	rec := pedir(t, Dependencias{}, "/readyz")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("se esperaba 503, llegó %d", rec.Code)
	}
}

func TestDocsSoloEnDesarrollo(t *testing.T) {
	if rec := pedir(t, Dependencias{PublicarDocs: true}, "/api/docs"); rec.Code != http.StatusOK {
		t.Fatalf("en dev /api/docs debe responder 200, llegó %d", rec.Code)
	}
	if rec := pedir(t, Dependencias{PublicarDocs: false}, "/api/docs"); rec.Code != http.StatusNotFound {
		t.Fatalf("en prod /api/docs debe responder 404, llegó %d", rec.Code)
	}
}
