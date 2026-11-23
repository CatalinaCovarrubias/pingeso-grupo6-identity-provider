package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/google/uuid"
)

// patronPermiso es el formato <recurso>:<accion> que aceptan los roles y los tokens.
var patronPermiso = regexp.MustCompile(`^[a-z][a-z0-9_]*:[a-z][a-z0-9_]*$`)

// Largos de contraseña: bcrypt ignora lo que pasa de 72 bytes.
const (
	largoMinimoContrasena = 8
	largoMaximoContrasena = 72
)

// normalizarCorreo quita espacios y pasa a minúsculas: el correo es único sin distinguirlas.
func normalizarCorreo(correo string) string {
	return strings.ToLower(strings.TrimSpace(correo))
}

func validarContrasenaNueva(contrasena string) error {
	if len(contrasena) < largoMinimoContrasena || len(contrasena) > largoMaximoContrasena {
		return nuevoError(ErrValidacion, fmt.Sprintf("La contraseña debe tener entre %d y %d caracteres",
			largoMinimoContrasena, largoMaximoContrasena))
	}
	return nil
}

// normalizarPermisos valida el formato, quita repetidos y ordena.
func normalizarPermisos(permisos []string) ([]string, error) {
	for _, p := range permisos {
		if !patronPermiso.MatchString(p) {
			return nil, nuevoError(ErrValidacion, fmt.Sprintf("El permiso %q no tiene el formato <recurso>:<accion>", p))
		}
	}
	return unicosOrdenados(permisos), nil
}

// unicosOrdenados devuelve una copia ordenada y sin repetidos; nunca nil.
func unicosOrdenados(valores []string) []string {
	resultado := append([]string{}, valores...)
	slices.Sort(resultado)
	return slices.Compact(resultado)
}

// rolesDeUsuario resuelve los roles de un usuario. Ignora IDs de roles que ya no existen.
func rolesDeUsuario(ctx context.Context, roles AlmacenRoles, ids []uuid.UUID) ([]Rol, error) {
	resultado := make([]Rol, 0, len(ids))
	for _, id := range ids {
		r, err := roles.ObtenerRol(ctx, id)
		if errors.Is(err, ErrNoEncontrado) {
			continue
		}
		if err != nil {
			return nil, err
		}
		resultado = append(resultado, r)
	}
	slices.SortFunc(resultado, func(a, b Rol) int { return strings.Compare(a.Nombre, b.Nombre) })
	return resultado, nil
}

// verificarRolesExisten comprueba que cada ID corresponda a un rol y quita repetidos.
func verificarRolesExisten(ctx context.Context, roles AlmacenRoles, ids []uuid.UUID) ([]uuid.UUID, error) {
	unicos := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if slices.Contains(unicos, id) {
			continue
		}
		if _, err := roles.ObtenerRol(ctx, id); err != nil {
			if errors.Is(err, ErrNoEncontrado) {
				return nil, nuevoError(ErrValidacion, fmt.Sprintf("El rol %s no existe", id))
			}
			return nil, err
		}
		unicos = append(unicos, id)
	}
	return unicos, nil
}

func resumir(roles []Rol) []RolResumen {
	resumen := make([]RolResumen, 0, len(roles))
	for _, r := range roles {
		resumen = append(resumen, RolResumen{ID: r.ID, Nombre: r.Nombre})
	}
	return resumen
}

func nombres(roles []Rol) []string {
	resultado := make([]string, 0, len(roles))
	for _, r := range roles {
		resultado = append(resultado, r.Nombre)
	}
	return resultado
}

// permisosEfectivos es la unión ordenada de los permisos de los roles.
func permisosEfectivos(roles []Rol) []string {
	var todos []string
	for _, r := range roles {
		todos = append(todos, r.Permisos...)
	}
	return unicosOrdenados(todos)
}

// detalle arma la vista pública del usuario.
func detalle(ctx context.Context, roles AlmacenRoles, u Usuario) (UsuarioDetalle, []Rol, error) {
	rs, err := rolesDeUsuario(ctx, roles, u.RolIDs)
	if err != nil {
		return UsuarioDetalle{}, nil, err
	}
	return UsuarioDetalle{
		ID: u.ID, Correo: u.Correo, Estado: u.Estado, Roles: resumir(rs),
		UltimoAcceso: u.UltimoAcceso, CreadoEn: u.CreadoEn, ActualizadoEn: u.ActualizadoEn,
	}, rs, nil
}

// noEncontrado traduce ErrNoEncontrado del store a un mensaje para el cliente.
func noEncontrado(err error, mensaje string) error {
	if errors.Is(err, ErrNoEncontrado) {
		return nuevoError(ErrNoEncontrado, mensaje)
	}
	return err
}
