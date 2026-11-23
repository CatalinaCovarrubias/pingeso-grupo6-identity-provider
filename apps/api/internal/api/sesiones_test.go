package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/CatalinaCovarrubias/pingeso-grupo6-identity-provider/apps/api/internal/service/tokens"
)

func TestLoginCorrectoFijaLaCookie(t *testing.T) {
	s := nuevoServidorPrueba(t)
	r := s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/sesiones", cuerpo: map[string]string{
		"correo": "admin@dev.local", "contrasena": "admin-dev-local",
	}}).esperar(http.StatusOK)

	cookies := r.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("se esperaba una cookie, llegaron %d", len(cookies))
	}
	c := cookies[0]
	if c.Name != "sesion" || !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteLaxMode || c.Path != "/" {
		t.Fatalf("atributos de la cookie incorrectos: %+v", c)
	}
	if c.MaxAge < 890 || c.MaxAge > 900 {
		t.Fatalf("Max-Age debe ser de unos 900 s, es %d", c.MaxAge)
	}

	cuerpo := r.json()
	if cuerpo["correo"] != "admin@dev.local" || cuerpo["id"] != ejemploIDUsuario {
		t.Fatalf("cuerpo inesperado: %v", cuerpo)
	}
	if strings.Contains(r.Body.String(), c.Value) {
		t.Fatal("el login no debe devolver el token en el body")
	}
}

func TestLoginIncorrecto(t *testing.T) {
	s := nuevoServidorPrueba(t)
	malaClave := s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/sesiones", cuerpo: map[string]string{
		"correo": "admin@dev.local", "contrasena": "otra",
	}}).esperar(http.StatusUnauthorized)
	sinCuenta := s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/sesiones", cuerpo: map[string]string{
		"correo": "nadie@dev.local", "contrasena": "otra",
	}}).esperar(http.StatusUnauthorized)
	if malaClave.json()["detail"] != sinCuenta.json()["detail"] {
		t.Fatal("correo inexistente y contraseña mala deben dar el mismo mensaje")
	}
	if len(malaClave.Result().Cookies()) != 0 {
		t.Fatal("un login fallido no debe fijar cookie")
	}
	s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/sesiones", cuerpo: map[string]string{
		"correo": "no-es-correo", "contrasena": "x",
	}}).esperar(http.StatusUnprocessableEntity)
}

func TestLoginConCuentaBloqueada(t *testing.T) {
	s := nuevoServidorPrueba(t)
	s.llamar(peticion{metodo: http.MethodPatch, ruta: "/api/usuarios/cccccccc-cccc-4ccc-8ccc-cccccccccccc",
		cookie: s.comoAdmin(), cuerpo: map[string]string{"estado": "bloqueado"}}).esperar(http.StatusOK)
	s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/sesiones", cuerpo: map[string]string{
		"correo": "usuario@dev.local", "contrasena": "usuario-dev-local",
	}}).esperar(http.StatusForbidden)
}

func TestSesionActualConCookieYConBearer(t *testing.T) {
	s := nuevoServidorPrueba(t)
	cookie := s.comoOperador()

	porCookie := s.llamar(peticion{metodo: http.MethodGet, ruta: "/api/sesiones/actual", cookie: cookie}).
		esperar(http.StatusOK).json()
	if porCookie["correo"] != "operador@dev.local" {
		t.Fatalf("usuario inesperado: %v", porCookie)
	}
	permisos, _ := porCookie["permisos"].([]any)
	if len(permisos) != 1 || permisos[0] != "perfiles:operar" {
		t.Fatalf("permisos inesperados: %v", porCookie["permisos"])
	}

	porBearer := s.llamar(peticion{metodo: http.MethodGet, ruta: "/api/sesiones/actual", token: cookie.Value}).
		esperar(http.StatusOK).json()
	if porBearer["id"] != porCookie["id"] {
		t.Fatal("el Bearer debe identificar al mismo usuario que la cookie")
	}
}

func TestRechazaTokensInvalidos(t *testing.T) {
	s := nuevoServidorPrueba(t)
	valido := s.comoAdmin().Value

	claims := jwt.RegisteredClaims{
		Issuer: emisorPruebas, Subject: ejemploIDUsuario, Audience: jwt.ClaimStrings{audienciaPruebas},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)), IssuedAt: jwt.NewNumericDate(time.Now()),
	}
	none := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	none.Header["kid"] = llavePruebas.Kid()
	tokenNone, _ := none.SignedString(jwt.UnsafeAllowNoneSignatureType)

	vencido, _, _ := tokens.NuevoEmisor(llavePruebas, emisorPruebas, audienciaPruebas, 15*time.Minute).
		ConReloj(func() time.Time { return time.Now().Add(-time.Hour) }).
		Emitir(ejemploIDUsuario, []string{"administrador"}, []string{"usuarios:administrar"})
	otraAudiencia, _, _ := tokens.NuevoEmisor(llavePruebas, emisorPruebas, "otro-curso", 15*time.Minute).
		Emitir(ejemploIDUsuario, nil, nil)

	casos := map[string]peticion{
		"sin token":      {},
		"sin firma":      {token: strings.Join(strings.Split(valido, ".")[:2], ".") + "."},
		"alg none":       {token: tokenNone},
		"vencido":        {token: vencido},
		"otra audiencia": {token: otraAudiencia},
		"cookie basura":  {cookie: &http.Cookie{Name: "sesion", Value: "basura"}},
		"esquema Basic": {header: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("a:b")),
		}},
	}
	for nombre, p := range casos {
		t.Run(nombre, func(t *testing.T) {
			p.metodo, p.ruta = http.MethodGet, "/api/sesiones/actual"
			s.llamar(p).esperar(http.StatusUnauthorized)
		})
	}
}

func TestUsuarioSinPermisoRecibe403(t *testing.T) {
	s := nuevoServidorPrueba(t)
	r := s.llamar(peticion{metodo: http.MethodGet, ruta: "/api/usuarios", cookie: s.comoUsuario()}).
		esperar(http.StatusForbidden)
	if !strings.Contains(r.json()["detail"].(string), "usuarios:administrar") {
		t.Fatalf("el 403 debe nombrar el permiso que falta: %s", r.Body.String())
	}
	// El operador tampoco puede asignar roles (MF-21).
	s.llamar(peticion{metodo: http.MethodPut, ruta: "/api/usuarios/bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb/roles",
		cookie: s.comoOperador(), cuerpo: map[string]any{"roles": []string{ejemploIDRol}}}).
		esperar(http.StatusForbidden)
}

func TestJWKSPublicaElKidDeFirma(t *testing.T) {
	s := nuevoServidorPrueba(t)
	token := s.comoAdmin().Value

	r := s.llamar(peticion{metodo: http.MethodGet, ruta: "/.well-known/jwks.json"}).esperar(http.StatusOK)
	if r.Header().Get("Cache-Control") != "public, max-age=300" {
		t.Fatalf("Cache-Control inesperado: %q", r.Header().Get("Cache-Control"))
	}
	var jwks struct {
		Keys []JWK `json:"keys"`
	}
	if err := json.Unmarshal(r.Body.Bytes(), &jwks); err != nil || len(jwks.Keys) != 1 {
		t.Fatalf("JWKS inesperado: %s", r.Body.String())
	}

	crudo, _ := base64.RawURLEncoding.DecodeString(strings.Split(token, ".")[0])
	var cabecera map[string]string
	_ = json.Unmarshal(crudo, &cabecera)
	if cabecera["kid"] != jwks.Keys[0].Kid || jwks.Keys[0].Alg != "RS256" {
		t.Fatalf("el kid del token (%s) debe estar en el JWKS (%+v)", cabecera["kid"], jwks.Keys[0])
	}
}

func TestCerrarSesionBorraLaCookie(t *testing.T) {
	s := nuevoServidorPrueba(t)
	r := s.llamar(peticion{metodo: http.MethodDelete, ruta: "/api/sesiones/actual", cookie: s.comoUsuario()}).
		esperar(http.StatusNoContent)
	c := r.Result().Cookies()
	if len(c) != 1 || c[0].Name != "sesion" || c[0].MaxAge >= 0 || c[0].Value != "" {
		t.Fatalf("la cookie debe borrarse: %+v", c)
	}
	if !strings.Contains(r.Header().Get("Set-Cookie"), "Max-Age=0") {
		t.Fatalf("Set-Cookie debe llevar Max-Age=0: %q", r.Header().Get("Set-Cookie"))
	}
}

func TestRenovarYCambiarContrasena(t *testing.T) {
	s := nuevoServidorPrueba(t)
	cookie := s.comoUsuario()

	r := s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/sesiones/actual/renovacion", cookie: cookie}).
		esperar(http.StatusOK)
	nueva := r.Result().Cookies()
	if len(nueva) != 1 || nueva[0].Value == "" || r.json()["expira_en"] == nil {
		t.Fatalf("la renovación debe fijar una cookie y devolver expira_en: %s", r.Body.String())
	}

	s.llamar(peticion{metodo: http.MethodPut, ruta: "/api/sesiones/actual/contrasena", cookie: cookie,
		cuerpo: map[string]string{"contrasena_actual": "usuario-dev-local", "contrasena_nueva": "otra-clave-larga"}}).
		esperar(http.StatusNoContent)
	s.llamar(peticion{metodo: http.MethodPut, ruta: "/api/sesiones/actual/contrasena", cookie: cookie,
		cuerpo: map[string]string{"contrasena_actual": "x", "contrasena_nueva": "corta"}}).
		esperar(http.StatusUnprocessableEntity)
}

func TestTokenDeUsuarioInexistenteNoSirveEnSesionActual(t *testing.T) {
	s := nuevoServidorPrueba(t)
	token, _, _ := s.emisor.Emitir("dddddddd-dddd-4ddd-8ddd-dddddddddddd", nil, nil)
	s.llamar(peticion{metodo: http.MethodGet, ruta: "/api/sesiones/actual", token: token}).
		esperar(http.StatusUnauthorized)
}

func TestErrorDesconocidoEs500SinDetalles(t *testing.T) {
	err := traducirError(errors.New("pq: relation \"usuarios\" does not exist"))
	if !strings.Contains(err.Error(), "Error interno del servidor") || strings.Contains(err.Error(), "relation") {
		t.Fatalf("el 500 no debe exponer detalles: %v", err)
	}
}
