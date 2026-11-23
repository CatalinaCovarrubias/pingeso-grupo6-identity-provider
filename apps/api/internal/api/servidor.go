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
	Version = "0.2.0"
)

// VerificadorBD comprueba que la base de datos responde. pgxpool.Pool lo cumple.
type VerificadorBD interface {
	Ping(ctx context.Context) error
}

// Dependencias agrupa lo que los handlers necesitan para funcionar.
type Dependencias struct {
	BD           VerificadorBD
	PublicarDocs bool // true en dev y test: publica /api/docs y el spec
	// RegistrarDev agrega POST /api/dev/tokens. Solo debe ser true con API_ENV=dev
	// (y al generar el contrato).
	RegistrarDev bool

	Sesiones   ServicioSesiones
	Desarrollo ServicioDesarrollo
	Usuarios   ServicioUsuarios
	Roles      ServicioRoles
	Perfiles   ServicioPerfiles
}

// Nombres de los esquemas de seguridad del contrato.
const (
	esquemaCookie = "cookieSesion"
	esquemaBearer = "bearer"
)

// NuevoServidor arma el router con todas las rutas y devuelve también la API de Huma,
// que se usa para generar el contrato OpenAPI.
func NuevoServidor(dep Dependencias) (http.Handler, huma.API) {
	mux := http.NewServeMux()

	cfg := huma.DefaultConfig(Titulo, Version)
	cfg.Info.Description = "Servicio de identidad de la LPS (PT 6, grupo 6): sesiones, usuarios, roles, " +
		"perfiles y emisión de JWT RS256.\n\n" +
		"Cada operación declara en `x-permiso` lo que exige: `publico`, `autenticado` o un permiso " +
		"`<recurso>:<accion>`. Sin token válido se responde 401; con token pero sin permiso, 403.\n\n" +
		"El token viaja en la cookie `sesion` (navegador) o en `Authorization: Bearer` (entre servicios). " +
		"Los demás servicios lo validan con las llaves de `/.well-known/jwks.json`."
	if dep.PublicarDocs {
		cfg.DocsPath = "/api/docs"
		cfg.OpenAPIPath = "/api/openapi"
	} else {
		cfg.DocsPath = ""
		cfg.OpenAPIPath = ""
		cfg.SchemasPath = ""
	}
	cfg.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		esquemaCookie: {
			Type: "apiKey", In: "cookie", Name: nombreCookie,
			Description: "JWT en la cookie de sesión (HttpOnly, Secure, SameSite=Lax), que fija POST /api/sesiones.",
		},
		esquemaBearer: {
			Type: "http", Scheme: "bearer", BearerFormat: "JWT",
			Description: "JWT en `Authorization: Bearer <token>`, para llamadas entre servicios.",
		},
	}

	api := humago.New(mux, cfg)
	api.UseMiddleware(middlewareAutenticacion(api, dep.Sesiones))

	registrarSalud(api, dep.BD)
	registrarSesiones(api, dep.Sesiones)
	registrarUsuarios(api, dep.Usuarios)
	registrarRoles(api, dep.Roles)
	registrarPerfiles(api, dep.Perfiles)

	return mux, api
}
