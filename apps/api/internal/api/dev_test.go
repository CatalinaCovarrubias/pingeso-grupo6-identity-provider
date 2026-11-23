package api

import (
	"net/http"
	"testing"
)

func TestTokensDePruebaNoExistenFueraDeDesarrollo(t *testing.T) {
	// Con API_ENV=test o prod, main arma el servidor con RegistrarDev=false.
	s := nuevoServidorPrueba(t, opcionesServidor{registrarDev: false})
	r := s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/dev/tokens", cuerpo: map[string]any{"roles": []string{"administrador"}}})
	if r.Code != http.StatusNotFound {
		t.Fatalf("fuera de dev la ruta no debe existir: llegó %d", r.Code)
	}
}

func TestTokenDePruebaSirveComoBearer(t *testing.T) {
	s := nuevoServidorPrueba(t, opcionesServidor{registrarDev: true})
	r := s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/dev/tokens", cuerpo: map[string]any{
		"roles": []string{"administrador"}, "permisos_extra": []string{"reservas:operar"},
		"usuario_id": ejemploIDUsuario,
	}}).esperar(http.StatusOK).json()

	claims := r["claims"].(map[string]any)
	if claims["iss"] != emisorPruebas || claims["sub"] != ejemploIDUsuario || claims["aud"].([]any)[0] != audienciaPruebas {
		t.Fatalf("claims inesperados: %v", claims)
	}
	permisos := claims["permisos"].([]any)
	if len(permisos) != 4 || permisos[1] != "reservas:operar" {
		t.Fatalf("permisos inesperados: %v", permisos)
	}

	token := r["token"].(string)
	s.llamar(peticion{metodo: http.MethodGet, ruta: "/api/usuarios", token: token}).esperar(http.StatusOK)
	s.llamar(peticion{metodo: http.MethodGet, ruta: "/api/sesiones/actual", token: token}).esperar(http.StatusOK)
}

func TestTokenDePruebaConDuracion(t *testing.T) {
	s := nuevoServidorPrueba(t, opcionesServidor{registrarDev: true})
	vigencia := func(cuerpo map[string]any) float64 {
		t.Helper()
		claims := s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/dev/tokens", cuerpo: cuerpo}).
			esperar(http.StatusOK).json()["claims"].(map[string]any)
		return claims["exp"].(float64) - claims["iat"].(float64)
	}
	casos := map[string]float64{"": 900, "1m": 60, "2h": 7200, "24h": 86400, "1440m": 86400, "90m": 5400}
	for duracion, segundos := range casos {
		cuerpo := map[string]any{}
		if duracion != "" {
			cuerpo["duracion"] = duracion
		}
		if got := vigencia(cuerpo); got != segundos {
			t.Errorf("duracion %q: se esperaban %v s, llegaron %v", duracion, segundos, got)
		}
	}
	for _, invalida := range []string{"25h", "1441m", "0m", "0h", "15", "1d", "15s", "-1h", "2h30m"} {
		s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/dev/tokens", cuerpo: map[string]any{"duracion": invalida}}).
			esperar(http.StatusUnprocessableEntity)
	}
}

func TestTokenDePruebaValida(t *testing.T) {
	s := nuevoServidorPrueba(t, opcionesServidor{registrarDev: true})
	for _, cuerpo := range []map[string]any{
		{"roles": []string{"no_existe"}},
		{"permisos_extra": []string{"Sin Formato"}},
		{"usuario_id": "no-es-uuid"},
	} {
		s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/dev/tokens", cuerpo: cuerpo}).
			esperar(http.StatusUnprocessableEntity)
	}
	sinNada := s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/dev/tokens", cuerpo: map[string]any{}}).
		esperar(http.StatusOK).json()
	if len(sinNada["claims"].(map[string]any)["permisos"].([]any)) != 0 {
		t.Fatal("sin roles ni permisos extra el token no debe tener permisos")
	}
}
