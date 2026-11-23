package service

import (
	"context"

	"github.com/google/uuid"
)

// Perfiles administra el perfil genérico de cada usuario.
type Perfiles struct {
	perfiles AlmacenPerfiles
	usuarios AlmacenUsuarios
	ahora    Reloj
}

// NuevoPerfiles arma el service.
func NuevoPerfiles(perfiles AlmacenPerfiles, usuarios AlmacenUsuarios, ahora Reloj) *Perfiles {
	return &Perfiles{perfiles: perfiles, usuarios: usuarios, ahora: ahora}
}

// Obtener devuelve el perfil de un usuario. Lo puede ver el propio usuario o quien
// tenga perfiles:operar.
func (s *Perfiles) Obtener(ctx context.Context, usuarioID uuid.UUID) (Perfil, error) {
	if err := s.autorizar(ctx, usuarioID); err != nil {
		return Perfil{}, err
	}
	if _, err := s.usuarios.ObtenerUsuario(ctx, usuarioID); err != nil {
		return Perfil{}, noEncontrado(err, mensajeUsuarioNoExiste)
	}
	p, err := s.perfiles.ObtenerPerfil(ctx, usuarioID)
	return p, noEncontrado(err, "El usuario todavía no tiene perfil")
}

// Guardar crea o reemplaza el perfil completo de un usuario.
func (s *Perfiles) Guardar(ctx context.Context, p Perfil) (Perfil, error) {
	if err := s.autorizar(ctx, p.UsuarioID); err != nil {
		return Perfil{}, err
	}
	if _, err := s.usuarios.ObtenerUsuario(ctx, p.UsuarioID); err != nil {
		return Perfil{}, noEncontrado(err, mensajeUsuarioNoExiste)
	}
	if p.Atributos == nil {
		p.Atributos = map[string]any{}
	}
	p.ActualizadoEn = s.ahora()
	if err := s.perfiles.GuardarPerfil(ctx, p); err != nil {
		return Perfil{}, err
	}
	return p, nil
}

func (s *Perfiles) autorizar(ctx context.Context, usuarioID uuid.UUID) error {
	sesion, err := sesionObligatoria(ctx)
	if err != nil {
		return err
	}
	if sesion.UsuarioID == usuarioID || sesion.Tiene(PermisoPerfilesOperar) {
		return nil
	}
	return nuevoError(ErrSinPermiso, "Solo puedes ver o editar tu propio perfil")
}
