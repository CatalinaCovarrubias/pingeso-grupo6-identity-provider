package api

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

// tiempoMaximoPing es cuánto espera /readyz a que PostgreSQL conteste.
const tiempoMaximoPing = 2 * time.Second

// SaludRespuesta es la respuesta de /healthz y /readyz.
type SaludRespuesta struct {
	Body struct {
		Estado string `json:"estado" example:"ok" doc:"Estado del servicio"`
	}
}

// registrarSalud agrega los endpoints de salud, que no requieren autenticación.
func registrarSalud(api huma.API, bd VerificadorBD) {
	huma.Register(api, publica(huma.Operation{
		OperationID: "verificarVida",
		Method:      http.MethodGet,
		Path:        "/healthz",
		Summary:     "Indica si el proceso está vivo",
		Tags:        []string{"Salud"},
	}), func(ctx context.Context, _ *struct{}) (*SaludRespuesta, error) {
		resp := &SaludRespuesta{}
		resp.Body.Estado = "ok"
		return resp, nil
	})

	huma.Register(api, publica(huma.Operation{
		OperationID: "verificarDisponibilidad",
		Method:      http.MethodGet,
		Path:        "/readyz",
		Summary:     "Indica si el servicio puede atender: responde 200 solo si PostgreSQL contesta",
		Tags:        []string{"Salud"},
		Errors:      []int{http.StatusServiceUnavailable},
	}), func(ctx context.Context, _ *struct{}) (*SaludRespuesta, error) {
		if bd == nil {
			return nil, huma.Error503ServiceUnavailable("la base de datos no está configurada")
		}
		ctxPing, cancelar := context.WithTimeout(ctx, tiempoMaximoPing)
		defer cancelar()
		if err := bd.Ping(ctxPing); err != nil {
			return nil, huma.Error503ServiceUnavailable("la base de datos no responde")
		}
		resp := &SaludRespuesta{}
		resp.Body.Estado = "ok"
		return resp, nil
	})
}
