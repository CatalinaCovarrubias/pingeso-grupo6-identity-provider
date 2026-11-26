package api

import (
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"github.com/PINGESO-CIVIL/usuarios-roles/apps/api/internal/service"
)

// Valores de x-permiso que no son un permiso <recurso>:<accion>.
const (
	permisoPublico     = "publico"
	permisoAutenticado = "autenticado"
)

// Claves de Operation.Metadata que lee el middleware.
const (
	metaPermiso       = "permiso"
	metaPermitePropio = "permitePropioUsuario"
)

// nombreCookie es la cookie que lleva el JWT en el navegador.
const nombreCookie = "sesion"

// publica marca una operación sin autenticación.
func publica(op huma.Operation) huma.Operation {
	return conPermiso(op, permisoPublico, false)
}

// autenticada exige un token válido, sin permiso específico.
func autenticada(op huma.Operation) huma.Operation {
	return conPermiso(op, permisoAutenticado, false)
}

// requiere exige un token válido con el permiso indicado.
func requiere(op huma.Operation, permiso string) huma.Operation {
	return conPermiso(op, permiso, false)
}

// requiereOPropio exige el permiso, salvo que el usuario actúe sobre sí mismo.
// El middleware solo exige el token; el service decide si es el propio usuario.
func requiereOPropio(op huma.Operation, permiso string) huma.Operation {
	return conPermiso(op, permiso, true)
}

// conPermiso deja el permiso visible en el contrato (security y x-permiso) y en
// Metadata, donde lo lee el middleware.
func conPermiso(op huma.Operation, permiso string, propio bool) huma.Operation {
	op.Metadata = map[string]any{metaPermiso: permiso, metaPermitePropio: propio}
	op.Extensions = map[string]any{"x-permiso": permiso}
	if propio {
		op.Extensions["x-permite-propio-usuario"] = true
	}
	if permiso == permisoPublico {
		return op
	}
	op.Security = []map[string][]string{{esquemaCookie: {}}, {esquemaBearer: {}}}
	op.Errors = append(op.Errors, http.StatusUnauthorized)
	if permiso != permisoAutenticado || propio {
		op.Errors = append(op.Errors, http.StatusForbidden)
	}
	return op
}

// middlewareAutenticacion lee el token del header Authorization (Bearer) o de la cookie,
// lo valida, deja la sesión en el contexto y aplica el permiso de la operación.
// Una operación sin permiso declarado se trata como "autenticado": falla cerrado.
func middlewareAutenticacion(api huma.API, sesiones ServicioSesiones) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		permiso, _ := ctx.Operation().Metadata[metaPermiso].(string)
		if permiso == permisoPublico {
			next(ctx)
			return
		}
		if permiso == "" {
			permiso = permisoAutenticado
		}

		token := extraerToken(ctx)
		if token == "" {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "Se requiere una sesión válida")
			return
		}
		sesion, err := sesiones.Autenticar(ctx.Context(), token)
		if err != nil {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "Se requiere una sesión válida")
			return
		}

		propio, _ := ctx.Operation().Metadata[metaPermitePropio].(bool)
		if permiso != permisoAutenticado && !propio && !sesion.Tiene(permiso) {
			_ = huma.WriteErr(api, ctx, http.StatusForbidden, "Falta el permiso "+permiso)
			return
		}
		next(huma.WithContext(ctx, service.ConSesion(ctx.Context(), sesion)))
	}
}

// extraerToken prefiere Authorization: Bearer (llamadas entre servicios) y si no
// viene usa la cookie de sesión (navegador).
func extraerToken(ctx huma.Context) string {
	if valor := ctx.Header("Authorization"); valor != "" {
		esquema, token, ok := strings.Cut(valor, " ")
		if ok && strings.EqualFold(esquema, "Bearer") {
			return strings.TrimSpace(token)
		}
		return ""
	}
	cookie, err := huma.ReadCookie(ctx, nombreCookie)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// cookieSesion arma la cookie con el token. Secure también en desarrollo: los
// navegadores y curl tratan http://localhost como origen seguro.
func cookieSesion(token string, maxAge int) http.Cookie {
	return http.Cookie{
		Name:     nombreCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}
