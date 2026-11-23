package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/CatalinaCovarrubias/pingeso-grupo6-identity-provider/apps/api/internal/service/tokens"
)

// Sesiones maneja el inicio y cierre de sesión, la validación de tokens y los
// tokens de prueba de desarrollo.
type Sesiones struct {
	usuarios     AlmacenUsuarios
	roles        AlmacenRoles
	emisor       *tokens.Emisor
	ahora        Reloj
	hashFicticio []byte
}

// SesionIniciada es el resultado de un login o una renovación.
type SesionIniciada struct {
	Usuario UsuarioActual
	Token   string
}

// TokenEmitido es un token con sus claims, para renovaciones y tokens de prueba.
type TokenEmitido struct {
	Token    string
	ExpiraEn time.Time
	Claims   tokens.Claims
}

// NuevoSesiones arma el service. costoHash es el costo de bcrypt.
func NuevoSesiones(usuarios AlmacenUsuarios, roles AlmacenRoles, emisor *tokens.Emisor, costoHash int, ahora Reloj) (*Sesiones, error) {
	// Hash de una contraseña que nadie tiene: se compara cuando el correo no existe,
	// para que la respuesta tarde lo mismo y no revele qué correos están registrados (MF-12).
	ficticio, err := bcrypt.GenerateFromPassword([]byte(uuid.NewString()), costoHash)
	if err != nil {
		return nil, fmt.Errorf("no se pudo preparar el hash ficticio: %w", err)
	}
	return &Sesiones{usuarios: usuarios, roles: roles, emisor: emisor, ahora: ahora, hashFicticio: ficticio}, nil
}

// Iniciar valida las credenciales y emite un token. Un correo inexistente y una
// contraseña incorrecta dan el mismo error (MF-10).
func (s *Sesiones) Iniciar(ctx context.Context, correo, contrasena string) (SesionIniciada, error) {
	credencialesInvalidas := nuevoError(ErrCredenciales, "Correo o contraseña incorrectos")

	u, err := s.usuarios.BuscarUsuarioPorCorreo(ctx, normalizarCorreo(correo))
	if errors.Is(err, ErrNoEncontrado) {
		_ = bcrypt.CompareHashAndPassword(s.hashFicticio, []byte(contrasena))
		return SesionIniciada{}, credencialesInvalidas
	}
	if err != nil {
		return SesionIniciada{}, err
	}
	if bcrypt.CompareHashAndPassword(u.HashContrasena, []byte(contrasena)) != nil {
		return SesionIniciada{}, credencialesInvalidas
	}
	// Solo con la contraseña correcta se revela que la cuenta no está activa.
	if u.Estado != EstadoActivo {
		return SesionIniciada{}, nuevoError(ErrCuentaNoActiva, "La cuenta no está activa")
	}

	ahora := s.ahora()
	u.UltimoAcceso = &ahora
	if err := s.usuarios.ActualizarUsuario(ctx, u); err != nil {
		return SesionIniciada{}, err
	}

	d, roles, err := detalle(ctx, s.roles, u)
	if err != nil {
		return SesionIniciada{}, err
	}
	token, claims, err := s.emisor.Emitir(u.ID.String(), nombres(roles), permisosEfectivos(roles))
	if err != nil {
		return SesionIniciada{}, err
	}
	return SesionIniciada{
		Usuario: UsuarioActual{UsuarioDetalle: d, Permisos: claims.Permisos, ExpiraEn: claims.ExpiresAt.Time},
		Token:   token,
	}, nil
}

// Autenticar valida un token y devuelve la sesión que representa.
func (s *Sesiones) Autenticar(_ context.Context, token string) (Sesion, error) {
	noAutenticado := nuevoError(ErrNoAutenticado, "Se requiere una sesión válida")
	claims, err := s.emisor.Validar(token)
	if err != nil {
		return Sesion{}, noAutenticado
	}
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return Sesion{}, noAutenticado
	}
	return Sesion{UsuarioID: id, Roles: claims.Roles, Permisos: claims.Permisos, ExpiraEn: claims.ExpiresAt.Time}, nil
}

// Actual devuelve el usuario de la sesión, con los permisos de su token. Si el usuario
// ya no existe (por ejemplo, un token de prueba con un usuario inventado) la sesión no vale.
func (s *Sesiones) Actual(ctx context.Context) (UsuarioActual, error) {
	sesion, err := sesionObligatoria(ctx)
	if err != nil {
		return UsuarioActual{}, err
	}
	u, err := s.usuarios.ObtenerUsuario(ctx, sesion.UsuarioID)
	if errors.Is(err, ErrNoEncontrado) {
		return UsuarioActual{}, nuevoError(ErrNoAutenticado, "La sesión no corresponde a un usuario registrado")
	}
	if err != nil {
		return UsuarioActual{}, err
	}
	d, _, err := detalle(ctx, s.roles, u)
	if err != nil {
		return UsuarioActual{}, err
	}
	return UsuarioActual{UsuarioDetalle: d, Permisos: sesion.Permisos, ExpiraEn: sesion.ExpiraEn}, nil
}

// Renovar emite un token nuevo con los mismos roles y permisos. Es un mock: la
// renovación real (que vuelve a leer roles y respeta bloqueos) llega en la Tanda 4.
func (s *Sesiones) Renovar(ctx context.Context) (TokenEmitido, error) {
	sesion, err := sesionObligatoria(ctx)
	if err != nil {
		return TokenEmitido{}, err
	}
	return s.emitir(sesion.UsuarioID, sesion.Roles, sesion.Permisos)
}

// CambiarContrasena es un mock: valida el formato y no cambia nada (Tanda 4).
func (s *Sesiones) CambiarContrasena(ctx context.Context, actual, nueva string) error {
	if _, err := sesionObligatoria(ctx); err != nil {
		return err
	}
	if actual == "" {
		return nuevoError(ErrValidacion, "Falta la contraseña actual")
	}
	return validarContrasenaNueva(nueva)
}

// JWKS devuelve las llaves públicas para validar los tokens.
func (s *Sesiones) JWKS() []tokens.JWK {
	return s.emisor.JWKS()
}

// SolicitudTokenPrueba pide un token de desarrollo.
type SolicitudTokenPrueba struct {
	Roles         []string      // nombres de roles existentes
	PermisosExtra []string      // permisos de otros servicios, con formato <recurso>:<accion>
	UsuarioID     *uuid.UUID    // si falta, se inventa uno
	Duracion      time.Duration // si es cero, se usa JWT_TTL; máximo DuracionMaximaPrueba
}

// DuracionMaximaPrueba es lo más que puede durar un token de prueba.
const DuracionMaximaPrueba = 24 * time.Hour

// EmitirTokenPrueba emite un token con los roles pedidos más permisos extra.
// Solo se expone con API_ENV=dev.
func (s *Sesiones) EmitirTokenPrueba(ctx context.Context, sol SolicitudTokenPrueba) (TokenEmitido, error) {
	duracion := sol.Duracion
	if duracion == 0 {
		duracion = s.emisor.TTL()
	}
	if duracion < 0 || duracion > DuracionMaximaPrueba {
		return TokenEmitido{}, nuevoError(ErrValidacion, "La duración debe ser positiva y de 24 h como máximo")
	}

	var roles []Rol
	for _, nombre := range sol.Roles {
		r, err := s.roles.BuscarRolPorNombre(ctx, nombre)
		if errors.Is(err, ErrNoEncontrado) {
			return TokenEmitido{}, nuevoError(ErrValidacion, fmt.Sprintf("El rol %q no existe", nombre))
		}
		if err != nil {
			return TokenEmitido{}, err
		}
		roles = append(roles, r)
	}
	extra, err := normalizarPermisos(sol.PermisosExtra)
	if err != nil {
		return TokenEmitido{}, err
	}
	permisos := unicosOrdenados(append(permisosEfectivos(roles), extra...))

	id := uuid.New()
	if sol.UsuarioID != nil {
		id = *sol.UsuarioID
	}
	return s.emitirConDuracion(id, unicosOrdenados(nombres(roles)), permisos, duracion)
}

func (s *Sesiones) emitir(id uuid.UUID, roles, permisos []string) (TokenEmitido, error) {
	return s.emitirConDuracion(id, roles, permisos, s.emisor.TTL())
}

func (s *Sesiones) emitirConDuracion(id uuid.UUID, roles, permisos []string, duracion time.Duration) (TokenEmitido, error) {
	token, claims, err := s.emisor.EmitirConDuracion(id.String(), roles, permisos, duracion)
	if err != nil {
		return TokenEmitido{}, err
	}
	return TokenEmitido{Token: token, ExpiraEn: claims.ExpiresAt.Time, Claims: claims}, nil
}
