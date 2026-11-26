// spa-server sirve la app web compilada y reenvía la API al backend.
//
//   - Sirve los archivos de /srv. Una ruta que no es un archivo responde index.html (SPA).
//   - /assets/* lleva nombres con hash: se cachea un año como inmutable.
//   - /api/* y /.well-known/* se reenvían a BACKEND_URL, así el navegador ve un solo
//     origen y la cookie de sesión funciona sin CORS.
//   - GET /healthz responde si el proceso vive. `spa-server salud` lo consulta y sirve de
//     healthcheck en la imagen distroless, que no tiene shell ni curl.
//
// Solo usa la biblioteca estándar.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"
)

// version se fija al compilar con -ldflags "-X main.version=...".
var version = "dev"

const (
	direccion      = ":8080"
	raizPorDefecto = "/srv"
	urlSalud       = "http://127.0.0.1:8080/healthz"
	timeoutBackend = 30 * time.Second
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "salud" {
		os.Exit(salud(urlSalud, 3*time.Second))
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).
		With("servicio", "web", "version", version)

	raiz := os.Getenv("SPA_RAIZ")
	if raiz == "" {
		raiz = raizPorDefecto
	}
	raizSegura, err := os.OpenRoot(raiz)
	if err != nil {
		logger.Error("no se puede abrir la raíz", "raiz", raiz, "error", err.Error())
		os.Exit(1)
	}
	defer raizSegura.Close()
	manejador, err := nuevoManejador(raizSegura.FS(), os.Getenv("BACKEND_URL"), logger)
	if err != nil {
		logger.Error("configuración inválida", "error", err.Error())
		os.Exit(1)
	}

	servidor := &http.Server{
		Addr:              direccion,
		Handler:           manejador,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	ctx, cancelar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelar()

	errores := make(chan error, 1)
	go func() { errores <- servidor.ListenAndServe() }()
	logger.Info("escuchando", "direccion", direccion, "raiz", raiz)

	select {
	case err := <-errores:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("el servidor se detuvo", "error", err.Error())
			os.Exit(1)
		}
	case <-ctx.Done():
		apagado, listo := context.WithTimeout(context.Background(), 10*time.Second)
		defer listo()
		if err := servidor.Shutdown(apagado); err != nil {
			logger.Error("apagado incompleto", "error", err.Error())
		}
	}
}

// nuevoManejador arma las rutas. archivos debe venir de os.Root, que impide salir de la
// raíz (también con enlaces simbólicos). Falla si BACKEND_URL no es válida.
func nuevoManejador(archivos fs.FS, backendURL string, logger *slog.Logger) (http.Handler, error) {
	destino, err := validarBackend(backendURL)
	if err != nil {
		return nil, err
	}

	proxy := nuevoProxy(destino, logger)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(`{"estado":"ok"}`))
	})
	mux.Handle("/api/", proxy)
	mux.Handle("/.well-known/", proxy)
	mux.Handle("/", conCabecerasSeguras(estaticos(archivos)))

	return conRegistro(mux, logger), nil
}

func validarBackend(backendURL string) (*url.URL, error) {
	if backendURL == "" {
		return nil, errors.New("falta BACKEND_URL (por ejemplo http://backend:8080)")
	}
	destino, err := url.Parse(backendURL)
	if err != nil || (destino.Scheme != "http" && destino.Scheme != "https") || destino.Host == "" {
		return nil, fmt.Errorf("BACKEND_URL no es una URL http(s) válida: %q", backendURL)
	}
	return destino, nil
}

// cabecerasSeguras se aplican a todo lo que sirve el spa-server, también a lo que viene
// del backend.
func cabecerasSeguras(h http.Header) {
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
}

func conCabecerasSeguras(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cabecerasSeguras(w.Header())
		// La app no se debe mostrar dentro de un iframe ajeno (clickjacking en el login).
		w.Header().Set("Content-Security-Policy", "frame-ancestors 'none'")
		siguiente.ServeHTTP(w, r)
	})
}

func nuevoProxy(destino *url.URL, logger *slog.Logger) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(destino)
			r.SetXForwarded()
			// Se conserva el Host del navegador: así el backend no publica su nombre interno
			// (por ejemplo en los $schema de sus respuestas).
			r.Out.Host = r.In.Host
		},
		Transport: &http.Transport{
			Proxy:                 nil,
			DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
			ResponseHeaderTimeout: timeoutBackend,
			IdleConnTimeout:       90 * time.Second,
			MaxIdleConnsPerHost:   16,
		},
		ModifyResponse: func(respuesta *http.Response) error {
			cabecerasSeguras(respuesta.Header)
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			logger.Error("backend no disponible", "ruta", r.URL.Path, "error", err.Error())
			cabecerasSeguras(w.Header())
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"type":"about:blank","title":"Bad Gateway","status":502,` +
				`"detail":"El servicio no está disponible en este momento."}`))
		},
	}
}

// estaticos sirve la app. Los archivos de /assets/ llevan hash en el nombre y se cachean
// para siempre; el resto (index.html sobre todo) se revalida en cada visita.
func estaticos(archivos fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
			return
		}

		nombre := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if strings.HasPrefix(nombre, "assets/") {
			if !servirArchivo(w, r, archivos, nombre, "public, max-age=31536000, immutable") {
				http.NotFound(w, r)
			}
			return
		}
		if nombre != "" && servirArchivo(w, r, archivos, nombre, "no-cache") {
			return
		}
		if !servirArchivo(w, r, archivos, "index.html", "no-cache") {
			http.NotFound(w, r)
		}
	})
}

// servirArchivo responde un archivo regular. Devuelve false si no existe o es un directorio.
func servirArchivo(w http.ResponseWriter, r *http.Request, archivos fs.FS, nombre, cache string) bool {
	info, err := fs.Stat(archivos, nombre)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	archivo, err := archivos.Open(nombre)
	if err != nil {
		return false
	}
	defer archivo.Close()

	contenido, ok := archivo.(interface {
		Read([]byte) (int, error)
		Seek(int64, int) (int64, error)
	})
	if !ok {
		return false
	}
	w.Header().Set("Cache-Control", cache)
	http.ServeContent(w, r, info.Name(), info.ModTime(), contenido)
	return true
}

// conRegistro escribe un log JSON por request y propaga X-Request-Id al backend.
func conRegistro(siguiente http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		id := r.Header.Get("X-Request-Id")
		if id == "" || len(id) > 128 {
			id = nuevoID()
			r.Header.Set("X-Request-Id", id)
		}
		w.Header().Set("X-Request-Id", id)

		registro := &registroEstado{ResponseWriter: w, estado: http.StatusOK}
		siguiente.ServeHTTP(registro, r)

		if r.URL.Path == "/healthz" {
			return
		}
		logger.Info("request",
			"request_id", id,
			"metodo", r.Method,
			"ruta", r.URL.Path,
			"estado", registro.estado,
			"duracion_ms", time.Since(inicio).Milliseconds(),
		)
	})
}

type registroEstado struct {
	http.ResponseWriter
	estado int
}

func (r *registroEstado) WriteHeader(estado int) {
	r.estado = estado
	r.ResponseWriter.WriteHeader(estado)
}

func (r *registroEstado) Unwrap() http.ResponseWriter { return r.ResponseWriter }

func nuevoID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// salud consulta /healthz y devuelve el código de salida: 0 si responde 200.
func salud(direccionSalud string, timeout time.Duration) int {
	cliente := http.Client{Timeout: timeout}
	respuesta, err := cliente.Get(direccionSalud)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sin respuesta:", err)
		return 1
	}
	defer respuesta.Body.Close()
	if respuesta.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, "estado inesperado:", respuesta.StatusCode)
		return 1
	}
	return 0
}
