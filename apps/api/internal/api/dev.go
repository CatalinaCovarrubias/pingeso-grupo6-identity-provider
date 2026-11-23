package api

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/CatalinaCovarrubias/pingeso-grupo6-identity-provider/apps/api/internal/service"
)

// SolicitudTokenPrueba pide un token de desarrollo.
type SolicitudTokenPrueba struct {
	Roles []string `json:"roles,omitempty" maxItems:"10" uniqueItems:"true" example:"[\"operador\"]" doc:"Nombres de roles existentes; sus permisos entran al token"`
	// PermisosExtra permite a otros grupos probar sus propios permisos.
	PermisosExtra []CodigoPermiso `json:"permisos_extra,omitempty" maxItems:"50" uniqueItems:"true" doc:"Permisos adicionales, por ejemplo de otro servicio"`
	UsuarioID     *string         `json:"usuario_id,omitempty" format:"uuid" example:"bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb" doc:"Valor de sub. Si falta, se inventa un UUID"`
	// Duracion admite minutos (1m a 1440m) u horas (1h a 24h). El patrón acota el máximo en el contrato.
	Duracion *string `json:"duracion,omitempty" pattern:"^(([1-9]|1[0-9]|2[0-4])h|([1-9][0-9]{0,2}|1[0-3][0-9]{2}|14[0-3][0-9]|1440)m)$" patternDescription:"minutos (1m a 1440m) u horas (1h a 24h)" example:"2h" doc:"Duración del token. Por defecto 15m (JWT_TTL); máximo 24h"`
}

// ClaimsToken es el contenido del JWT emitido.
type ClaimsToken struct {
	Iss      string   `json:"iss" example:"pingeso-identidad"`
	Sub      string   `json:"sub" format:"uuid"`
	Aud      []string `json:"aud" example:"[\"pingeso\"]"`
	Exp      int64    `json:"exp" doc:"Vencimiento, segundos Unix"`
	Iat      int64    `json:"iat" doc:"Emisión, segundos Unix"`
	Nbf      int64    `json:"nbf" doc:"Válido desde, segundos Unix"`
	Jti      string   `json:"jti" format:"uuid"`
	Roles    []string `json:"roles" example:"[\"operador\"]"`
	Permisos []string `json:"permisos" example:"[\"perfiles:operar\",\"reservas:operar\"]"`
}

// TokenPrueba es el token de desarrollo y sus claims decodificados.
type TokenPrueba struct {
	Token    string      `json:"token" doc:"JWT RS256 para usar en Authorization: Bearer"`
	ExpiraEn time.Time   `json:"expira_en"`
	Claims   ClaimsToken `json:"claims"`
}

type entradaTokenPrueba struct {
	Body SolicitudTokenPrueba
}

type salidaTokenPrueba struct {
	Body TokenPrueba
}

// registrarDesarrollo agrega las rutas que solo existen con API_ENV=dev.
func registrarDesarrollo(api huma.API, s ServicioDesarrollo) {
	huma.Register(api, publica(huma.Operation{
		OperationID: "emitirTokenPrueba",
		Method:      http.MethodPost,
		Path:        "/api/dev/tokens",
		Summary:     "Emite un JWT de prueba (solo con API_ENV=dev)",
		Description: "Solo existe con `API_ENV=dev`; con `test` o `prod` la ruta no existe (404). " +
			"Emite un token firmado con la llave real del servicio, con los permisos de los roles pedidos " +
			"más `permisos_extra`. El token viene en el body para usarlo como `Authorization: Bearer`.",
		Tags:   []string{etiquetaDesarrollo},
		Errors: []int{http.StatusUnprocessableEntity},
	}), func(ctx context.Context, in *entradaTokenPrueba) (*salidaTokenPrueba, error) {
		sol := service.SolicitudTokenPrueba{Roles: in.Body.Roles, PermisosExtra: aTextos(in.Body.PermisosExtra)}
		if in.Body.UsuarioID != nil {
			id := aUUID(*in.Body.UsuarioID)
			sol.UsuarioID = &id
		}
		if in.Body.Duracion != nil {
			// El contrato ya validó el formato; un valor ilegible deja 0 y se usa JWT_TTL.
			sol.Duracion, _ = time.ParseDuration(*in.Body.Duracion)
		}
		emitido, err := s.EmitirTokenPrueba(ctx, sol)
		if err != nil {
			return nil, traducirError(err)
		}
		c := emitido.Claims
		return &salidaTokenPrueba{Body: TokenPrueba{
			Token:    emitido.Token,
			ExpiraEn: emitido.ExpiraEn,
			Claims: ClaimsToken{
				Iss: c.Issuer, Sub: c.Subject, Aud: c.Audience, Exp: c.ExpiresAt.Unix(), Iat: c.IssuedAt.Unix(),
				Nbf: c.NotBefore.Unix(), Jti: c.ID, Roles: c.Roles, Permisos: c.Permisos,
			},
		}}, nil
	})
}
