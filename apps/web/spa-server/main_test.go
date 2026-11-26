package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func raizDePrueba(t *testing.T) string {
	t.Helper()
	raiz := t.TempDir()
	escribir := func(nombre, contenido string) {
		ruta := filepath.Join(raiz, nombre)
		if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(ruta, []byte(contenido), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	escribir("index.html", "<!doctype html><title>app</title>")
	escribir("assets/app-abc123.js", "console.log(1)")
	escribir("mockServiceWorker.js", "// worker")
	// Un archivo fuera de la raíz que no se debe poder leer.
	if err := os.WriteFile(filepath.Join(filepath.Dir(raiz), "secreto.txt"), []byte("contenido-secreto"), 0o600); err != nil {
		t.Fatal(err)
	}
	return raiz
}

func silencioso() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func servidorDePrueba(t *testing.T, backendURL string) *httptest.Server {
	t.Helper()
	raiz, err := os.OpenRoot(raizDePrueba(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = raiz.Close() })
	manejador, err := nuevoManejador(raiz.FS(), backendURL, silencioso())
	if err != nil {
		t.Fatal(err)
	}
	servidor := httptest.NewServer(manejador)
	t.Cleanup(servidor.Close)
	return servidor
}

func pedir(t *testing.T, metodo, url string) (*http.Response, string) {
	t.Helper()
	req, err := http.NewRequest(metodo, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	cliente := http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := cliente.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	cuerpo, _ := io.ReadAll(resp.Body)
	return resp, string(cuerpo)
}

func TestSirveLaAppConCabecerasSeguras(t *testing.T) {
	srv := servidorDePrueba(t, "http://127.0.0.1:1")

	for _, ruta := range []string{"/", "/una/ruta/de/la/spa", "/index.html"} {
		resp, cuerpo := pedir(t, http.MethodGet, srv.URL+ruta)
		if resp.StatusCode != http.StatusOK || !strings.Contains(cuerpo, "<title>app</title>") {
			t.Fatalf("%s: estado %d, cuerpo %q", ruta, resp.StatusCode, cuerpo)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-cache" {
			t.Errorf("%s: Cache-Control = %q", ruta, got)
		}
		if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
			t.Errorf("%s: X-Content-Type-Options = %q", ruta, got)
		}
		if got := resp.Header.Get("Referrer-Policy"); got != "strict-origin-when-cross-origin" {
			t.Errorf("%s: Referrer-Policy = %q", ruta, got)
		}
		if got := resp.Header.Get("Content-Security-Policy"); got != "frame-ancestors 'none'" {
			t.Errorf("%s: Content-Security-Policy = %q", ruta, got)
		}
		if resp.Header.Get("X-Request-Id") == "" {
			t.Errorf("%s: falta X-Request-Id", ruta)
		}
	}
}

func TestAssetsInmutablesYSinFallback(t *testing.T) {
	srv := servidorDePrueba(t, "http://127.0.0.1:1")

	resp, cuerpo := pedir(t, http.MethodGet, srv.URL+"/assets/app-abc123.js")
	if resp.StatusCode != http.StatusOK || cuerpo != "console.log(1)" {
		t.Fatalf("estado %d, cuerpo %q", resp.StatusCode, cuerpo)
	}
	if got := resp.Header.Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Errorf("Cache-Control = %q", got)
	}
	if got := resp.Header.Get("Content-Type"); !strings.HasPrefix(got, "text/javascript") {
		t.Errorf("Content-Type = %q", got)
	}

	// Un asset que no existe es un 404, no el index.html (evita servir HTML como JS).
	resp, _ = pedir(t, http.MethodGet, srv.URL+"/assets/no-existe.js")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("asset inexistente: estado %d", resp.StatusCode)
	}
}

func TestSirveOtrosArchivosDeLaRaiz(t *testing.T) {
	srv := servidorDePrueba(t, "http://127.0.0.1:1")
	resp, cuerpo := pedir(t, http.MethodGet, srv.URL+"/mockServiceWorker.js")
	if resp.StatusCode != http.StatusOK || cuerpo != "// worker" {
		t.Fatalf("estado %d, cuerpo %q", resp.StatusCode, cuerpo)
	}
}

func TestNoSaleDeLaRaiz(t *testing.T) {
	srv := servidorDePrueba(t, "http://127.0.0.1:1")
	for _, ruta := range []string{"/../secreto.txt", "/assets/../../secreto.txt", "/%2e%2e/secreto.txt"} {
		_, cuerpo := pedir(t, http.MethodGet, srv.URL+ruta)
		if strings.Contains(cuerpo, "contenido-secreto") {
			t.Errorf("%s devolvió el archivo de fuera de la raíz", ruta)
		}
	}
}

func TestSoloGetYHeadEnEstaticos(t *testing.T) {
	srv := servidorDePrueba(t, "http://127.0.0.1:1")
	resp, _ := pedir(t, http.MethodPost, srv.URL+"/")
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("POST /: estado %d", resp.StatusCode)
	}
	resp, _ = pedir(t, http.MethodHead, srv.URL+"/")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("HEAD /: estado %d", resp.StatusCode)
	}
}

func TestHealthz(t *testing.T) {
	srv := servidorDePrueba(t, "http://127.0.0.1:1")
	resp, cuerpo := pedir(t, http.MethodGet, srv.URL+"/healthz")
	if resp.StatusCode != http.StatusOK || cuerpo != `{"estado":"ok"}` {
		t.Fatalf("estado %d, cuerpo %q", resp.StatusCode, cuerpo)
	}
	if code := salud(srv.URL+"/healthz", time.Second); code != 0 {
		t.Errorf("salud = %d con el servidor arriba", code)
	}
	if code := salud(srv.URL+"/no-existe-healthz/../assets/x", time.Second); code != 1 {
		t.Errorf("salud = %d con una respuesta distinta de 200", code)
	}
	if code := salud("http://127.0.0.1:1/healthz", 200*time.Millisecond); code != 1 {
		t.Errorf("salud = %d sin servidor", code)
	}
}

func TestReenviaApiYWellKnownAlBackend(t *testing.T) {
	var vistas, hosts []string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vistas = append(vistas, r.Method+" "+r.URL.RequestURI())
		hosts = append(hosts, r.Host)
		if r.Header.Get("X-Forwarded-Proto") == "" || r.Header.Get("X-Request-Id") == "" {
			t.Errorf("faltan cabeceras reenviadas: %v", r.Header)
		}
		http.SetCookie(w, &http.Cookie{Name: "sesion", Value: "jwt", HttpOnly: true, Secure: true, Path: "/"})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer backend.Close()
	srv := servidorDePrueba(t, backend.URL)

	resp, cuerpo := pedir(t, http.MethodPost, srv.URL+"/api/sesiones?x=1")
	if resp.StatusCode != http.StatusOK || cuerpo != `{"ok":true}` {
		t.Fatalf("estado %d, cuerpo %q", resp.StatusCode, cuerpo)
	}
	if !strings.Contains(resp.Header.Get("Set-Cookie"), "sesion=jwt") {
		t.Errorf("la cookie no llegó al navegador: %q", resp.Header.Get("Set-Cookie"))
	}
	if got := resp.Header.Values("X-Content-Type-Options"); len(got) != 1 || got[0] != "nosniff" {
		t.Errorf("X-Content-Type-Options = %v", got)
	}

	pedir(t, http.MethodGet, srv.URL+"/.well-known/jwks.json")
	esperado := []string{"POST /api/sesiones?x=1", "GET /.well-known/jwks.json"}
	if strings.Join(vistas, "|") != strings.Join(esperado, "|") {
		t.Errorf("el backend vio %v, se esperaba %v", vistas, esperado)
	}
	// El backend recibe el Host del navegador, no su nombre interno.
	for _, host := range hosts {
		if host != strings.TrimPrefix(srv.URL, "http://") {
			t.Errorf("Host reenviado = %q, se esperaba %q", host, srv.URL)
		}
	}
}

func TestBackendCaidoResponde502(t *testing.T) {
	srv := servidorDePrueba(t, "http://127.0.0.1:1")
	resp, cuerpo := pedir(t, http.MethodGet, srv.URL+"/api/sesiones/actual")
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("estado %d", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "application/problem+json" || !strings.Contains(cuerpo, `"status":502`) {
		t.Errorf("cuerpo %q (%s)", cuerpo, resp.Header.Get("Content-Type"))
	}
}

func TestConfiguracionInvalida(t *testing.T) {
	casos := map[string]string{
		"sin backend":       "",
		"sin esquema":       "backend:8080",
		"esquema no http":   "ftp://backend",
		"url que no se lee": "http://[::1",
	}
	for nombre, backend := range casos {
		if _, err := nuevoManejador(os.DirFS(t.TempDir()), backend, silencioso()); err == nil {
			t.Errorf("%s: se esperaba un error con BACKEND_URL=%q", nombre, backend)
		}
	}
}
