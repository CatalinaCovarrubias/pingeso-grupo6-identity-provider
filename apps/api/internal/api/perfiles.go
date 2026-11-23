package api

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/CatalinaCovarrubias/pingeso-grupo6-identity-provider/apps/api/internal/service"
)

// Perfil es la ficha genérica de un usuario.
type Perfil struct {
	UsuarioID     string         `json:"usuario_id" format:"uuid" example:"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"`
	Nombre        string         `json:"nombre" example:"Ana"`
	Apellido      string         `json:"apellido" example:"Pérez"`
	Telefono      *string        `json:"telefono,omitempty" example:"+56912345678"`
	FotoURL       *string        `json:"foto_url,omitempty" format:"uri" example:"https://boveda.example.cl/archivos/foto.jpg"`
	Atributos     map[string]any `json:"atributos" doc:"Campos libres que define cada cliente"`
	ActualizadoEn time.Time      `json:"actualizado_en"`
}

func aPerfil(p service.Perfil) Perfil {
	return Perfil{
		UsuarioID: p.UsuarioID.String(), Nombre: p.Nombre, Apellido: p.Apellido, Telefono: p.Telefono,
		FotoURL: p.FotoURL, Atributos: p.Atributos, ActualizadoEn: p.ActualizadoEn,
	}
}

type salidaPerfil struct {
	Body Perfil
}

// DatosPerfil es el perfil completo que se guarda.
type DatosPerfil struct {
	Nombre    string         `json:"nombre" minLength:"1" maxLength:"80" example:"Ana"`
	Apellido  string         `json:"apellido" minLength:"1" maxLength:"80" example:"Pérez"`
	Telefono  *string        `json:"telefono,omitempty" pattern:"^\\+?[0-9]{8,15}$" patternDescription:"entre 8 y 15 dígitos, con + opcional" example:"+56912345678"`
	FotoURL   *string        `json:"foto_url,omitempty" format:"uri" maxLength:"500" example:"https://boveda.example.cl/archivos/foto.jpg"`
	Atributos map[string]any `json:"atributos,omitempty" maxProperties:"50" doc:"Campos libres que define cada cliente, por ejemplo {\"estatura_cm\": 172}"`
}

type entradaGuardarPerfil struct {
	EntradaID
	Body DatosPerfil
}

func registrarPerfiles(api huma.API, s ServicioPerfiles) {
	huma.Register(api, requiereOPropio(huma.Operation{
		OperationID: "obtenerPerfil",
		Method:      http.MethodGet,
		Path:        "/api/usuarios/{id}/perfil",
		Summary:     "Obtiene el perfil de un usuario",
		Description: "Lo puede ver el propio usuario o quien tenga `perfiles:operar`. 404 si aún no tiene perfil.",
		Tags:        []string{etiquetaPerfiles},
		Errors:      []int{http.StatusNotFound},
	}, service.PermisoPerfilesOperar), func(ctx context.Context, in *EntradaID) (*salidaPerfil, error) {
		p, err := s.Obtener(ctx, in.uuid())
		if err != nil {
			return nil, traducirError(err)
		}
		return &salidaPerfil{Body: aPerfil(p)}, nil
	})

	huma.Register(api, requiereOPropio(huma.Operation{
		OperationID: "guardarPerfil",
		Method:      http.MethodPut,
		Path:        "/api/usuarios/{id}/perfil",
		Summary:     "Crea o reemplaza el perfil de un usuario",
		Description: "Lo puede editar el propio usuario o quien tenga `perfiles:operar`.",
		Tags:        []string{etiquetaPerfiles},
		Errors:      []int{http.StatusNotFound},
	}, service.PermisoPerfilesOperar), func(ctx context.Context, in *entradaGuardarPerfil) (*salidaPerfil, error) {
		p, err := s.Guardar(ctx, service.Perfil{
			UsuarioID: in.uuid(), Nombre: in.Body.Nombre, Apellido: in.Body.Apellido,
			Telefono: in.Body.Telefono, FotoURL: in.Body.FotoURL, Atributos: in.Body.Atributos,
		})
		if err != nil {
			return nil, traducirError(err)
		}
		return &salidaPerfil{Body: aPerfil(p)}, nil
	})
}
