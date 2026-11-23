package api

import (
	"errors"
	"log/slog"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/CatalinaCovarrubias/pingeso-grupo6-identity-provider/apps/api/internal/service"
)

// Etiquetas del contrato.
const (
	etiquetaSesiones   = "Sesiones"
	etiquetaUsuarios   = "Usuarios"
	etiquetaRoles      = "Roles"
	etiquetaPerfiles   = "Perfiles"
	etiquetaDesarrollo = "Desarrollo"
)

// Ejemplos de IDs: son los del seed de desarrollo.
const (
	ejemploIDUsuario = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	ejemploIDRol     = "11111111-1111-4111-8111-111111111111"
)

// IDRol es un UUID de rol dentro de un arreglo. Implementa SchemaProvider para que el
// contrato valide el formato de cada elemento.
type IDRol string

// Schema describe IDRol en el contrato.
func (IDRol) Schema(huma.Registry) *huma.Schema {
	return &huma.Schema{Type: huma.TypeString, Format: "uuid", Examples: []any{ejemploIDRol}}
}

// CodigoPermiso es un permiso con formato <recurso>:<accion>.
type CodigoPermiso string

// Schema describe CodigoPermiso en el contrato.
func (CodigoPermiso) Schema(huma.Registry) *huma.Schema {
	maximo := 80
	return &huma.Schema{
		Type:        huma.TypeString,
		Pattern:     `^[a-z][a-z0-9_]*:[a-z][a-z0-9_]*$`,
		MaxLength:   &maximo,
		Description: "Permiso con formato <recurso>:<accion>, por ejemplo reservas:operar",
		Examples:    []any{"usuarios:administrar"},
	}
}

// conMinimoUnCampo agrega minProperties: 1 a un esquema de PATCH.
func conMinimoUnCampo(s *huma.Schema) *huma.Schema {
	uno := 1
	s.MinProperties = &uno
	return s
}

// RolResumen es la referencia a un rol junto a un usuario.
type RolResumen struct {
	ID     string `json:"id" format:"uuid" example:"11111111-1111-4111-8111-111111111111"`
	Nombre string `json:"nombre" example:"administrador"`
}

// Usuario es la vista pública de una cuenta.
type Usuario struct {
	ID            string       `json:"id" format:"uuid" example:"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"`
	Correo        string       `json:"correo" format:"email" example:"admin@dev.local"`
	Estado        string       `json:"estado" enum:"activo,bloqueado,inactivo" doc:"Solo una cuenta activa puede iniciar sesión"`
	Roles         []RolResumen `json:"roles"`
	UltimoAcceso  *time.Time   `json:"ultimo_acceso,omitempty" doc:"Último inicio de sesión; falta si nunca ha ingresado"`
	CreadoEn      time.Time    `json:"creado_en"`
	ActualizadoEn time.Time    `json:"actualizado_en"`
}

// UsuarioActual es el usuario de la sesión con sus permisos efectivos.
type UsuarioActual struct {
	Usuario
	Permisos []string  `json:"permisos" example:"[\"perfiles:operar\",\"roles:administrar\",\"usuarios:administrar\"]"`
	ExpiraEn time.Time `json:"expira_en" doc:"Momento en que vence el token de la sesión"`
}

func aUsuario(d service.UsuarioDetalle) Usuario {
	return Usuario{
		ID: d.ID.String(), Correo: d.Correo, Estado: string(d.Estado), Roles: aRolesResumen(d.Roles),
		UltimoAcceso: d.UltimoAcceso, CreadoEn: d.CreadoEn, ActualizadoEn: d.ActualizadoEn,
	}
}

func aUsuarioActual(u service.UsuarioActual) UsuarioActual {
	return UsuarioActual{Usuario: aUsuario(u.UsuarioDetalle), Permisos: u.Permisos, ExpiraEn: u.ExpiraEn}
}

func aRolesResumen(roles []service.RolResumen) []RolResumen {
	resultado := make([]RolResumen, 0, len(roles))
	for _, r := range roles {
		resultado = append(resultado, RolResumen{ID: r.ID.String(), Nombre: r.Nombre})
	}
	return resultado
}

// aUUIDs convierte IDs que el contrato ya validó. Si aun así uno no se puede leer,
// queda como uuid.Nil, que no corresponde a ningún rol (el service responde 422).
func aUUIDs(ids []IDRol) []uuid.UUID {
	resultado := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		resultado = append(resultado, aUUID(string(id)))
	}
	return resultado
}

// aUUID lee un UUID sin hacer panic: un valor ilegible queda como uuid.Nil.
func aUUID(valor string) uuid.UUID {
	id, err := uuid.Parse(valor)
	if err != nil {
		return uuid.Nil
	}
	return id
}

func aTextos(permisos []CodigoPermiso) []string {
	resultado := make([]string, 0, len(permisos))
	for _, p := range permisos {
		resultado = append(resultado, string(p))
	}
	return resultado
}

// EntradaID es el parámetro {id} de las rutas. Es exportado porque Huma solo lee
// parámetros de structs embebidos exportados.
type EntradaID struct {
	ID string `path:"id" format:"uuid" doc:"Identificador UUID" example:"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"`
}

func (e EntradaID) uuid() uuid.UUID {
	return aUUID(e.ID)
}

// traducirError convierte un error del service en una respuesta RFC 9457. Los errores
// desconocidos se registran en el log y se responden como 500 sin detalles.
func traducirError(err error) error {
	var negocio *service.ErrorNegocio
	if !errors.As(err, &negocio) {
		slog.Error("error interno", "error", err)
		return huma.Error500InternalServerError("Error interno del servidor")
	}
	mensaje := negocio.Error()
	switch {
	case errors.Is(err, service.ErrNoEncontrado):
		return huma.Error404NotFound(mensaje)
	case errors.Is(err, service.ErrConflicto):
		return huma.Error409Conflict(mensaje)
	case errors.Is(err, service.ErrValidacion):
		return huma.Error422UnprocessableEntity(mensaje)
	case errors.Is(err, service.ErrCredenciales), errors.Is(err, service.ErrNoAutenticado):
		return huma.Error401Unauthorized(mensaje)
	case errors.Is(err, service.ErrCuentaNoActiva), errors.Is(err, service.ErrSinPermiso):
		return huma.Error403Forbidden(mensaje)
	default:
		slog.Error("error de negocio sin traducción", "error", err)
		return huma.Error500InternalServerError("Error interno del servidor")
	}
}
