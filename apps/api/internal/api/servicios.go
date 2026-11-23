package api

import (
	"context"

	"github.com/google/uuid"

	"github.com/CatalinaCovarrubias/pingeso-grupo6-identity-provider/apps/api/internal/service"
	"github.com/CatalinaCovarrubias/pingeso-grupo6-identity-provider/apps/api/internal/service/tokens"
)

// Interfaces de los services que usan los handlers. Los handlers no conocen el store.

// ServicioSesiones inicia y cierra sesiones y valida tokens.
type ServicioSesiones interface {
	Iniciar(ctx context.Context, correo, contrasena string) (service.SesionIniciada, error)
	Autenticar(ctx context.Context, token string) (service.Sesion, error)
	Actual(ctx context.Context) (service.UsuarioActual, error)
	Renovar(ctx context.Context) (service.TokenEmitido, error)
	CambiarContrasena(ctx context.Context, actual, nueva string) error
	JWKS() []tokens.JWK
}

// ServicioDesarrollo emite tokens de prueba. Solo existe con API_ENV=dev.
type ServicioDesarrollo interface {
	EmitirTokenPrueba(ctx context.Context, sol service.SolicitudTokenPrueba) (service.TokenEmitido, error)
}

// ServicioUsuarios administra cuentas y sus roles.
type ServicioUsuarios interface {
	Listar(ctx context.Context, f service.FiltroUsuarios) (service.PaginaUsuarios, error)
	Obtener(ctx context.Context, id uuid.UUID) (service.UsuarioDetalle, error)
	Crear(ctx context.Context, n service.NuevoUsuario) (service.UsuarioDetalle, error)
	Actualizar(ctx context.Context, id uuid.UUID, c service.CambiosUsuario) (service.UsuarioDetalle, error)
	Desactivar(ctx context.Context, id uuid.UUID) error
	Roles(ctx context.Context, id uuid.UUID) ([]service.RolResumen, error)
	AsignarRoles(ctx context.Context, id uuid.UUID, rolIDs []uuid.UUID) ([]service.RolResumen, error)
}

// ServicioRoles administra roles y el catálogo de permisos.
type ServicioRoles interface {
	Listar(ctx context.Context) ([]service.Rol, error)
	Obtener(ctx context.Context, id uuid.UUID) (service.Rol, error)
	Crear(ctx context.Context, n service.NuevoRol) (service.Rol, error)
	Actualizar(ctx context.Context, id uuid.UUID, c service.CambiosRol) (service.Rol, error)
	Eliminar(ctx context.Context, id uuid.UUID) error
	Permisos(ctx context.Context) []service.Permiso
}

// ServicioPerfiles administra perfiles.
type ServicioPerfiles interface {
	Obtener(ctx context.Context, usuarioID uuid.UUID) (service.Perfil, error)
	Guardar(ctx context.Context, p service.Perfil) (service.Perfil, error)
}

// Comprobación en compilación de que los services cumplen las interfaces.
var (
	_ ServicioSesiones   = (*service.Sesiones)(nil)
	_ ServicioDesarrollo = (*service.Sesiones)(nil)
	_ ServicioUsuarios   = (*service.Usuarios)(nil)
	_ ServicioRoles      = (*service.Roles)(nil)
	_ ServicioPerfiles   = (*service.Perfiles)(nil)
)
