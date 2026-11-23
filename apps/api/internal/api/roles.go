package api

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/CatalinaCovarrubias/pingeso-grupo6-identity-provider/apps/api/internal/service"
)

// --- Structs del contrato ---

// Rol agrupa permisos de este y de otros servicios.
type Rol struct {
	ID            string    `json:"id" format:"uuid" example:"11111111-1111-4111-8111-111111111111"`
	Nombre        string    `json:"nombre" example:"administrador"`
	Descripcion   string    `json:"descripcion" example:"Administra usuarios, roles y perfiles"`
	Permisos      []string  `json:"permisos" example:"[\"perfiles:operar\",\"roles:administrar\",\"usuarios:administrar\"]"`
	CreadoEn      time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
}

func aRol(r service.Rol) Rol {
	return Rol{
		ID: r.ID.String(), Nombre: r.Nombre, Descripcion: r.Descripcion, Permisos: r.Permisos,
		CreadoEn: r.CreadoEn, ActualizadoEn: r.ActualizadoEn,
	}
}

type salidaRol struct {
	Body Rol
}

// ListaRoles son todos los roles.
type ListaRoles struct {
	Datos []Rol `json:"datos"`
}

type salidaListaRoles struct {
	Body ListaRoles
}

// SolicitudCrearRol son los datos de un rol nuevo.
type SolicitudCrearRol struct {
	Nombre      string          `json:"nombre" pattern:"^[a-z][a-z0-9_]{2,39}$" patternDescription:"minúsculas, números y _; entre 3 y 40 caracteres" example:"recepcionista"`
	Descripcion string          `json:"descripcion,omitempty" maxLength:"200" example:"Atiende la recepción"`
	Permisos    []CodigoPermiso `json:"permisos" maxItems:"50" uniqueItems:"true" doc:"Permisos de este o de otros servicios"`
}

type entradaCrearRol struct {
	Body SolicitudCrearRol
}

// CambiosRol son los campos que se pueden editar; se envía al menos uno.
type CambiosRol struct {
	Descripcion *string          `json:"descripcion,omitempty" maxLength:"200"`
	Permisos    *[]CodigoPermiso `json:"permisos,omitempty" maxItems:"50" uniqueItems:"true"`
}

// TransformSchema exige al menos un campo en el contrato.
func (CambiosRol) TransformSchema(_ huma.Registry, s *huma.Schema) *huma.Schema {
	return conMinimoUnCampo(s)
}

type entradaActualizarRol struct {
	EntradaID
	Body CambiosRol
}

// Permiso es una entrada del catálogo.
type Permiso struct {
	Codigo      string `json:"codigo" example:"usuarios:administrar"`
	Descripcion string `json:"descripcion" example:"Crear, editar, desactivar y listar usuarios"`
}

// ListaPermisos es el catálogo de permisos propios del servicio.
type ListaPermisos struct {
	Datos []Permiso `json:"datos"`
}

type salidaPermisos struct {
	Body ListaPermisos
}

// --- Rutas ---

func registrarRoles(api huma.API, s ServicioRoles) {
	huma.Register(api, requiere(huma.Operation{
		OperationID: "listarRoles",
		Method:      http.MethodGet,
		Path:        "/api/roles",
		Summary:     "Lista todos los roles",
		Tags:        []string{etiquetaRoles},
	}, service.PermisoRolesAdministrar), func(ctx context.Context, _ *struct{}) (*salidaListaRoles, error) {
		roles, err := s.Listar(ctx)
		if err != nil {
			return nil, traducirError(err)
		}
		salida := &salidaListaRoles{}
		salida.Body.Datos = make([]Rol, 0, len(roles))
		for _, r := range roles {
			salida.Body.Datos = append(salida.Body.Datos, aRol(r))
		}
		return salida, nil
	})

	huma.Register(api, requiere(huma.Operation{
		OperationID:   "crearRol",
		Method:        http.MethodPost,
		Path:          "/api/roles",
		Summary:       "Crea un rol",
		Description:   "Acepta cualquier permiso con formato `<recurso>:<accion>`, también de otros servicios.",
		Tags:          []string{etiquetaRoles},
		DefaultStatus: http.StatusCreated,
		Errors:        []int{http.StatusConflict},
	}, service.PermisoRolesAdministrar), func(ctx context.Context, in *entradaCrearRol) (*salidaRol, error) {
		r, err := s.Crear(ctx, service.NuevoRol{
			Nombre: in.Body.Nombre, Descripcion: in.Body.Descripcion, Permisos: aTextos(in.Body.Permisos),
		})
		if err != nil {
			return nil, traducirError(err)
		}
		return &salidaRol{Body: aRol(r)}, nil
	})

	huma.Register(api, requiere(huma.Operation{
		OperationID: "obtenerRol",
		Method:      http.MethodGet,
		Path:        "/api/roles/{id}",
		Summary:     "Obtiene un rol",
		Tags:        []string{etiquetaRoles},
		Errors:      []int{http.StatusNotFound},
	}, service.PermisoRolesAdministrar), func(ctx context.Context, in *EntradaID) (*salidaRol, error) {
		r, err := s.Obtener(ctx, in.uuid())
		if err != nil {
			return nil, traducirError(err)
		}
		return &salidaRol{Body: aRol(r)}, nil
	})

	huma.Register(api, requiere(huma.Operation{
		OperationID: "actualizarRol",
		Method:      http.MethodPatch,
		Path:        "/api/roles/{id}",
		Summary:     "Cambia la descripción o los permisos de un rol",
		Description: "El rol administrador está protegido (409).",
		Tags:        []string{etiquetaRoles},
		Errors:      []int{http.StatusNotFound, http.StatusConflict},
	}, service.PermisoRolesAdministrar), func(ctx context.Context, in *entradaActualizarRol) (*salidaRol, error) {
		cambios := service.CambiosRol{Descripcion: in.Body.Descripcion}
		if in.Body.Permisos != nil {
			permisos := aTextos(*in.Body.Permisos)
			cambios.Permisos = &permisos
		}
		r, err := s.Actualizar(ctx, in.uuid(), cambios)
		if err != nil {
			return nil, traducirError(err)
		}
		return &salidaRol{Body: aRol(r)}, nil
	})

	huma.Register(api, requiere(huma.Operation{
		OperationID:   "eliminarRol",
		Method:        http.MethodDelete,
		Path:          "/api/roles/{id}",
		Summary:       "Elimina un rol que no esté asignado",
		Description:   "Si algún usuario tiene el rol, o es el rol administrador, responde 409.",
		Tags:          []string{etiquetaRoles},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusNotFound, http.StatusConflict},
	}, service.PermisoRolesAdministrar), func(ctx context.Context, in *EntradaID) (*struct{}, error) {
		if err := s.Eliminar(ctx, in.uuid()); err != nil {
			return nil, traducirError(err)
		}
		return nil, nil
	})

	huma.Register(api, autenticada(huma.Operation{
		OperationID: "listarPermisos",
		Method:      http.MethodGet,
		Path:        "/api/permisos",
		Summary:     "Lista los permisos propios del servicio de identidad",
		Tags:        []string{etiquetaRoles},
	}), func(ctx context.Context, _ *struct{}) (*salidaPermisos, error) {
		salida := &salidaPermisos{}
		salida.Body.Datos = []Permiso{}
		for _, p := range s.Permisos(ctx) {
			salida.Body.Datos = append(salida.Body.Datos, Permiso{Codigo: p.Codigo, Descripcion: p.Descripcion})
		}
		return salida, nil
	})
}
