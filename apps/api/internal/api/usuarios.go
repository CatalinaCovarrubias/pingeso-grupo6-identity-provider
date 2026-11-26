package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/PINGESO-CIVIL/usuarios-roles/apps/api/internal/service"
)

// --- Structs del contrato ---

type entradaListarUsuarios struct {
	Pagina int    `query:"pagina" minimum:"1" maximum:"100000" default:"1" doc:"Número de página, desde 1"`
	Limite int    `query:"limite" minimum:"1" maximum:"100" default:"20" doc:"Usuarios por página, máximo 100"`
	Estado string `query:"estado" enum:"activo,bloqueado,inactivo" doc:"Filtra por estado"`
	Correo string `query:"correo" maxLength:"254" doc:"Filtra por correo (coincidencia parcial, sin distinguir mayúsculas)"`
}

// Paginacion describe la página devuelta.
type Paginacion struct {
	Pagina       int `json:"pagina" example:"1"`
	Limite       int `json:"limite" example:"20"`
	Total        int `json:"total" example:"57" doc:"Total de coincidencias"`
	TotalPaginas int `json:"total_paginas" example:"3"`
}

// ListaUsuarios es una página de usuarios.
type ListaUsuarios struct {
	Datos      []Usuario  `json:"datos"`
	Paginacion Paginacion `json:"paginacion"`
}

type salidaListaUsuarios struct {
	Body ListaUsuarios
}

type salidaUsuario struct {
	Body Usuario
}

// SolicitudCrearUsuario son los datos de una cuenta nueva.
type SolicitudCrearUsuario struct {
	Correo     string  `json:"correo" format:"email" maxLength:"254" example:"ana@ejemplo.cl"`
	Contrasena string  `json:"contrasena" minLength:"8" maxLength:"72" example:"una-clave-segura"`
	Roles      []IDRol `json:"roles,omitempty" maxItems:"20" uniqueItems:"true" doc:"IDs de los roles iniciales"`
}

type entradaCrearUsuario struct {
	Body SolicitudCrearUsuario
}

// CambiosUsuario son los campos que se pueden editar; se envía al menos uno.
type CambiosUsuario struct {
	Correo *string `json:"correo,omitempty" format:"email" maxLength:"254" example:"ana@ejemplo.cl"`
	Estado *string `json:"estado,omitempty" enum:"activo,bloqueado,inactivo"`
}

// TransformSchema exige al menos un campo en el contrato.
func (CambiosUsuario) TransformSchema(_ huma.Registry, s *huma.Schema) *huma.Schema {
	return conMinimoUnCampo(s)
}

type entradaActualizarUsuario struct {
	EntradaID
	Body CambiosUsuario
}

// RolesAsignados son los roles de un usuario.
type RolesAsignados struct {
	Roles []RolResumen `json:"roles"`
}

type salidaRoles struct {
	Body RolesAsignados
}

// AsignacionRoles reemplaza la lista completa de roles de un usuario.
type AsignacionRoles struct {
	Roles []IDRol `json:"roles" maxItems:"20" uniqueItems:"true" doc:"Lista completa de roles: reemplaza la anterior"`
}

type entradaAsignarRoles struct {
	EntradaID
	Body AsignacionRoles
}

// --- Rutas ---

func registrarUsuarios(api huma.API, s ServicioUsuarios) {
	huma.Register(api, requiere(huma.Operation{
		OperationID: "listarUsuarios",
		Method:      http.MethodGet,
		Path:        "/api/usuarios",
		Summary:     "Lista usuarios con paginación y filtros",
		Tags:        []string{etiquetaUsuarios},
	}, service.PermisoUsuariosAdministrar), func(ctx context.Context, in *entradaListarUsuarios) (*salidaListaUsuarios, error) {
		pagina, err := s.Listar(ctx, service.FiltroUsuarios{
			Pagina: in.Pagina, Limite: in.Limite, Estado: service.EstadoUsuario(in.Estado), Correo: in.Correo,
		})
		if err != nil {
			return nil, traducirError(err)
		}
		salida := &salidaListaUsuarios{}
		salida.Body.Datos = make([]Usuario, 0, len(pagina.Datos))
		for _, d := range pagina.Datos {
			salida.Body.Datos = append(salida.Body.Datos, aUsuario(d))
		}
		salida.Body.Paginacion = Paginacion{
			Pagina: in.Pagina, Limite: in.Limite, Total: pagina.Total,
			TotalPaginas: (pagina.Total + in.Limite - 1) / in.Limite,
		}
		return salida, nil
	})

	huma.Register(api, requiere(huma.Operation{
		OperationID:   "crearUsuario",
		Method:        http.MethodPost,
		Path:          "/api/usuarios",
		Summary:       "Crea un usuario activo",
		Description:   "El correo es único sin distinguir mayúsculas y se guarda en minúsculas.",
		Tags:          []string{etiquetaUsuarios},
		DefaultStatus: http.StatusCreated,
		Errors:        []int{http.StatusConflict},
	}, service.PermisoUsuariosAdministrar), func(ctx context.Context, in *entradaCrearUsuario) (*salidaUsuario, error) {
		u, err := s.Crear(ctx, service.NuevoUsuario{
			Correo: in.Body.Correo, Contrasena: in.Body.Contrasena, RolIDs: aUUIDs(in.Body.Roles),
		})
		if err != nil {
			return nil, traducirError(err)
		}
		return &salidaUsuario{Body: aUsuario(u)}, nil
	})

	huma.Register(api, requiere(huma.Operation{
		OperationID: "obtenerUsuario",
		Method:      http.MethodGet,
		Path:        "/api/usuarios/{id}",
		Summary:     "Obtiene un usuario",
		Tags:        []string{etiquetaUsuarios},
		Errors:      []int{http.StatusNotFound},
	}, service.PermisoUsuariosAdministrar), func(ctx context.Context, in *EntradaID) (*salidaUsuario, error) {
		u, err := s.Obtener(ctx, in.uuid())
		if err != nil {
			return nil, traducirError(err)
		}
		return &salidaUsuario{Body: aUsuario(u)}, nil
	})

	huma.Register(api, requiere(huma.Operation{
		OperationID: "actualizarUsuario",
		Method:      http.MethodPatch,
		Path:        "/api/usuarios/{id}",
		Summary:     "Cambia el correo o el estado de un usuario",
		Description: "No permite dejar al sistema sin administradores activos (409).",
		Tags:        []string{etiquetaUsuarios},
		Errors:      []int{http.StatusNotFound, http.StatusConflict},
	}, service.PermisoUsuariosAdministrar), func(ctx context.Context, in *entradaActualizarUsuario) (*salidaUsuario, error) {
		cambios := service.CambiosUsuario{Correo: in.Body.Correo}
		if in.Body.Estado != nil {
			estado := service.EstadoUsuario(*in.Body.Estado)
			cambios.Estado = &estado
		}
		u, err := s.Actualizar(ctx, in.uuid(), cambios)
		if err != nil {
			return nil, traducirError(err)
		}
		return &salidaUsuario{Body: aUsuario(u)}, nil
	})

	huma.Register(api, requiere(huma.Operation{
		OperationID:   "desactivarUsuario",
		Method:        http.MethodDelete,
		Path:          "/api/usuarios/{id}",
		Summary:       "Desactiva un usuario (estado inactivo)",
		Description:   "No borra datos. No permite desactivar al último administrador activo (409).",
		Tags:          []string{etiquetaUsuarios},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusNotFound, http.StatusConflict},
	}, service.PermisoUsuariosAdministrar), func(ctx context.Context, in *EntradaID) (*struct{}, error) {
		if err := s.Desactivar(ctx, in.uuid()); err != nil {
			return nil, traducirError(err)
		}
		return nil, nil
	})

	huma.Register(api, requiere(huma.Operation{
		OperationID: "obtenerRolesUsuario",
		Method:      http.MethodGet,
		Path:        "/api/usuarios/{id}/roles",
		Summary:     "Lista los roles asignados a un usuario",
		Tags:        []string{etiquetaUsuarios},
		Errors:      []int{http.StatusNotFound},
	}, service.PermisoRolesAdministrar), func(ctx context.Context, in *EntradaID) (*salidaRoles, error) {
		roles, err := s.Roles(ctx, in.uuid())
		if err != nil {
			return nil, traducirError(err)
		}
		salida := &salidaRoles{}
		salida.Body.Roles = aRolesResumen(roles)
		return salida, nil
	})

	huma.Register(api, requiere(huma.Operation{
		OperationID: "asignarRolesUsuario",
		Method:      http.MethodPut,
		Path:        "/api/usuarios/{id}/roles",
		Summary:     "Reemplaza los roles asignados a un usuario",
		Description: "Un rol inexistente da 422. No permite quitar el rol administrador al último administrador activo (409).",
		Tags:        []string{etiquetaUsuarios},
		Errors:      []int{http.StatusNotFound, http.StatusConflict},
	}, service.PermisoRolesAdministrar), func(ctx context.Context, in *entradaAsignarRoles) (*salidaRoles, error) {
		roles, err := s.AsignarRoles(ctx, in.uuid(), aUUIDs(in.Body.Roles))
		if err != nil {
			return nil, traducirError(err)
		}
		salida := &salidaRoles{}
		salida.Body.Roles = aRolesResumen(roles)
		return salida, nil
	})
}
