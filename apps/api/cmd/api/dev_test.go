package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/PINGESO-CIVIL/usuarios-roles/apps/api/internal/api"
	"github.com/PINGESO-CIVIL/usuarios-roles/apps/api/internal/config"
	"github.com/PINGESO-CIVIL/usuarios-roles/apps/api/internal/service/tokens"
)

// TestTokensDePruebaSoloEnDesarrollo arma el servidor igual que main para cada API_ENV.
func TestTokensDePruebaSoloEnDesarrollo(t *testing.T) {
	llave, err := tokens.GenerarLlave()
	if err != nil {
		t.Fatal(err)
	}
	emisor := tokens.NuevoEmisor(llave, "pingeso-identidad", "pingeso", 15*time.Minute)
	silencioso := slog.New(slog.NewTextHandler(io.Discard, nil))

	esperado := map[string]int{"dev": http.StatusOK, "test": http.StatusNotFound, "prod": http.StatusNotFound}
	for ambiente, estado := range esperado {
		t.Run(ambiente, func(t *testing.T) {
			dep, err := armarDependencias(context.Background(), config.Configuracion{Ambiente: ambiente}, emisor, silencioso)
			if err != nil {
				t.Fatal(err)
			}
			manejador, _ := api.NuevoServidor(dep)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/dev/tokens", strings.NewReader(`{"roles":["administrador"]}`))
			req.Header.Set("Content-Type", "application/json")
			manejador.ServeHTTP(rec, req)
			if rec.Code != estado {
				t.Fatalf("con API_ENV=%s se esperaba %d, llegó %d", ambiente, estado, rec.Code)
			}
		})
	}
}
