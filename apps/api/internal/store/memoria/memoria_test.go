package memoria

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/PINGESO-CIVIL/usuarios-roles/apps/api/internal/service"
)

func conSeed(t *testing.T) *Memoria {
	t.Helper()
	m := Nueva()
	if err := CargarSeed(context.Background(), m, bcrypt.MinCost); err != nil {
		t.Fatalf("no se pudo cargar el seed: %v", err)
	}
	return m
}

func TestSeedGuardaHashYNoLaContrasena(t *testing.T) {
	m := conSeed(t)
	for _, s := range UsuariosSeed {
		u, err := m.BuscarUsuarioPorCorreo(context.Background(), s.Correo)
		if err != nil {
			t.Fatalf("%s debe existir: %v", s.Correo, err)
		}
		if string(u.HashContrasena) == s.Contrasena {
			t.Fatalf("%s guarda la contraseña en claro", s.Correo)
		}
		if bcrypt.CompareHashAndPassword(u.HashContrasena, []byte(s.Contrasena)) != nil {
			t.Fatalf("el hash de %s no corresponde a su contraseña", s.Correo)
		}
	}
}

func TestCorreoUnicoSinDistinguirMayusculas(t *testing.T) {
	m := conSeed(t)
	ctx := context.Background()

	if _, err := m.BuscarUsuarioPorCorreo(ctx, "ADMIN@dev.LOCAL"); err != nil {
		t.Fatalf("la búsqueda no debe distinguir mayúsculas: %v", err)
	}
	err := m.CrearUsuario(ctx, service.Usuario{ID: uuid.New(), Correo: "Admin@Dev.Local"})
	if !errors.Is(err, service.ErrConflicto) {
		t.Fatalf("se esperaba ErrConflicto, llegó %v", err)
	}
	otro, _ := m.ObtenerUsuario(ctx, IDUsuarioOperador)
	otro.Correo = "ADMIN@dev.local"
	if err := m.ActualizarUsuario(ctx, otro); !errors.Is(err, service.ErrConflicto) {
		t.Fatalf("se esperaba ErrConflicto al actualizar, llegó %v", err)
	}
}

func TestCreacionConcurrenteConElMismoCorreo(t *testing.T) {
	m := Nueva()
	var exitos, conflictos int
	var mu sync.Mutex
	var wg sync.WaitGroup
	// La mitad usa mayúsculas: el correo debe ser único sin distinguirlas.
	correos := []string{"ana@x.cl", "ANA@X.CL"}
	for i := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := m.CrearUsuario(context.Background(), service.Usuario{ID: uuid.New(), Correo: correos[i%2]})
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				exitos++
			} else if errors.Is(err, service.ErrConflicto) {
				conflictos++
			}
		}()
	}
	wg.Wait()
	if exitos != 1 || conflictos != 19 {
		t.Fatalf("se esperaba 1 éxito y 19 conflictos, llegaron %d y %d", exitos, conflictos)
	}
}

func TestListarFiltraYPagina(t *testing.T) {
	m := conSeed(t)
	ctx := context.Background()

	pagina, total, err := m.ListarUsuarios(ctx, service.FiltroUsuarios{Pagina: 1, Limite: 2})
	if err != nil || total != 3 || len(pagina) != 2 {
		t.Fatalf("página 1 inesperada: %d de %d, %v", len(pagina), total, err)
	}
	pagina, _, _ = m.ListarUsuarios(ctx, service.FiltroUsuarios{Pagina: 2, Limite: 2})
	if len(pagina) != 1 {
		t.Fatalf("la página 2 debe tener 1 usuario, tiene %d", len(pagina))
	}
	pagina, _, _ = m.ListarUsuarios(ctx, service.FiltroUsuarios{Pagina: 9, Limite: 2})
	if len(pagina) != 0 {
		t.Fatal("una página fuera de rango debe venir vacía")
	}
	pagina, total, _ = m.ListarUsuarios(ctx, service.FiltroUsuarios{Pagina: 1, Limite: 10, Correo: "OPERA"})
	if total != 1 || pagina[0].ID != IDUsuarioOperador {
		t.Fatalf("filtro por correo inesperado: %+v", pagina)
	}
	pagina, total, _ = m.ListarUsuarios(ctx, service.FiltroUsuarios{Pagina: 1, Limite: 10, Estado: service.EstadoBloqueado})
	if total != 0 || len(pagina) != 0 {
		t.Fatal("no hay usuarios bloqueados en el seed")
	}
}

func TestDevuelveCopias(t *testing.T) {
	m := conSeed(t)
	ctx := context.Background()

	u, _ := m.ObtenerUsuario(ctx, IDUsuarioAdmin)
	u.RolIDs[0] = uuid.Nil
	otra, _ := m.ObtenerUsuario(ctx, IDUsuarioAdmin)
	if otra.RolIDs[0] != IDRolAdministrador {
		t.Fatal("modificar la copia alteró lo guardado")
	}

	r, _ := m.ObtenerRol(ctx, IDRolAdministrador)
	r.Permisos[0] = "x:y"
	otro, _ := m.ObtenerRol(ctx, IDRolAdministrador)
	if otro.Permisos[0] == "x:y" {
		t.Fatal("modificar la copia del rol alteró lo guardado")
	}

	telefono := "+56911111111"
	ahora := time.Now()
	_ = m.GuardarPerfil(ctx, service.Perfil{UsuarioID: IDUsuarioAdmin, Telefono: &telefono, FotoURL: &telefono, Atributos: map[string]any{"a": 1}})
	p, _ := m.ObtenerPerfil(ctx, IDUsuarioAdmin)
	*p.Telefono = "cambiado"
	p.Atributos["a"] = 2
	p2, _ := m.ObtenerPerfil(ctx, IDUsuarioAdmin)
	if *p2.Telefono != telefono || p2.Atributos["a"] != 1 {
		t.Fatal("modificar la copia del perfil alteró lo guardado")
	}

	u.UltimoAcceso = &ahora
	u.RolIDs = []uuid.UUID{IDRolAdministrador}
	_ = m.ActualizarUsuario(ctx, u)
	conAcceso, _ := m.ObtenerUsuario(ctx, IDUsuarioAdmin)
	*conAcceso.UltimoAcceso = time.Time{}
	final, _ := m.ObtenerUsuario(ctx, IDUsuarioAdmin)
	if final.UltimoAcceso.IsZero() {
		t.Fatal("modificar la copia del último acceso alteró lo guardado")
	}
}

func TestRoles(t *testing.T) {
	m := conSeed(t)
	ctx := context.Background()

	roles, _ := m.ListarRoles(ctx)
	if len(roles) != 3 || roles[0].Nombre != "administrador" {
		t.Fatalf("roles inesperados: %+v", roles)
	}
	if _, err := m.BuscarRolPorNombre(ctx, "operador"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.BuscarRolPorNombre(ctx, "no-existe"); !errors.Is(err, service.ErrNoEncontrado) {
		t.Fatal("se esperaba ErrNoEncontrado")
	}
	if err := m.CrearRol(ctx, service.Rol{ID: uuid.New(), Nombre: "operador"}); !errors.Is(err, service.ErrConflicto) {
		t.Fatal("se esperaba ErrConflicto por nombre repetido")
	}
	operador, _ := m.ObtenerRol(ctx, IDRolOperador)
	operador.Nombre = "usuario"
	if err := m.ActualizarRol(ctx, operador); !errors.Is(err, service.ErrConflicto) {
		t.Fatal("se esperaba ErrConflicto al renombrar a un nombre ocupado")
	}
	if err := m.ActualizarRol(ctx, service.Rol{ID: uuid.New()}); !errors.Is(err, service.ErrNoEncontrado) {
		t.Fatal("se esperaba ErrNoEncontrado")
	}

	n, _ := m.ContarUsuariosConRol(ctx, IDRolAdministrador)
	activos, _ := m.ContarUsuariosActivosConRol(ctx, IDRolAdministrador)
	if n != 1 || activos != 1 {
		t.Fatalf("conteos inesperados: %d, %d", n, activos)
	}

	if err := m.EliminarRol(ctx, IDRolOperador); err != nil {
		t.Fatal(err)
	}
	if _, err := m.ObtenerRol(ctx, IDRolOperador); !errors.Is(err, service.ErrNoEncontrado) {
		t.Fatal("el rol eliminado no debe existir")
	}
	if err := m.EliminarRol(ctx, IDRolOperador); !errors.Is(err, service.ErrNoEncontrado) {
		t.Fatal("eliminar dos veces debe dar ErrNoEncontrado")
	}
}

func TestNoEncontrados(t *testing.T) {
	m := Nueva()
	ctx := context.Background()
	if _, err := m.ObtenerUsuario(ctx, uuid.New()); !errors.Is(err, service.ErrNoEncontrado) {
		t.Fatal("usuario")
	}
	if err := m.ActualizarUsuario(ctx, service.Usuario{ID: uuid.New()}); !errors.Is(err, service.ErrNoEncontrado) {
		t.Fatal("actualizar usuario")
	}
	if _, err := m.ObtenerPerfil(ctx, uuid.New()); !errors.Is(err, service.ErrNoEncontrado) {
		t.Fatal("perfil")
	}
	if _, err := m.BuscarUsuarioPorCorreo(ctx, "nadie@x.cl"); !errors.Is(err, service.ErrNoEncontrado) {
		t.Fatal("correo")
	}
}
