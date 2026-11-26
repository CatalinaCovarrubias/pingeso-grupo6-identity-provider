package memoria

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/PINGESO-CIVIL/usuarios-roles/apps/api/internal/service"
)

// Datos seed de desarrollo, documentados en docs/jwt-de-prueba.md. Los IDs son fijos
// para que los otros grupos puedan usarlos en sus propias pruebas.
var (
	IDRolAdministrador = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	IDRolOperador      = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	IDRolUsuario       = uuid.MustParse("33333333-3333-4333-8333-333333333333")

	IDUsuarioAdmin    = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	IDUsuarioOperador = uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	IDUsuarioUsuario  = uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc")
)

// UsuarioSeed describe una cuenta de desarrollo con su contraseña en claro.
// Las contraseñas son públicas a propósito: solo existen en dev y test.
type UsuarioSeed struct {
	ID         uuid.UUID
	Correo     string
	Contrasena string
	Rol        uuid.UUID
	Nombre     string
}

// UsuariosSeed son las cuentas de desarrollo.
var UsuariosSeed = []UsuarioSeed{
	{IDUsuarioAdmin, "admin@dev.local", "admin-dev-local", IDRolAdministrador, "Ada"},
	{IDUsuarioOperador, "operador@dev.local", "operador-dev-local", IDRolOperador, "Olga"},
	{IDUsuarioUsuario, "usuario@dev.local", "usuario-dev-local", IDRolUsuario, "Ulises"},
}

// fechaSeed es fija para que las respuestas de los datos seed sean deterministas.
var fechaSeed = time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)

// CargarSeed agrega los roles, usuarios y perfiles de desarrollo. costoHash es el
// costo de bcrypt (las pruebas usan bcrypt.MinCost para ir más rápido).
func CargarSeed(ctx context.Context, m *Memoria, costoHash int) error {
	roles := []service.Rol{
		{
			ID: IDRolAdministrador, Nombre: service.RolAdministrador,
			Descripcion: "Administra usuarios, roles y perfiles",
			Permisos: []string{
				service.PermisoPerfilesOperar, service.PermisoRolesAdministrar, service.PermisoUsuariosAdministrar,
			},
		},
		{
			ID: IDRolOperador, Nombre: "operador",
			Descripcion: "Opera los perfiles de los usuarios",
			Permisos:    []string{service.PermisoPerfilesOperar},
		},
		{
			ID: IDRolUsuario, Nombre: "usuario",
			Descripcion: "Usuario final: solo su propia sesión y su perfil",
			Permisos:    []string{},
		},
	}
	for _, r := range roles {
		r.CreadoEn, r.ActualizadoEn = fechaSeed, fechaSeed
		if err := m.CrearRol(ctx, r); err != nil {
			return fmt.Errorf("seed del rol %s: %w", r.Nombre, err)
		}
	}

	for _, s := range UsuariosSeed {
		hash, err := bcrypt.GenerateFromPassword([]byte(s.Contrasena), costoHash)
		if err != nil {
			return fmt.Errorf("seed de %s: %w", s.Correo, err)
		}
		u := service.Usuario{
			ID: s.ID, Correo: s.Correo, HashContrasena: hash, Estado: service.EstadoActivo,
			RolIDs: []uuid.UUID{s.Rol}, CreadoEn: fechaSeed, ActualizadoEn: fechaSeed,
		}
		if err := m.CrearUsuario(ctx, u); err != nil {
			return fmt.Errorf("seed de %s: %w", s.Correo, err)
		}
		perfil := service.Perfil{
			UsuarioID: s.ID, Nombre: s.Nombre, Apellido: "Desarrollo",
			Atributos: map[string]any{}, ActualizadoEn: fechaSeed,
		}
		if err := m.GuardarPerfil(ctx, perfil); err != nil {
			return fmt.Errorf("seed del perfil de %s: %w", s.Correo, err)
		}
	}
	return nil
}
