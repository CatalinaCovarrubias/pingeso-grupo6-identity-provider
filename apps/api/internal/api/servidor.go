// Paquete api contiene los handlers HTTP (Huma) y los structs que forman el contrato.
// Solo traduce HTTP: las reglas de negocio viven en internal/service.
package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

// Titulo y Version identifican el contrato publicado en contracts/openapi.yaml.
const (
	Titulo  = "API Usuarios y Roles"
	Version = "0.1.0"
)

// VerificadorBD comprueba que la base de datos responde. pgxpool.Pool lo cumple.
type VerificadorBD interface {
	Ping(ctx context.Context) error
}

// Dependencias agrupa lo que los handlers necesitan para funcionar.
type Dependencias struct {
	BD           VerificadorBD
	PublicarDocs bool // true en dev y test: publica /api/docs y el spec
}

// NuevoServidor arma el router con todas las rutas y devuelve también la API de Huma,
// que se usa para generar el contrato OpenAPI.
func NuevoServidor(dep Dependencias) (http.Handler, huma.API) {
	mux := http.NewServeMux()

	cfg := huma.DefaultConfig(Titulo, Version)
	cfg.Info.Description = "Servicio base de la LPS: autenticación, usuarios, roles y perfiles (PT 6, grupo 6)."
	if dep.PublicarDocs {
		cfg.DocsPath = "/api/docs"
		cfg.OpenAPIPath = "/api/openapi"
	} else {
		cfg.DocsPath = ""
		cfg.OpenAPIPath = ""
		cfg.SchemasPath = ""
	}

	api := humago.New(mux, cfg)
	registrarSalud(api, dep.BD)

	return mux, api
}
