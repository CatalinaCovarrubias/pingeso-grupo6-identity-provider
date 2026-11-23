package api

import (
	"net/http"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
)

const (
	rutaAdmin    = "/api/usuarios/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	rutaOperador = "/api/usuarios/bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	rutaUsuario  = "/api/usuarios/cccccccc-cccc-4ccc-8ccc-cccccccccccc"
	idRolUsuario = "33333333-3333-4333-8333-333333333333"
)

func TestListarUsuariosPaginado(t *testing.T) {
	s := nuevoServidorPrueba(t)
	admin := s.comoAdmin()

	r := s.llamar(peticion{metodo: http.MethodGet, ruta: "/api/usuarios?limite=2", cookie: admin}).
		esperar(http.StatusOK).json()
	datos := r["datos"].([]any)
	paginacion := r["paginacion"].(map[string]any)
	if len(datos) != 2 || paginacion["total"] != 3.0 || paginacion["total_paginas"] != 2.0 || paginacion["pagina"] != 1.0 {
		t.Fatalf("página inesperada: %v", r)
	}
	filtrado := s.llamar(peticion{metodo: http.MethodGet, ruta: "/api/usuarios?correo=opera&estado=activo", cookie: admin}).
		esperar(http.StatusOK).json()
	if len(filtrado["datos"].([]any)) != 1 {
		t.Fatalf("filtro inesperado: %v", filtrado)
	}
	for _, ruta := range []string{"/api/usuarios?limite=101", "/api/usuarios?pagina=0", "/api/usuarios?estado=raro"} {
		s.llamar(peticion{metodo: http.MethodGet, ruta: ruta, cookie: admin}).esperar(http.StatusUnprocessableEntity)
	}
}

func TestCrudDeUsuarios(t *testing.T) {
	s := nuevoServidorPrueba(t)
	admin := s.comoAdmin()

	creado := s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/usuarios", cookie: admin, cuerpo: map[string]any{
		"correo": "Ana@Ejemplo.cl", "contrasena": "una-clave-segura", "roles": []string{idRolUsuario},
	}}).esperar(http.StatusCreated).json()
	if creado["correo"] != "ana@ejemplo.cl" || creado["estado"] != "activo" {
		t.Fatalf("usuario creado inesperado: %v", creado)
	}
	ruta := "/api/usuarios/" + creado["id"].(string)

	s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/usuarios", cookie: admin, cuerpo: map[string]any{
		"correo": "ANA@ejemplo.cl", "contrasena": "una-clave-segura",
	}}).esperar(http.StatusConflict)
	s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/usuarios", cookie: admin, cuerpo: map[string]any{
		"correo": "b@ejemplo.cl", "contrasena": "corta",
	}}).esperar(http.StatusUnprocessableEntity)
	s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/usuarios", cookie: admin, cuerpo: map[string]any{
		"correo": "b@ejemplo.cl", "contrasena": "una-clave-segura", "roles": []string{"no-es-uuid"},
	}}).esperar(http.StatusUnprocessableEntity)
	s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/usuarios", cookie: admin, cuerpo: map[string]any{
		"correo": "b@ejemplo.cl", "contrasena": "una-clave-segura", "roles": []string{"dddddddd-dddd-4ddd-8ddd-dddddddddddd"},
	}}).esperar(http.StatusUnprocessableEntity)

	s.llamar(peticion{metodo: http.MethodGet, ruta: ruta, cookie: admin}).esperar(http.StatusOK)
	s.llamar(peticion{metodo: http.MethodGet, ruta: "/api/usuarios/dddddddd-dddd-4ddd-8ddd-dddddddddddd", cookie: admin}).
		esperar(http.StatusNotFound)
	s.llamar(peticion{metodo: http.MethodGet, ruta: "/api/usuarios/no-es-uuid", cookie: admin}).
		esperar(http.StatusUnprocessableEntity)

	editado := s.llamar(peticion{metodo: http.MethodPatch, ruta: ruta, cookie: admin, cuerpo: map[string]any{
		"estado": "bloqueado",
	}}).esperar(http.StatusOK).json()
	if editado["estado"] != "bloqueado" {
		t.Fatalf("edición inesperada: %v", editado)
	}
	s.llamar(peticion{metodo: http.MethodPatch, ruta: ruta, cookie: admin, cuerpo: map[string]any{}}).
		esperar(http.StatusUnprocessableEntity)
	s.llamar(peticion{metodo: http.MethodPatch, ruta: ruta, cookie: admin, cuerpo: map[string]any{"estado": "borrado"}}).
		esperar(http.StatusUnprocessableEntity)

	s.llamar(peticion{metodo: http.MethodDelete, ruta: ruta, cookie: admin}).esperar(http.StatusNoContent)
	final := s.llamar(peticion{metodo: http.MethodGet, ruta: ruta, cookie: admin}).esperar(http.StatusOK).json()
	if final["estado"] != "inactivo" {
		t.Fatalf("DELETE debe dejar el usuario inactivo: %v", final)
	}
}

func TestNoSeDesactivaAlUltimoAdministrador(t *testing.T) {
	s := nuevoServidorPrueba(t)
	admin := s.comoAdmin()
	s.llamar(peticion{metodo: http.MethodDelete, ruta: rutaAdmin, cookie: admin}).esperar(http.StatusConflict)
	s.llamar(peticion{metodo: http.MethodPut, ruta: rutaAdmin + "/roles", cookie: admin,
		cuerpo: map[string]any{"roles": []string{}}}).esperar(http.StatusConflict)
}

func TestRolesDeUsuario(t *testing.T) {
	s := nuevoServidorPrueba(t)
	admin := s.comoAdmin()

	r := s.llamar(peticion{metodo: http.MethodGet, ruta: rutaOperador + "/roles", cookie: admin}).
		esperar(http.StatusOK).json()
	if roles := r["roles"].([]any); len(roles) != 1 || roles[0].(map[string]any)["nombre"] != "operador" {
		t.Fatalf("roles inesperados: %v", r)
	}
	nuevos := s.llamar(peticion{metodo: http.MethodPut, ruta: rutaUsuario + "/roles", cookie: admin,
		cuerpo: map[string]any{"roles": []string{idRolUsuario, ejemploIDRol}}}).esperar(http.StatusOK).json()
	if len(nuevos["roles"].([]any)) != 2 {
		t.Fatalf("asignación inesperada: %v", nuevos)
	}
	s.llamar(peticion{metodo: http.MethodPut, ruta: rutaUsuario + "/roles", cookie: admin,
		cuerpo: map[string]any{"roles": []string{idRolUsuario, idRolUsuario}}}).esperar(http.StatusUnprocessableEntity)
}

func TestCrudDeRoles(t *testing.T) {
	s := nuevoServidorPrueba(t)
	admin := s.comoAdmin()

	creado := s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/roles", cookie: admin, cuerpo: map[string]any{
		"nombre": "recepcionista", "descripcion": "Atiende", "permisos": []string{"reservas:operar", "perfiles:operar"},
	}}).esperar(http.StatusCreated).json()
	ruta := "/api/roles/" + creado["id"].(string)
	if permisos := creado["permisos"].([]any); permisos[0] != "perfiles:operar" {
		t.Fatalf("los permisos deben venir ordenados: %v", permisos)
	}

	invalidos := []map[string]any{
		{"nombre": "Con Espacios", "permisos": []string{}},
		{"nombre": "ok_rol", "permisos": []string{"sin-dos-puntos"}},
		{"nombre": "ok_rol", "permisos": []string{"Reservas:Operar"}},
		{"nombre": "ok_rol", "permisos": []string{"a:b", "a:b"}},
	}
	for _, cuerpo := range invalidos {
		s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/roles", cookie: admin, cuerpo: cuerpo}).
			esperar(http.StatusUnprocessableEntity)
	}
	s.llamar(peticion{metodo: http.MethodPost, ruta: "/api/roles", cookie: admin, cuerpo: map[string]any{
		"nombre": "recepcionista", "permisos": []string{},
	}}).esperar(http.StatusConflict)

	s.llamar(peticion{metodo: http.MethodGet, ruta: ruta, cookie: admin}).esperar(http.StatusOK)
	lista := s.llamar(peticion{metodo: http.MethodGet, ruta: "/api/roles", cookie: admin}).esperar(http.StatusOK).json()
	if len(lista["datos"].([]any)) != 4 {
		t.Fatalf("se esperaban 4 roles: %v", lista)
	}

	editado := s.llamar(peticion{metodo: http.MethodPatch, ruta: ruta, cookie: admin, cuerpo: map[string]any{
		"permisos": []string{"reservas:administrar"},
	}}).esperar(http.StatusOK).json()
	if editado["descripcion"] != "Atiende" || len(editado["permisos"].([]any)) != 1 {
		t.Fatalf("edición inesperada: %v", editado)
	}
	s.llamar(peticion{metodo: http.MethodPatch, ruta: ruta, cookie: admin, cuerpo: map[string]any{}}).
		esperar(http.StatusUnprocessableEntity)
	s.llamar(peticion{metodo: http.MethodPatch, ruta: "/api/roles/" + ejemploIDRol, cookie: admin,
		cuerpo: map[string]any{"descripcion": "x"}}).esperar(http.StatusConflict)

	s.llamar(peticion{metodo: http.MethodDelete, ruta: "/api/roles/" + idRolUsuario, cookie: admin}).
		esperar(http.StatusConflict)
	s.llamar(peticion{metodo: http.MethodDelete, ruta: ruta, cookie: admin}).esperar(http.StatusNoContent)
	s.llamar(peticion{metodo: http.MethodGet, ruta: ruta, cookie: admin}).esperar(http.StatusNotFound)
}

func TestPermisosParaCualquierSesion(t *testing.T) {
	s := nuevoServidorPrueba(t)
	r := s.llamar(peticion{metodo: http.MethodGet, ruta: "/api/permisos", cookie: s.comoUsuario()}).
		esperar(http.StatusOK).json()
	if len(r["datos"].([]any)) != 3 {
		t.Fatalf("catálogo inesperado: %v", r)
	}
	s.llamar(peticion{metodo: http.MethodGet, ruta: "/api/permisos"}).esperar(http.StatusUnauthorized)
}

func TestPerfiles(t *testing.T) {
	s := nuevoServidorPrueba(t)
	usuario := s.comoUsuario()

	propio := s.llamar(peticion{metodo: http.MethodGet, ruta: rutaUsuario + "/perfil", cookie: usuario}).
		esperar(http.StatusOK).json()
	if propio["nombre"] != "Ulises" {
		t.Fatalf("perfil inesperado: %v", propio)
	}
	s.llamar(peticion{metodo: http.MethodGet, ruta: rutaAdmin + "/perfil", cookie: usuario}).esperar(http.StatusForbidden)
	s.llamar(peticion{metodo: http.MethodGet, ruta: rutaAdmin + "/perfil", cookie: s.comoOperador()}).esperar(http.StatusOK)

	guardado := s.llamar(peticion{metodo: http.MethodPut, ruta: rutaUsuario + "/perfil", cookie: usuario, cuerpo: map[string]any{
		"nombre": "Ulises", "apellido": "Pérez", "telefono": "+56912345678", "atributos": map[string]any{"estatura_cm": 172},
	}}).esperar(http.StatusOK).json()
	if guardado["telefono"] != "+56912345678" || guardado["atributos"].(map[string]any)["estatura_cm"] != 172.0 {
		t.Fatalf("perfil guardado inesperado: %v", guardado)
	}
	s.llamar(peticion{metodo: http.MethodPut, ruta: rutaUsuario + "/perfil", cookie: usuario, cuerpo: map[string]any{
		"nombre": "", "apellido": "Pérez",
	}}).esperar(http.StatusUnprocessableEntity)
	s.llamar(peticion{metodo: http.MethodPut, ruta: rutaUsuario + "/perfil", cookie: usuario, cuerpo: map[string]any{
		"nombre": "U", "apellido": "P", "telefono": "llámame",
	}}).esperar(http.StatusUnprocessableEntity)
	s.llamar(peticion{metodo: http.MethodPut, ruta: rutaAdmin + "/perfil", cookie: usuario, cuerpo: map[string]any{
		"nombre": "U", "apellido": "P",
	}}).esperar(http.StatusForbidden)
}

func TestCadaOperacionDeclaraSuPermiso(t *testing.T) {
	_, api := NuevoServidor(Dependencias{PublicarDocs: true, RegistrarDev: true})
	total := 0
	for ruta, item := range api.OpenAPI().Paths {
		operaciones := map[string]*huma.Operation{
			"GET": item.Get, "POST": item.Post, "PUT": item.Put, "PATCH": item.Patch, "DELETE": item.Delete,
		}
		for metodo, op := range operaciones {
			if op == nil {
				continue
			}
			total++
			permiso, _ := op.Extensions["x-permiso"].(string)
			if permiso == "" {
				t.Errorf("%s %s no declara x-permiso", metodo, ruta)
				continue
			}
			publica := permiso == permisoPublico
			if publica != (len(op.Security) == 0) {
				t.Errorf("%s %s: x-permiso=%s no coincide con security=%v", metodo, ruta, permiso, op.Security)
			}
			if !publica && permiso != permisoAutenticado && !strings.Contains(permiso, ":") {
				t.Errorf("%s %s: permiso con formato inesperado %q", metodo, ruta, permiso)
			}
		}
	}
	if total < 20 {
		t.Fatalf("se esperaban al menos 20 operaciones, hay %d", total)
	}
}
