// Paquete memoria implementa los puertos de almacenamiento del service en memoria.
// Es el store simulado del Hito 1: los datos se pierden al reiniciar. No tiene
// reglas de negocio, solo las restricciones que tendrá la base (unicidad y existencia).
package memoria

import (
	"cmp"
	"context"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/CatalinaCovarrubias/pingeso-grupo6-identity-provider/apps/api/internal/service"
)

// Memoria guarda usuarios, roles y perfiles. Es seguro usarla desde varias goroutines.
type Memoria struct {
	mu       sync.RWMutex
	usuarios map[uuid.UUID]service.Usuario
	roles    map[uuid.UUID]service.Rol
	perfiles map[uuid.UUID]service.Perfil
}

// Comprobación en compilación de que Memoria cumple los puertos.
var (
	_ service.AlmacenUsuarios = (*Memoria)(nil)
	_ service.AlmacenRoles    = (*Memoria)(nil)
	_ service.AlmacenPerfiles = (*Memoria)(nil)
)

// Nueva crea un almacén vacío.
func Nueva() *Memoria {
	return &Memoria{
		usuarios: map[uuid.UUID]service.Usuario{},
		roles:    map[uuid.UUID]service.Rol{},
		perfiles: map[uuid.UUID]service.Perfil{},
	}
}

// --- Usuarios ---

// ObtenerUsuario busca por ID.
func (m *Memoria) ObtenerUsuario(_ context.Context, id uuid.UUID) (service.Usuario, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.usuarios[id]
	if !ok {
		return service.Usuario{}, service.ErrNoEncontrado
	}
	return copiarUsuario(u), nil
}

// BuscarUsuarioPorCorreo busca sin distinguir mayúsculas.
func (m *Memoria) BuscarUsuarioPorCorreo(_ context.Context, correo string) (service.Usuario, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.usuarios {
		if strings.EqualFold(u.Correo, correo) {
			return copiarUsuario(u), nil
		}
	}
	return service.Usuario{}, service.ErrNoEncontrado
}

// ListarUsuarios filtra, ordena por creación (y correo para desempatar) y pagina.
func (m *Memoria) ListarUsuarios(_ context.Context, f service.FiltroUsuarios) ([]service.Usuario, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	correo := strings.ToLower(f.Correo)
	var coincidencias []service.Usuario
	for _, u := range m.usuarios {
		if f.Estado != "" && u.Estado != f.Estado {
			continue
		}
		if correo != "" && !strings.Contains(strings.ToLower(u.Correo), correo) {
			continue
		}
		coincidencias = append(coincidencias, u)
	}
	slices.SortFunc(coincidencias, func(a, b service.Usuario) int {
		return cmp.Or(a.CreadoEn.Compare(b.CreadoEn), strings.Compare(a.Correo, b.Correo))
	})

	total := len(coincidencias)
	inicio := min(max(f.Pagina-1, 0)*f.Limite, total)
	fin := min(inicio+f.Limite, total)
	pagina := make([]service.Usuario, 0, fin-inicio)
	for _, u := range coincidencias[inicio:fin] {
		pagina = append(pagina, copiarUsuario(u))
	}
	return pagina, total, nil
}

// CrearUsuario rechaza IDs o correos repetidos.
func (m *Memoria) CrearUsuario(_ context.Context, u service.Usuario) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, existe := m.usuarios[u.ID]; existe || m.correoOcupado(u.Correo, uuid.Nil) {
		return service.ErrConflicto
	}
	m.usuarios[u.ID] = copiarUsuario(u)
	return nil
}

// ActualizarUsuario reemplaza el usuario completo.
func (m *Memoria) ActualizarUsuario(_ context.Context, u service.Usuario) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, existe := m.usuarios[u.ID]; !existe {
		return service.ErrNoEncontrado
	}
	if m.correoOcupado(u.Correo, u.ID) {
		return service.ErrConflicto
	}
	m.usuarios[u.ID] = copiarUsuario(u)
	return nil
}

// ContarUsuariosConRol cuenta a todos los usuarios que tienen el rol, en cualquier estado.
func (m *Memoria) ContarUsuariosConRol(_ context.Context, rolID uuid.UUID) (int, error) {
	return m.contar(func(u service.Usuario) bool { return slices.Contains(u.RolIDs, rolID) }), nil
}

// ContarUsuariosActivosConRol cuenta a los usuarios activos que tienen el rol.
func (m *Memoria) ContarUsuariosActivosConRol(_ context.Context, rolID uuid.UUID) (int, error) {
	return m.contar(func(u service.Usuario) bool {
		return u.Estado == service.EstadoActivo && slices.Contains(u.RolIDs, rolID)
	}), nil
}

func (m *Memoria) contar(cumple func(service.Usuario) bool) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := 0
	for _, u := range m.usuarios {
		if cumple(u) {
			n++
		}
	}
	return n
}

// correoOcupado indica si otro usuario (distinto de excepto) ya usa el correo.
// Se llama con el candado tomado.
func (m *Memoria) correoOcupado(correo string, excepto uuid.UUID) bool {
	for id, u := range m.usuarios {
		if id != excepto && strings.EqualFold(u.Correo, correo) {
			return true
		}
	}
	return false
}

// --- Roles ---

// ListarRoles devuelve todos los roles ordenados por nombre.
func (m *Memoria) ListarRoles(_ context.Context) ([]service.Rol, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	roles := make([]service.Rol, 0, len(m.roles))
	for _, r := range m.roles {
		roles = append(roles, copiarRol(r))
	}
	slices.SortFunc(roles, func(a, b service.Rol) int { return strings.Compare(a.Nombre, b.Nombre) })
	return roles, nil
}

// ObtenerRol busca por ID.
func (m *Memoria) ObtenerRol(_ context.Context, id uuid.UUID) (service.Rol, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.roles[id]
	if !ok {
		return service.Rol{}, service.ErrNoEncontrado
	}
	return copiarRol(r), nil
}

// BuscarRolPorNombre busca por nombre exacto.
func (m *Memoria) BuscarRolPorNombre(_ context.Context, nombre string) (service.Rol, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, r := range m.roles {
		if r.Nombre == nombre {
			return copiarRol(r), nil
		}
	}
	return service.Rol{}, service.ErrNoEncontrado
}

// CrearRol rechaza IDs o nombres repetidos.
func (m *Memoria) CrearRol(_ context.Context, r service.Rol) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, existe := m.roles[r.ID]; existe || m.nombreOcupado(r.Nombre, uuid.Nil) {
		return service.ErrConflicto
	}
	m.roles[r.ID] = copiarRol(r)
	return nil
}

// ActualizarRol reemplaza el rol completo.
func (m *Memoria) ActualizarRol(_ context.Context, r service.Rol) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, existe := m.roles[r.ID]; !existe {
		return service.ErrNoEncontrado
	}
	if m.nombreOcupado(r.Nombre, r.ID) {
		return service.ErrConflicto
	}
	m.roles[r.ID] = copiarRol(r)
	return nil
}

// EliminarRol borra el rol.
func (m *Memoria) EliminarRol(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, existe := m.roles[id]; !existe {
		return service.ErrNoEncontrado
	}
	delete(m.roles, id)
	return nil
}

func (m *Memoria) nombreOcupado(nombre string, excepto uuid.UUID) bool {
	for id, r := range m.roles {
		if id != excepto && r.Nombre == nombre {
			return true
		}
	}
	return false
}

// --- Perfiles ---

// ObtenerPerfil busca el perfil del usuario.
func (m *Memoria) ObtenerPerfil(_ context.Context, usuarioID uuid.UUID) (service.Perfil, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.perfiles[usuarioID]
	if !ok {
		return service.Perfil{}, service.ErrNoEncontrado
	}
	return copiarPerfil(p), nil
}

// GuardarPerfil crea o reemplaza el perfil del usuario.
func (m *Memoria) GuardarPerfil(_ context.Context, p service.Perfil) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.perfiles[p.UsuarioID] = copiarPerfil(p)
	return nil
}

// --- Copias, para que nadie modifique lo guardado por referencia ---

func copiarUsuario(u service.Usuario) service.Usuario {
	u.HashContrasena = slices.Clone(u.HashContrasena)
	u.RolIDs = slices.Clone(u.RolIDs)
	if u.UltimoAcceso != nil {
		t := *u.UltimoAcceso
		u.UltimoAcceso = &t
	}
	return u
}

func copiarRol(r service.Rol) service.Rol {
	r.Permisos = slices.Clone(r.Permisos)
	return r
}

func copiarPerfil(p service.Perfil) service.Perfil {
	p.Atributos = maps.Clone(p.Atributos)
	if p.Telefono != nil {
		v := *p.Telefono
		p.Telefono = &v
	}
	if p.FotoURL != nil {
		v := *p.FotoURL
		p.FotoURL = &v
	}
	return p
}
