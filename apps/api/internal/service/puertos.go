package service

import (
	"context"

	"github.com/google/uuid"
)

// Interfaces de almacenamiento que consume el service. Hoy las implementa
// internal/store/memoria; en la Tanda 4 las implementará el store de PostgreSQL.
//
// Contrato común: si el registro no existe se devuelve ErrNoEncontrado, y si choca
// con una restricción de unicidad (correo o nombre de rol) se devuelve ErrConflicto.
// Los métodos devuelven copias: modificar lo recibido no altera lo guardado.

// AlmacenUsuarios guarda las cuentas.
type AlmacenUsuarios interface {
	ObtenerUsuario(ctx context.Context, id uuid.UUID) (Usuario, error)
	// BuscarUsuarioPorCorreo no distingue mayúsculas.
	BuscarUsuarioPorCorreo(ctx context.Context, correo string) (Usuario, error)
	// ListarUsuarios devuelve la página pedida, ordenada por fecha de creación, y el total.
	ListarUsuarios(ctx context.Context, filtro FiltroUsuarios) ([]Usuario, int, error)
	// CrearUsuario rechaza con ErrConflicto un correo repetido sin distinguir mayúsculas.
	CrearUsuario(ctx context.Context, u Usuario) error
	ActualizarUsuario(ctx context.Context, u Usuario) error
	ContarUsuariosConRol(ctx context.Context, rolID uuid.UUID) (int, error)
	ContarUsuariosActivosConRol(ctx context.Context, rolID uuid.UUID) (int, error)
}

// AlmacenRoles guarda los roles.
type AlmacenRoles interface {
	ListarRoles(ctx context.Context) ([]Rol, error)
	ObtenerRol(ctx context.Context, id uuid.UUID) (Rol, error)
	BuscarRolPorNombre(ctx context.Context, nombre string) (Rol, error)
	CrearRol(ctx context.Context, r Rol) error
	ActualizarRol(ctx context.Context, r Rol) error
	EliminarRol(ctx context.Context, id uuid.UUID) error
}

// AlmacenPerfiles guarda un perfil por usuario.
type AlmacenPerfiles interface {
	ObtenerPerfil(ctx context.Context, usuarioID uuid.UUID) (Perfil, error)
	GuardarPerfil(ctx context.Context, p Perfil) error
}
