package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// Roles administra los roles y el catálogo de permisos.
type Roles struct {
	roles    AlmacenRoles
	usuarios AlmacenUsuarios
	ahora    Reloj
}

// NuevoRoles arma el service.
func NuevoRoles(roles AlmacenRoles, usuarios AlmacenUsuarios, ahora Reloj) *Roles {
	return &Roles{roles: roles, usuarios: usuarios, ahora: ahora}
}

const (
	mensajeRolNoExiste   = "El rol no existe"
	mensajeRolProtegido  = "El rol administrador está protegido: no se modifica ni se elimina"
	mensajeNombreOcupado = "Ya existe un rol con ese nombre"
)

// Listar devuelve todos los roles.
func (s *Roles) Listar(ctx context.Context) ([]Rol, error) {
	return s.roles.ListarRoles(ctx)
}

// Obtener devuelve un rol.
func (s *Roles) Obtener(ctx context.Context, id uuid.UUID) (Rol, error) {
	r, err := s.roles.ObtenerRol(ctx, id)
	return r, noEncontrado(err, mensajeRolNoExiste)
}

// NuevoRol son los datos para crear un rol. Los permisos pueden ser de cualquier
// servicio, siempre que cumplan el formato <recurso>:<accion>.
type NuevoRol struct {
	Nombre      string
	Descripcion string
	Permisos    []string
}

// Crear registra un rol con nombre único.
func (s *Roles) Crear(ctx context.Context, n NuevoRol) (Rol, error) {
	permisos, err := normalizarPermisos(n.Permisos)
	if err != nil {
		return Rol{}, err
	}
	ahora := s.ahora()
	r := Rol{
		ID: uuid.New(), Nombre: n.Nombre, Descripcion: n.Descripcion, Permisos: permisos,
		CreadoEn: ahora, ActualizadoEn: ahora,
	}
	if err := s.roles.CrearRol(ctx, r); err != nil {
		if errors.Is(err, ErrConflicto) {
			return Rol{}, nuevoError(ErrConflicto, mensajeNombreOcupado)
		}
		return Rol{}, err
	}
	return r, nil
}

// CambiosRol son los campos editables; los nil no se tocan.
type CambiosRol struct {
	Descripcion *string
	Permisos    *[]string
}

// Actualizar cambia la descripción o los permisos. El rol administrador no se modifica,
// para que nadie le quite permisos y deje el sistema sin administración.
func (s *Roles) Actualizar(ctx context.Context, id uuid.UUID, c CambiosRol) (Rol, error) {
	if c.Descripcion == nil && c.Permisos == nil {
		return Rol{}, nuevoError(ErrValidacion, "Indica al menos un campo para cambiar")
	}
	r, err := s.roles.ObtenerRol(ctx, id)
	if err != nil {
		return Rol{}, noEncontrado(err, mensajeRolNoExiste)
	}
	if r.Nombre == RolAdministrador {
		return Rol{}, nuevoError(ErrConflicto, mensajeRolProtegido)
	}
	if c.Descripcion != nil {
		r.Descripcion = *c.Descripcion
	}
	if c.Permisos != nil {
		permisos, err := normalizarPermisos(*c.Permisos)
		if err != nil {
			return Rol{}, err
		}
		r.Permisos = permisos
	}
	r.ActualizadoEn = s.ahora()
	if err := s.roles.ActualizarRol(ctx, r); err != nil {
		return Rol{}, noEncontrado(err, mensajeRolNoExiste)
	}
	return r, nil
}

// Eliminar borra un rol que nadie tiene asignado (MF-23).
func (s *Roles) Eliminar(ctx context.Context, id uuid.UUID) error {
	r, err := s.roles.ObtenerRol(ctx, id)
	if err != nil {
		return noEncontrado(err, mensajeRolNoExiste)
	}
	if r.Nombre == RolAdministrador {
		return nuevoError(ErrConflicto, mensajeRolProtegido)
	}
	asignados, err := s.usuarios.ContarUsuariosConRol(ctx, id)
	if err != nil {
		return err
	}
	if asignados > 0 {
		return nuevoError(ErrConflicto, "El rol está asignado a usuarios: quítalo de ellos antes de eliminarlo")
	}
	return noEncontrado(s.roles.EliminarRol(ctx, id), mensajeRolNoExiste)
}

// Permisos devuelve el catálogo de permisos propios del servicio.
func (s *Roles) Permisos(_ context.Context) []Permiso {
	return CatalogoPermisos()
}
