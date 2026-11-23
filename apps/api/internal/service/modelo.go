// Paquete service contiene las reglas de negocio del servicio de identidad.
// No conoce HTTP ni SQL: recibe y entrega tipos propios y habla con el almacenamiento
// a través de las interfaces de puertos.go.
package service

import (
	"slices"
	"time"

	"github.com/google/uuid"
)

// EstadoUsuario indica si una cuenta puede iniciar sesión.
type EstadoUsuario string

// Estados posibles de una cuenta. Solo una cuenta activa puede iniciar sesión.
const (
	EstadoActivo    EstadoUsuario = "activo"
	EstadoBloqueado EstadoUsuario = "bloqueado"
	EstadoInactivo  EstadoUsuario = "inactivo"
)

// Valido indica si el estado es uno de los conocidos.
func (e EstadoUsuario) Valido() bool {
	return e == EstadoActivo || e == EstadoBloqueado || e == EstadoInactivo
}

// Permisos propios del servicio de identidad, en formato <recurso>:<accion>.
const (
	PermisoUsuariosAdministrar = "usuarios:administrar"
	PermisoRolesAdministrar    = "roles:administrar"
	PermisoPerfilesOperar      = "perfiles:operar"
)

// RolAdministrador es el rol protegido: no se modifica ni se elimina, y siempre
// debe quedar al menos un usuario activo que lo tenga.
const RolAdministrador = "administrador"

// Permiso describe un permiso del catálogo.
type Permiso struct {
	Codigo      string
	Descripcion string
}

var catalogoPermisos = []Permiso{
	{PermisoPerfilesOperar, "Ver y editar el perfil de cualquier usuario"},
	{PermisoRolesAdministrar, "Crear, editar y eliminar roles, y asignarlos a usuarios"},
	{PermisoUsuariosAdministrar, "Crear, editar, desactivar y listar usuarios"},
}

// CatalogoPermisos devuelve los permisos propios del servicio de identidad.
// Los roles pueden incluir además permisos de otros servicios (por ejemplo reservas:operar).
func CatalogoPermisos() []Permiso {
	return slices.Clone(catalogoPermisos)
}

// Usuario es una cuenta tal como se guarda. HashContrasena nunca sale del service.
type Usuario struct {
	ID             uuid.UUID
	Correo         string
	HashContrasena []byte
	Estado         EstadoUsuario
	RolIDs         []uuid.UUID
	UltimoAcceso   *time.Time
	CreadoEn       time.Time
	ActualizadoEn  time.Time
}

// Rol agrupa permisos.
type Rol struct {
	ID            uuid.UUID
	Nombre        string
	Descripcion   string
	Permisos      []string
	CreadoEn      time.Time
	ActualizadoEn time.Time
}

// RolResumen es la referencia a un rol que se muestra junto a un usuario.
type RolResumen struct {
	ID     uuid.UUID
	Nombre string
}

// UsuarioDetalle es lo que el service entrega de un usuario: sin hash y con los roles resueltos.
type UsuarioDetalle struct {
	ID            uuid.UUID
	Correo        string
	Estado        EstadoUsuario
	Roles         []RolResumen
	UltimoAcceso  *time.Time
	CreadoEn      time.Time
	ActualizadoEn time.Time
}

// UsuarioActual es el usuario de la sesión, con los permisos de su token.
type UsuarioActual struct {
	UsuarioDetalle
	Permisos []string
	ExpiraEn time.Time
}

// Perfil es la ficha genérica de un usuario. Atributos guarda los campos propios de cada cliente.
type Perfil struct {
	UsuarioID     uuid.UUID
	Nombre        string
	Apellido      string
	Telefono      *string
	FotoURL       *string
	Atributos     map[string]any
	ActualizadoEn time.Time
}

// FiltroUsuarios define la página y los filtros al listar usuarios.
type FiltroUsuarios struct {
	Pagina int // desde 1
	Limite int // entre 1 y 100
	Estado EstadoUsuario
	Correo string // coincidencia parcial, sin distinguir mayúsculas
}

// PaginaUsuarios es una página del listado y el total de coincidencias.
type PaginaUsuarios struct {
	Datos []UsuarioDetalle
	Total int
}

// Reloj entrega la hora actual. Se reemplaza en pruebas.
type Reloj func() time.Time
