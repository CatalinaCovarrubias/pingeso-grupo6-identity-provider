package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/PINGESO-CIVIL/usuarios-roles/apps/api/internal/service"
	"github.com/PINGESO-CIVIL/usuarios-roles/apps/api/internal/service/tokens"
	"github.com/PINGESO-CIVIL/usuarios-roles/apps/api/internal/store/memoria"
)

var (
	llavePruebas *tokens.Llave
	unaLlave     sync.Once
)

const (
	emisorPruebas    = "pingeso-identidad"
	audienciaPruebas = "pingeso"
)

// servidorPrueba es la API completa sobre un store en memoria con el seed.
type servidorPrueba struct {
	t       *testing.T
	handler http.Handler
	emisor  *tokens.Emisor
}

type opcionesServidor struct {
	registrarDev bool
}

func nuevoServidorPrueba(t *testing.T, opciones ...opcionesServidor) *servidorPrueba {
	t.Helper()
	unaLlave.Do(func() {
		var err error
		if llavePruebas, err = tokens.GenerarLlave(); err != nil {
			t.Fatal(err)
		}
	})
	var op opcionesServidor
	if len(opciones) > 0 {
		op = opciones[0]
	}
	m := memoria.Nueva()
	if err := memoria.CargarSeed(context.Background(), m, bcrypt.MinCost); err != nil {
		t.Fatal(err)
	}
	emisor := tokens.NuevoEmisor(llavePruebas, emisorPruebas, audienciaPruebas, 15*time.Minute)
	sesiones, err := service.NuevoSesiones(m, m, emisor, bcrypt.MinCost, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	handler, _ := NuevoServidor(Dependencias{
		PublicarDocs: true,
		RegistrarDev: op.registrarDev,
		Sesiones:     sesiones,
		Desarrollo:   sesiones,
		Usuarios:     service.NuevoUsuarios(m, m, bcrypt.MinCost, time.Now),
		Roles:        service.NuevoRoles(m, m, time.Now),
		Perfiles:     service.NuevoPerfiles(m, m, time.Now),
	})
	return &servidorPrueba{t: t, handler: handler, emisor: emisor}
}

// peticion describe una llamada de prueba.
type peticion struct {
	metodo string
	ruta   string
	cuerpo any
	token  string       // Authorization: Bearer
	cookie *http.Cookie // cookie de sesión
	header map[string]string
}

type respuesta struct {
	*httptest.ResponseRecorder
	t *testing.T
}

func (s *servidorPrueba) llamar(p peticion) respuesta {
	s.t.Helper()
	var cuerpo io.Reader
	if p.cuerpo != nil {
		crudo, err := json.Marshal(p.cuerpo)
		if err != nil {
			s.t.Fatal(err)
		}
		cuerpo = bytes.NewReader(crudo)
	}
	req := httptest.NewRequest(p.metodo, p.ruta, cuerpo)
	if p.cuerpo != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if p.token != "" {
		req.Header.Set("Authorization", "Bearer "+p.token)
	}
	if p.cookie != nil {
		req.AddCookie(p.cookie)
	}
	for nombre, valor := range p.header {
		req.Header.Set(nombre, valor)
	}
	rec := httptest.NewRecorder()
	s.handler.ServeHTTP(rec, req)
	return respuesta{ResponseRecorder: rec, t: s.t}
}

// iniciarSesion hace login con un usuario seed y devuelve la cookie.
func (s *servidorPrueba) iniciarSesion(correo, contrasena string) *http.Cookie {
	s.t.Helper()
	r := s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/sesiones", cuerpo: map[string]string{
		"correo": correo, "contrasena": contrasena,
	}})
	r.esperar(http.StatusOK)
	for _, c := range r.Result().Cookies() {
		if c.Name == "sesion" {
			return c
		}
	}
	s.t.Fatal("el login no fijó la cookie sesion")
	return nil
}

func (s *servidorPrueba) comoAdmin() *http.Cookie {
	return s.iniciarSesion("admin@dev.local", "admin-dev-local")
}

func (s *servidorPrueba) comoOperador() *http.Cookie {
	return s.iniciarSesion("operador@dev.local", "operador-dev-local")
}

func (s *servidorPrueba) comoUsuario() *http.Cookie {
	return s.iniciarSesion("usuario@dev.local", "usuario-dev-local")
}

func (r respuesta) esperar(estado int) respuesta {
	r.t.Helper()
	if r.Code != estado {
		r.t.Fatalf("se esperaba %d, llegó %d: %s", estado, r.Code, r.Body.String())
	}
	if estado >= 400 && r.Header().Get("Content-Type") != "application/problem+json" {
		r.t.Fatalf("los errores deben ser application/problem+json, llegó %q", r.Header().Get("Content-Type"))
	}
	return r
}

func (r respuesta) json() map[string]any {
	r.t.Helper()
	var datos map[string]any
	if err := json.Unmarshal(r.Body.Bytes(), &datos); err != nil {
		r.t.Fatalf("respuesta no es JSON: %v: %s", err, r.Body.String())
	}
	return datos
}
