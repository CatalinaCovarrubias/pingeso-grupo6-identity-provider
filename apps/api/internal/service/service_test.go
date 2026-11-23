package service_test

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/CatalinaCovarrubias/pingeso-grupo6-identity-provider/apps/api/internal/service"
	"github.com/CatalinaCovarrubias/pingeso-grupo6-identity-provider/apps/api/internal/service/tokens"
	"github.com/CatalinaCovarrubias/pingeso-grupo6-identity-provider/apps/api/internal/store/memoria"
)

var (
	llave     *tokens.Llave
	unaLlave  sync.Once
	instante  = time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)
	relojFijo = func() time.Time { return instante }
)

// entorno reúne los services armados sobre un store en memoria con el seed.
type entorno struct {
	almacen  *memoria.Memoria
	sesiones *service.Sesiones
	usuarios *service.Usuarios
	roles    *service.Roles
	perfiles *service.Perfiles
}

func nuevoEntorno(t *testing.T) entorno {
	t.Helper()
	unaLlave.Do(func() {
		var err error
		if llave, err = tokens.GenerarLlave(); err != nil {
			t.Fatal(err)
		}
	})
	m := memoria.Nueva()
	if err := memoria.CargarSeed(context.Background(), m, bcrypt.MinCost); err != nil {
		t.Fatal(err)
	}
	emisor := tokens.NuevoEmisor(llave, "pingeso-identidad", "pingeso", 15*time.Minute).ConReloj(relojFijo)
	sesiones, err := service.NuevoSesiones(m, m, emisor, bcrypt.MinCost, relojFijo)
	if err != nil {
		t.Fatal(err)
	}
	return entorno{
		almacen:  m,
		sesiones: sesiones,
		usuarios: service.NuevoUsuarios(m, m, bcrypt.MinCost, relojFijo),
		roles:    service.NuevoRoles(m, m, relojFijo),
		perfiles: service.NuevoPerfiles(m, m, relojFijo),
	}
}

func como(id uuid.UUID, permisos ...string) context.Context {
	return service.ConSesion(context.Background(), service.Sesion{UsuarioID: id, Permisos: permisos})
}

func esperar(t *testing.T, err, causa error) {
	t.Helper()
	if !errors.Is(err, causa) {
		t.Fatalf("se esperaba %v, llegó %v", causa, err)
	}
}

// --- Sesiones ---

func TestIniciarSesionCorrecta(t *testing.T) {
	e := nuevoEntorno(t)
	s, err := e.sesiones.Iniciar(context.Background(), "  ADMIN@dev.local ", "admin-dev-local")
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if s.Token == "" || s.Usuario.ID != memoria.IDUsuarioAdmin {
		t.Fatalf("sesión inesperada: %+v", s.Usuario)
	}
	esperados := []string{"perfiles:operar", "roles:administrar", "usuarios:administrar"}
	if !slices.Equal(s.Usuario.Permisos, esperados) {
		t.Fatalf("permisos inesperados: %v", s.Usuario.Permisos)
	}
	if !s.Usuario.ExpiraEn.Equal(instante.Add(15 * time.Minute)) {
		t.Fatalf("expiración inesperada: %v", s.Usuario.ExpiraEn)
	}
	u, _ := e.almacen.ObtenerUsuario(context.Background(), memoria.IDUsuarioAdmin)
	if u.UltimoAcceso == nil || !u.UltimoAcceso.Equal(instante) {
		t.Fatal("el login debe registrar el último acceso")
	}
}

func TestIniciarSesionIncorrectaNoRevelaElCorreo(t *testing.T) {
	e := nuevoEntorno(t)
	_, errClave := e.sesiones.Iniciar(context.Background(), "admin@dev.local", "otra-clave")
	_, errCorreo := e.sesiones.Iniciar(context.Background(), "nadie@dev.local", "otra-clave")
	esperar(t, errClave, service.ErrCredenciales)
	esperar(t, errCorreo, service.ErrCredenciales)
	if errClave.Error() != errCorreo.Error() {
		t.Fatalf("los mensajes deben ser iguales: %q y %q", errClave, errCorreo)
	}
}

func TestIniciarSesionConCuentaNoActiva(t *testing.T) {
	e := nuevoEntorno(t)
	bloqueado := service.EstadoBloqueado
	if _, err := e.usuarios.Actualizar(context.Background(), memoria.IDUsuarioOperador, service.CambiosUsuario{Estado: &bloqueado}); err != nil {
		t.Fatal(err)
	}
	_, err := e.sesiones.Iniciar(context.Background(), "operador@dev.local", "operador-dev-local")
	esperar(t, err, service.ErrCuentaNoActiva)
	// Con la contraseña mala no se revela que está bloqueada.
	_, err = e.sesiones.Iniciar(context.Background(), "operador@dev.local", "mala")
	esperar(t, err, service.ErrCredenciales)
}

func TestAutenticar(t *testing.T) {
	e := nuevoEntorno(t)
	s, _ := e.sesiones.Iniciar(context.Background(), "operador@dev.local", "operador-dev-local")

	sesion, err := e.sesiones.Autenticar(context.Background(), s.Token)
	if err != nil {
		t.Fatal(err)
	}
	if sesion.UsuarioID != memoria.IDUsuarioOperador || !sesion.Tiene("perfiles:operar") || sesion.Tiene("roles:administrar") {
		t.Fatalf("sesión inesperada: %+v", sesion)
	}
	_, err = e.sesiones.Autenticar(context.Background(), "basura")
	esperar(t, err, service.ErrNoAutenticado)
}

func TestAutenticarRechazaSubQueNoEsUUID(t *testing.T) {
	e := nuevoEntorno(t)
	emisor := tokens.NuevoEmisor(llave, "pingeso-identidad", "pingeso", 15*time.Minute).ConReloj(relojFijo)
	token, _, _ := emisor.Emitir("no-es-uuid", nil, nil)
	_, err := e.sesiones.Autenticar(context.Background(), token)
	esperar(t, err, service.ErrNoAutenticado)
}

func TestSesionActual(t *testing.T) {
	e := nuevoEntorno(t)
	_, err := e.sesiones.Actual(context.Background())
	esperar(t, err, service.ErrNoAutenticado)

	_, err = e.sesiones.Actual(como(uuid.New()))
	esperar(t, err, service.ErrNoAutenticado)

	actual, err := e.sesiones.Actual(como(memoria.IDUsuarioUsuario))
	if err != nil {
		t.Fatal(err)
	}
	if actual.Correo != "usuario@dev.local" || actual.Roles[0].Nombre != "usuario" {
		t.Fatalf("usuario actual inesperado: %+v", actual)
	}
}

func TestRenovarEmiteOtroTokenConLosMismosPermisos(t *testing.T) {
	e := nuevoEntorno(t)
	s, _ := e.sesiones.Iniciar(context.Background(), "admin@dev.local", "admin-dev-local")
	sesion, _ := e.sesiones.Autenticar(context.Background(), s.Token)

	renovado, err := e.sesiones.Renovar(service.ConSesion(context.Background(), sesion))
	if err != nil {
		t.Fatal(err)
	}
	if renovado.Token == s.Token || !slices.Equal(renovado.Claims.Permisos, sesion.Permisos) {
		t.Fatal("la renovación debe emitir un token nuevo con los mismos permisos")
	}
	_, err = e.sesiones.Renovar(context.Background())
	esperar(t, err, service.ErrNoAutenticado)
}

func TestCambiarContrasenaMock(t *testing.T) {
	e := nuevoEntorno(t)
	ctx := como(memoria.IDUsuarioUsuario)
	if err := e.sesiones.CambiarContrasena(ctx, "usuario-dev-local", "nueva-clave-segura"); err != nil {
		t.Fatal(err)
	}
	esperar(t, e.sesiones.CambiarContrasena(ctx, "x", "corta"), service.ErrValidacion)
	esperar(t, e.sesiones.CambiarContrasena(ctx, "", "nueva-clave-segura"), service.ErrValidacion)
	esperar(t, e.sesiones.CambiarContrasena(context.Background(), "x", "nueva-clave-segura"), service.ErrNoAutenticado)
}

func TestEmitirTokenPrueba(t *testing.T) {
	e := nuevoEntorno(t)
	id := uuid.New()
	tok, err := e.sesiones.EmitirTokenPrueba(context.Background(), service.SolicitudTokenPrueba{
		Roles:         []string{"operador", "operador"},
		PermisosExtra: []string{"reservas:operar", "perfiles:operar"},
		UsuarioID:     &id,
	})
	if err != nil {
		t.Fatal(err)
	}
	if tok.Claims.Subject != id.String() || !slices.Equal(tok.Claims.Roles, []string{"operador"}) {
		t.Fatalf("claims inesperados: %+v", tok.Claims)
	}
	if !slices.Equal(tok.Claims.Permisos, []string{"perfiles:operar", "reservas:operar"}) {
		t.Fatalf("permisos inesperados: %v", tok.Claims.Permisos)
	}

	sinUsuario, err := e.sesiones.EmitirTokenPrueba(context.Background(), service.SolicitudTokenPrueba{})
	if err != nil || uuid.Validate(sinUsuario.Claims.Subject) != nil {
		t.Fatalf("sin usuario_id debe inventar un UUID: %v", err)
	}

	_, err = e.sesiones.EmitirTokenPrueba(context.Background(), service.SolicitudTokenPrueba{Roles: []string{"superusuario"}})
	esperar(t, err, service.ErrValidacion)
	_, err = e.sesiones.EmitirTokenPrueba(context.Background(), service.SolicitudTokenPrueba{PermisosExtra: []string{"Reservas:Todo"}})
	esperar(t, err, service.ErrValidacion)
}

func TestJWKSDelService(t *testing.T) {
	e := nuevoEntorno(t)
	if llaves := e.sesiones.JWKS(); len(llaves) != 1 || llaves[0].Kid != llave.Kid() {
		t.Fatal("el JWKS debe publicar el kid de la llave de firma")
	}
}

// --- Usuarios ---

func TestCrearUsuario(t *testing.T) {
	e := nuevoEntorno(t)
	ctx := context.Background()

	u, err := e.usuarios.Crear(ctx, service.NuevoUsuario{
		Correo: " Ana@Ejemplo.CL", Contrasena: "clave-segura", RolIDs: []uuid.UUID{memoria.IDRolUsuario, memoria.IDRolUsuario},
	})
	if err != nil {
		t.Fatal(err)
	}
	if u.Correo != "ana@ejemplo.cl" || u.Estado != service.EstadoActivo || len(u.Roles) != 1 {
		t.Fatalf("usuario inesperado: %+v", u)
	}
	guardado, _ := e.almacen.ObtenerUsuario(ctx, u.ID)
	if bcrypt.CompareHashAndPassword(guardado.HashContrasena, []byte("clave-segura")) != nil {
		t.Fatal("se debe guardar el hash de la contraseña")
	}

	_, err = e.usuarios.Crear(ctx, service.NuevoUsuario{Correo: "ANA@ejemplo.cl", Contrasena: "clave-segura"})
	esperar(t, err, service.ErrConflicto)
	_, err = e.usuarios.Crear(ctx, service.NuevoUsuario{Correo: "b@x.cl", Contrasena: "corta"})
	esperar(t, err, service.ErrValidacion)
	_, err = e.usuarios.Crear(ctx, service.NuevoUsuario{Correo: "b@x.cl", Contrasena: "clave-segura", RolIDs: []uuid.UUID{uuid.New()}})
	esperar(t, err, service.ErrValidacion)
}

func TestListarYObtenerUsuarios(t *testing.T) {
	e := nuevoEntorno(t)
	ctx := context.Background()

	pagina, err := e.usuarios.Listar(ctx, service.FiltroUsuarios{Pagina: 1, Limite: 2})
	if err != nil || pagina.Total != 3 || len(pagina.Datos) != 2 {
		t.Fatalf("página inesperada: %+v, %v", pagina, err)
	}
	_, err = e.usuarios.Listar(ctx, service.FiltroUsuarios{Pagina: 1, Limite: 2, Estado: "raro"})
	esperar(t, err, service.ErrValidacion)

	_, err = e.usuarios.Obtener(ctx, uuid.New())
	esperar(t, err, service.ErrNoEncontrado)
	roles, err := e.usuarios.Roles(ctx, memoria.IDUsuarioOperador)
	if err != nil || roles[0].ID != memoria.IDRolOperador {
		t.Fatalf("roles inesperados: %+v, %v", roles, err)
	}
}

func TestActualizarUsuario(t *testing.T) {
	e := nuevoEntorno(t)
	ctx := context.Background()

	_, err := e.usuarios.Actualizar(ctx, memoria.IDUsuarioUsuario, service.CambiosUsuario{})
	esperar(t, err, service.ErrValidacion)
	raro := service.EstadoUsuario("raro")
	_, err = e.usuarios.Actualizar(ctx, memoria.IDUsuarioUsuario, service.CambiosUsuario{Estado: &raro})
	esperar(t, err, service.ErrValidacion)
	ocupado := "Operador@dev.local"
	_, err = e.usuarios.Actualizar(ctx, memoria.IDUsuarioUsuario, service.CambiosUsuario{Correo: &ocupado})
	esperar(t, err, service.ErrConflicto)
	_, err = e.usuarios.Actualizar(ctx, uuid.New(), service.CambiosUsuario{Correo: &ocupado})
	esperar(t, err, service.ErrNoEncontrado)

	nuevo := "Ulises@Dev.Local"
	u, err := e.usuarios.Actualizar(ctx, memoria.IDUsuarioUsuario, service.CambiosUsuario{Correo: &nuevo})
	if err != nil || u.Correo != "ulises@dev.local" || !u.ActualizadoEn.Equal(instante) {
		t.Fatalf("actualización inesperada: %+v, %v", u, err)
	}
}

func TestDesactivarUsuario(t *testing.T) {
	e := nuevoEntorno(t)
	ctx := context.Background()

	if err := e.usuarios.Desactivar(ctx, memoria.IDUsuarioUsuario); err != nil {
		t.Fatal(err)
	}
	u, _ := e.usuarios.Obtener(ctx, memoria.IDUsuarioUsuario)
	if u.Estado != service.EstadoInactivo {
		t.Fatalf("estado inesperado: %s", u.Estado)
	}
	esperar(t, e.usuarios.Desactivar(ctx, uuid.New()), service.ErrNoEncontrado)
}

func TestSiempreQuedaUnAdministradorActivo(t *testing.T) {
	e := nuevoEntorno(t)
	ctx := context.Background()

	esperar(t, e.usuarios.Desactivar(ctx, memoria.IDUsuarioAdmin), service.ErrConflicto)
	bloqueado := service.EstadoBloqueado
	_, err := e.usuarios.Actualizar(ctx, memoria.IDUsuarioAdmin, service.CambiosUsuario{Estado: &bloqueado})
	esperar(t, err, service.ErrConflicto)
	_, err = e.usuarios.AsignarRoles(ctx, memoria.IDUsuarioAdmin, []uuid.UUID{memoria.IDRolUsuario})
	esperar(t, err, service.ErrConflicto)

	// Con un segundo administrador activo, el primero sí se puede desactivar.
	if _, err := e.usuarios.AsignarRoles(ctx, memoria.IDUsuarioOperador, []uuid.UUID{memoria.IDRolAdministrador}); err != nil {
		t.Fatal(err)
	}
	if err := e.usuarios.Desactivar(ctx, memoria.IDUsuarioAdmin); err != nil {
		t.Fatalf("con otro administrador se debe poder desactivar: %v", err)
	}
	// Y ahora el segundo es el último.
	esperar(t, e.usuarios.Desactivar(ctx, memoria.IDUsuarioOperador), service.ErrConflicto)
}

func TestAsignarRoles(t *testing.T) {
	e := nuevoEntorno(t)
	ctx := context.Background()

	roles, err := e.usuarios.AsignarRoles(ctx, memoria.IDUsuarioUsuario, []uuid.UUID{memoria.IDRolUsuario, memoria.IDRolOperador})
	if err != nil || len(roles) != 2 || roles[0].Nombre != "operador" {
		t.Fatalf("roles inesperados: %+v, %v", roles, err)
	}
	_, err = e.usuarios.AsignarRoles(ctx, memoria.IDUsuarioUsuario, []uuid.UUID{uuid.New()})
	esperar(t, err, service.ErrValidacion)
	_, err = e.usuarios.AsignarRoles(ctx, uuid.New(), nil)
	esperar(t, err, service.ErrNoEncontrado)

	vacios, err := e.usuarios.AsignarRoles(ctx, memoria.IDUsuarioUsuario, nil)
	if err != nil || len(vacios) != 0 {
		t.Fatalf("se debe poder quitar todos los roles: %+v, %v", vacios, err)
	}
}

// --- Roles ---

func TestCrudDeRoles(t *testing.T) {
	e := nuevoEntorno(t)
	ctx := context.Background()

	r, err := e.roles.Crear(ctx, service.NuevoRol{Nombre: "recepcionista", Permisos: []string{"reservas:operar", "perfiles:operar", "reservas:operar"}})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(r.Permisos, []string{"perfiles:operar", "reservas:operar"}) {
		t.Fatalf("permisos sin normalizar: %v", r.Permisos)
	}
	_, err = e.roles.Crear(ctx, service.NuevoRol{Nombre: "recepcionista"})
	esperar(t, err, service.ErrConflicto)
	_, err = e.roles.Crear(ctx, service.NuevoRol{Nombre: "otro", Permisos: []string{"sin-accion"}})
	esperar(t, err, service.ErrValidacion)

	descripcion := "Atiende la recepción"
	actualizado, err := e.roles.Actualizar(ctx, r.ID, service.CambiosRol{Descripcion: &descripcion, Permisos: &[]string{"reservas:operar"}})
	if err != nil || actualizado.Descripcion != descripcion || len(actualizado.Permisos) != 1 {
		t.Fatalf("actualización inesperada: %+v, %v", actualizado, err)
	}
	_, err = e.roles.Actualizar(ctx, r.ID, service.CambiosRol{})
	esperar(t, err, service.ErrValidacion)
	_, err = e.roles.Actualizar(ctx, r.ID, service.CambiosRol{Permisos: &[]string{"MAL"}})
	esperar(t, err, service.ErrValidacion)
	_, err = e.roles.Actualizar(ctx, uuid.New(), service.CambiosRol{Descripcion: &descripcion})
	esperar(t, err, service.ErrNoEncontrado)

	obtenido, err := e.roles.Obtener(ctx, r.ID)
	if err != nil || obtenido.Nombre != "recepcionista" {
		t.Fatal("el rol creado debe poder leerse")
	}
	lista, _ := e.roles.Listar(ctx)
	if len(lista) != 4 {
		t.Fatalf("se esperaban 4 roles, hay %d", len(lista))
	}

	if err := e.roles.Eliminar(ctx, r.ID); err != nil {
		t.Fatal(err)
	}
	_, err = e.roles.Obtener(ctx, r.ID)
	esperar(t, err, service.ErrNoEncontrado)
	esperar(t, e.roles.Eliminar(ctx, r.ID), service.ErrNoEncontrado)
}

func TestRolesProtegidosYAsignados(t *testing.T) {
	e := nuevoEntorno(t)
	ctx := context.Background()
	descripcion := "x"

	_, err := e.roles.Actualizar(ctx, memoria.IDRolAdministrador, service.CambiosRol{Descripcion: &descripcion})
	esperar(t, err, service.ErrConflicto)
	esperar(t, e.roles.Eliminar(ctx, memoria.IDRolAdministrador), service.ErrConflicto)
	esperar(t, e.roles.Eliminar(ctx, memoria.IDRolOperador), service.ErrConflicto)
}

func TestCatalogoDePermisos(t *testing.T) {
	e := nuevoEntorno(t)
	catalogo := e.roles.Permisos(context.Background())
	if len(catalogo) != 3 || catalogo[0].Codigo != "perfiles:operar" {
		t.Fatalf("catálogo inesperado: %+v", catalogo)
	}
	catalogo[0].Codigo = "alterado"
	if service.CatalogoPermisos()[0].Codigo != "perfiles:operar" {
		t.Fatal("el catálogo debe entregarse como copia")
	}
}

// --- Perfiles ---

func TestPerfilPropioYAjeno(t *testing.T) {
	e := nuevoEntorno(t)

	p, err := e.perfiles.Obtener(como(memoria.IDUsuarioUsuario), memoria.IDUsuarioUsuario)
	if err != nil || p.Nombre != "Ulises" {
		t.Fatalf("el usuario debe ver su perfil: %+v, %v", p, err)
	}
	_, err = e.perfiles.Obtener(como(memoria.IDUsuarioUsuario), memoria.IDUsuarioAdmin)
	esperar(t, err, service.ErrSinPermiso)
	if _, err := e.perfiles.Obtener(como(memoria.IDUsuarioOperador, "perfiles:operar"), memoria.IDUsuarioAdmin); err != nil {
		t.Fatalf("perfiles:operar debe ver perfiles ajenos: %v", err)
	}
	_, err = e.perfiles.Obtener(context.Background(), memoria.IDUsuarioUsuario)
	esperar(t, err, service.ErrNoAutenticado)
	_, err = e.perfiles.Obtener(como(uuid.Nil, "perfiles:operar"), uuid.New())
	esperar(t, err, service.ErrNoEncontrado)
}

func TestGuardarPerfil(t *testing.T) {
	e := nuevoEntorno(t)
	ctx := context.Background()
	nuevo, _ := e.usuarios.Crear(ctx, service.NuevoUsuario{Correo: "sin@perfil.cl", Contrasena: "clave-segura"})

	_, err := e.perfiles.Obtener(como(nuevo.ID), nuevo.ID)
	esperar(t, err, service.ErrNoEncontrado)

	telefono := "+56912345678"
	p, err := e.perfiles.Guardar(como(nuevo.ID), service.Perfil{UsuarioID: nuevo.ID, Nombre: "Sin", Apellido: "Perfil", Telefono: &telefono})
	if err != nil || p.Atributos == nil || !p.ActualizadoEn.Equal(instante) {
		t.Fatalf("perfil inesperado: %+v, %v", p, err)
	}
	_, err = e.perfiles.Guardar(como(nuevo.ID), service.Perfil{UsuarioID: memoria.IDUsuarioAdmin, Nombre: "X", Apellido: "Y"})
	esperar(t, err, service.ErrSinPermiso)
	_, err = e.perfiles.Guardar(como(uuid.Nil, "perfiles:operar"), service.Perfil{UsuarioID: uuid.New(), Nombre: "X", Apellido: "Y"})
	esperar(t, err, service.ErrNoEncontrado)
}

func TestErrorNegocioConservaLaCausa(t *testing.T) {
	e := nuevoEntorno(t)
	_, err := e.usuarios.Obtener(context.Background(), uuid.New())
	var negocio *service.ErrorNegocio
	if !errors.As(err, &negocio) || negocio.Error() != "El usuario no existe" {
		t.Fatalf("se esperaba un ErrorNegocio con mensaje, llegó %v", err)
	}
}
