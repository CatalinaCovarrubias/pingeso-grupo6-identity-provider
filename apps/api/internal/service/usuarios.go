package service

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Usuarios administra las cuentas y sus roles.
type Usuarios struct {
	usuarios  AlmacenUsuarios
	roles     AlmacenRoles
	costoHash int
	ahora     Reloj
}

// NuevoUsuarios arma el service. costoHash es el costo de bcrypt.
func NuevoUsuarios(usuarios AlmacenUsuarios, roles AlmacenRoles, costoHash int, ahora Reloj) *Usuarios {
	return &Usuarios{usuarios: usuarios, roles: roles, costoHash: costoHash, ahora: ahora}
}

const mensajeUsuarioNoExiste = "El usuario no existe"

// Listar devuelve una página de usuarios.
func (s *Usuarios) Listar(ctx context.Context, f FiltroUsuarios) (PaginaUsuarios, error) {
	if f.Estado != "" && !f.Estado.Valido() {
		return PaginaUsuarios{}, nuevoError(ErrValidacion, "Estado desconocido")
	}
	lista, total, err := s.usuarios.ListarUsuarios(ctx, f)
	if err != nil {
		return PaginaUsuarios{}, err
	}
	datos := make([]UsuarioDetalle, 0, len(lista))
	for _, u := range lista {
		d, _, err := detalle(ctx, s.roles, u)
		if err != nil {
			return PaginaUsuarios{}, err
		}
		datos = append(datos, d)
	}
	return PaginaUsuarios{Datos: datos, Total: total}, nil
}

// Obtener devuelve un usuario.
func (s *Usuarios) Obtener(ctx context.Context, id uuid.UUID) (UsuarioDetalle, error) {
	u, err := s.usuarios.ObtenerUsuario(ctx, id)
	if err != nil {
		return UsuarioDetalle{}, noEncontrado(err, mensajeUsuarioNoExiste)
	}
	d, _, err := detalle(ctx, s.roles, u)
	return d, err
}

// NuevoUsuario son los datos para crear una cuenta.
type NuevoUsuario struct {
	Correo     string
	Contrasena string
	RolIDs     []uuid.UUID
}

// Crear registra una cuenta activa. El correo es único sin distinguir mayúsculas (MF-22).
func (s *Usuarios) Crear(ctx context.Context, n NuevoUsuario) (UsuarioDetalle, error) {
	if err := validarContrasenaNueva(n.Contrasena); err != nil {
		return UsuarioDetalle{}, err
	}
	rolIDs, err := verificarRolesExisten(ctx, s.roles, n.RolIDs)
	if err != nil {
		return UsuarioDetalle{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(n.Contrasena), s.costoHash)
	if err != nil {
		return UsuarioDetalle{}, fmt.Errorf("no se pudo calcular el hash: %w", err)
	}
	ahora := s.ahora()
	u := Usuario{
		ID: uuid.New(), Correo: normalizarCorreo(n.Correo), HashContrasena: hash,
		Estado: EstadoActivo, RolIDs: rolIDs, CreadoEn: ahora, ActualizadoEn: ahora,
	}
	if err := s.usuarios.CrearUsuario(ctx, u); err != nil {
		return UsuarioDetalle{}, conflictoDeCorreo(err)
	}
	d, _, err := detalle(ctx, s.roles, u)
	return d, err
}

// CambiosUsuario son los campos editables; los nil no se tocan.
type CambiosUsuario struct {
	Correo *string
	Estado *EstadoUsuario
}

// Actualizar cambia el correo o el estado.
func (s *Usuarios) Actualizar(ctx context.Context, id uuid.UUID, c CambiosUsuario) (UsuarioDetalle, error) {
	if c.Correo == nil && c.Estado == nil {
		return UsuarioDetalle{}, nuevoError(ErrValidacion, "Indica al menos un campo para cambiar")
	}
	if c.Estado != nil && !c.Estado.Valido() {
		return UsuarioDetalle{}, nuevoError(ErrValidacion, "Estado desconocido")
	}
	antes, err := s.usuarios.ObtenerUsuario(ctx, id)
	if err != nil {
		return UsuarioDetalle{}, noEncontrado(err, mensajeUsuarioNoExiste)
	}
	despues := antes
	if c.Correo != nil {
		despues.Correo = normalizarCorreo(*c.Correo)
	}
	if c.Estado != nil {
		despues.Estado = *c.Estado
	}
	if err := s.guardar(ctx, antes, &despues); err != nil {
		return UsuarioDetalle{}, err
	}
	d, _, err := detalle(ctx, s.roles, despues)
	return d, err
}

// Desactivar deja la cuenta como inactiva. No borra datos.
func (s *Usuarios) Desactivar(ctx context.Context, id uuid.UUID) error {
	antes, err := s.usuarios.ObtenerUsuario(ctx, id)
	if err != nil {
		return noEncontrado(err, mensajeUsuarioNoExiste)
	}
	despues := antes
	despues.Estado = EstadoInactivo
	return s.guardar(ctx, antes, &despues)
}

// Roles devuelve los roles asignados.
func (s *Usuarios) Roles(ctx context.Context, id uuid.UUID) ([]RolResumen, error) {
	d, err := s.Obtener(ctx, id)
	if err != nil {
		return nil, err
	}
	return d.Roles, nil
}

// AsignarRoles reemplaza la lista completa de roles del usuario.
func (s *Usuarios) AsignarRoles(ctx context.Context, id uuid.UUID, rolIDs []uuid.UUID) ([]RolResumen, error) {
	antes, err := s.usuarios.ObtenerUsuario(ctx, id)
	if err != nil {
		return nil, noEncontrado(err, mensajeUsuarioNoExiste)
	}
	ids, err := verificarRolesExisten(ctx, s.roles, rolIDs)
	if err != nil {
		return nil, err
	}
	despues := antes
	despues.RolIDs = ids
	if err := s.guardar(ctx, antes, &despues); err != nil {
		return nil, err
	}
	d, _, err := detalle(ctx, s.roles, despues)
	if err != nil {
		return nil, err
	}
	return d.Roles, nil
}

// guardar aplica la regla del último administrador (MF-24), fija ActualizadoEn en
// despues y persiste el cambio.
// En memoria la verificación y el guardado no son atómicos; la Tanda 4 lo resuelve
// con una transacción en PostgreSQL.
func (s *Usuarios) guardar(ctx context.Context, antes Usuario, despues *Usuario) error {
	if err := s.verificarQuedaAdministrador(ctx, antes, *despues); err != nil {
		return err
	}
	despues.ActualizadoEn = s.ahora()
	if err := s.usuarios.ActualizarUsuario(ctx, *despues); err != nil {
		return conflictoDeCorreo(noEncontrado(err, mensajeUsuarioNoExiste))
	}
	return nil
}

// verificarQuedaAdministrador impide que el cambio deje al sistema sin administradores activos.
func (s *Usuarios) verificarQuedaAdministrador(ctx context.Context, antes, despues Usuario) error {
	admin, err := s.roles.BuscarRolPorNombre(ctx, RolAdministrador)
	if errors.Is(err, ErrNoEncontrado) {
		return nil
	}
	if err != nil {
		return err
	}
	eraAdmin := antes.Estado == EstadoActivo && slices.Contains(antes.RolIDs, admin.ID)
	siguePorSerlo := despues.Estado == EstadoActivo && slices.Contains(despues.RolIDs, admin.ID)
	if !eraAdmin || siguePorSerlo {
		return nil
	}
	activos, err := s.usuarios.ContarUsuariosActivosConRol(ctx, admin.ID)
	if err != nil {
		return err
	}
	if activos <= 1 {
		return nuevoError(ErrConflicto, "Debe quedar al menos un administrador activo")
	}
	return nil
}

func conflictoDeCorreo(err error) error {
	var negocio *ErrorNegocio
	if errors.Is(err, ErrConflicto) && !errors.As(err, &negocio) {
		return nuevoError(ErrConflicto, "Ya existe un usuario con ese correo")
	}
	return err
}
