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
