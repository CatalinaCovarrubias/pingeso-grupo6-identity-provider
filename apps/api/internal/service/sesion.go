package service

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
)

// Sesion es la identidad de quien hace la petición, sacada de un token válido.
type Sesion struct {
	UsuarioID uuid.UUID
	Roles     []string
	Permisos  []string
	ExpiraEn  time.Time
}

// Tiene indica si la sesión incluye el permiso.
func (s Sesion) Tiene(permiso string) bool {
	return slices.Contains(s.Permisos, permiso)
}

type claveSesion struct{}

// ConSesion guarda la sesión en el contexto. La usa el middleware de autenticación.
func ConSesion(ctx context.Context, s Sesion) context.Context {
	return context.WithValue(ctx, claveSesion{}, s)
}

// SesionDe recupera la sesión del contexto, si la hay.
func SesionDe(ctx context.Context) (Sesion, bool) {
	s, ok := ctx.Value(claveSesion{}).(Sesion)
	return s, ok
}

// sesionObligatoria devuelve la sesión o ErrNoAutenticado.
func sesionObligatoria(ctx context.Context) (Sesion, error) {
	s, ok := SesionDe(ctx)
	if !ok {
		return Sesion{}, nuevoError(ErrNoAutenticado, "Se requiere una sesión válida")
	}
	return s, nil
}
