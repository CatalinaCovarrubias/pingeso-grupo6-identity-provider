# JWT de prueba para los otros grupos

Esta guía explica cómo obtener tokens válidos del servicio de identidad (grupo 6) y cómo
validarlos en tu servicio. Los tokens son reales: van firmados con RS256 y se validan con
el JWKS publicado, igual que en producción. Lo único simulado es el almacenamiento: los
usuarios viven en memoria y se pierden al reiniciar.

El contrato completo está en [`contracts/openapi.yaml`](../contracts/openapi.yaml). Cópialo a
`contracts/externos/identidad.yaml` en tu repo.

## 1. Levantar el servicio

Con Docker Compose, desde la raíz de este repo:

```bash
cp .env.example .env
# Completa POSTGRES_PASSWORD y DATABASE_URL. Deja JWT_LLAVE_PRIVADA_ARCHIVO vacía:
# en dev el servicio genera una llave en memoria al arrancar.
docker compose up -d --build
```

La API queda en `http://localhost:5161`, y la documentación interactiva en
`http://localhost:5161/api/docs`.

Sin Docker, con Go 1.26:

```bash
cd apps/api
API_PORT=5161 API_ENV=dev DATABASE_URL=postgres://x:y@localhost:5432/z \
FRONTEND_URL=http://localhost:5160 TZ=America/Santiago \
JWT_EMISOR=pingeso-identidad JWT_AUDIENCIA=pingeso JWT_TTL=15m \
go run ./cmd/api
```

En Windows usa PowerShell (`$env:TZ='America/Santiago'`, y así con cada variable), porque Git
Bash no le pasa `TZ` a los programas de Windows. La base de datos no se usa todavía; solo
`/readyz` la consulta.

> **Ojo con la llave temporal.** Si `JWT_LLAVE_PRIVADA_ARCHIVO` está vacía, cada reinicio genera
> una llave nueva con otro `kid` y los tokens anteriores dejan de valer. `keyfunc` vuelve a
> descargar el JWKS cuando ve un `kid` desconocido, así que basta con pedir un token nuevo.
> Si necesitas tokens estables entre reinicios, genera una llave fija:
> `openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out .secretos/jwt.pem`
> y apunta `JWT_LLAVE_PRIVADA_ARCHIVO` a ese archivo. `.secretos/` y `*.pem` no se suben al repo.

## 2. Usuarios seed

Existen solo con `API_ENV=dev` o `test`. Las contraseñas son públicas a propósito.

| Correo               | Contraseña           | Rol             | ID del usuario                         |
| -------------------- | -------------------- | --------------- | -------------------------------------- |
| `admin@dev.local`    | `admin-dev-local`    | `administrador` | `aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa` |
| `operador@dev.local` | `operador-dev-local` | `operador`      | `bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb` |
| `usuario@dev.local`  | `usuario-dev-local`  | `usuario`       | `cccccccc-cccc-4ccc-8ccc-cccccccccccc` |

| Rol             | ID                                     | Permisos                                                       |
| --------------- | -------------------------------------- | -------------------------------------------------------------- |
| `administrador` | `11111111-1111-4111-8111-111111111111` | `perfiles:operar`, `roles:administrar`, `usuarios:administrar` |
| `operador`      | `22222222-2222-4222-8222-222222222222` | `perfiles:operar`                                              |
| `usuario`       | `33333333-3333-4333-8333-333333333333` | ninguno: solo su sesión y su perfil                            |

## 3. Obtener un token

### Opción A: login, igual que el navegador

El login deja el token en la cookie `sesion` (`HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/`) y
**no** lo devuelve en el body.

```bash
curl -i -c cookies.txt -H 'Content-Type: application/json' \
  -d '{"correo":"admin@dev.local","contrasena":"admin-dev-local"}' \
  http://localhost:5161/api/sesiones

curl -b cookies.txt http://localhost:5161/api/sesiones/actual
```

### Opción B: `POST /api/dev/tokens`, para llamadas entre servicios

Existe solo con `API_ENV=dev`; con `test` o `prod` responde 404. Devuelve el token en el body,
para que lo uses en `Authorization: Bearer`. Acepta los roles seed y, además,
**permisos de tu propio servicio** en `permisos_extra`.

```bash
curl -s -H 'Content-Type: application/json' \
  -d '{"roles":["operador"],"permisos_extra":["reservas:operar"]}' \
  http://localhost:5161/api/dev/tokens
```

| Campo            | Obligatorio | Qué hace                                                           |
| ---------------- | ----------- | ------------------------------------------------------------------ |
| `roles`          | no          | Nombres de roles existentes. Sus permisos entran al token          |
| `permisos_extra` | no          | Permisos adicionales con formato `<recurso>:<accion>` (minúsculas) |
| `usuario_id`     | no          | UUID que irá en `sub`. Si falta, se inventa uno                    |

Un token con un `usuario_id` inventado sirve para tu servicio. En cambio,
`GET /api/sesiones/actual` lo rechaza con 401, porque ese usuario no existe aquí. Si necesitas
que exista, usa uno de los IDs seed.

Respuesta:

```json
{
  "token": "eyJhbGciOiJSUzI1NiIsImtpZCI6Ii4uLiIsInR5cCI6IkpXVCJ9.<payload>.<firma>",
  "expira_en": "2026-11-23T15:15:00Z",
  "claims": { "...": "los mismos de la sección 4" }
}
```

Luego, en tus llamadas:

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:5161/api/sesiones/actual
```

## 4. Claims del token

```json
// header
{ "alg": "RS256", "typ": "JWT", "kid": "eTarqguzfSNc6aYMSuIX1nsITVo0zt8xS0by3E3l32g" }
// payload
{
  "iss": "pingeso-identidad",
  "sub": "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
  "aud": ["pingeso"],
  "exp": 1795464900,
  "iat": 1795464000,
  "nbf": 1795464000,
  "jti": "5f2d8a4e-1b3c-4d7e-8f90-a1b2c3d4e5f6",
  "roles": ["operador"],
  "permisos": ["perfiles:operar", "reservas:operar"]
}
```

- `sub` es el UUID del usuario. Úsalo como referencia al usuario en tus datos.
- `aud` es un arreglo con un solo valor.
- El token dura 15 minutos (`JWT_TTL`). No hay refresh real todavía.
- `permisos` es la unión de los permisos de sus roles. **Autoriza siempre por permiso, no por
  nombre de rol**: los roles los configura cada producto.
- El `kid` cambia cuando cambia la llave. No lo copies fijo en tu código.

## 5. Validar el token en tu servicio

Configura `AUTH_JWKS_URL` según dónde corra tu servicio:

| Tu servicio corre en                    | `AUTH_JWKS_URL`                                                      |
| --------------------------------------- | -------------------------------------------------------------------- |
| Tu máquina, sin Docker                  | `http://localhost:5161/.well-known/jwks.json`                        |
| Docker, con este servicio en tu máquina | `http://host.docker.internal:5161/.well-known/jwks.json`             |
| La misma red de Compose                 | `http://backend:8080/.well-known/jwks.json` (el nombre del servicio) |

El JWKS es público y responde con `Cache-Control: public, max-age=300`. Incluye un campo extra
`$schema` que agrega Huma; `keyfunc` lo ignora.

Ejemplo en Go con `github.com/MicahParks/keyfunc/v3 v3.9.1` y `github.com/golang-jwt/jwt/v5 v5.3.1`.
Lo compilamos y probamos contra un token y un JWKS reales de este servicio:

```go
package identidad

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

// Claims son los claims que emite el servicio de identidad.
type Claims struct {
	jwt.RegisteredClaims
	Roles    []string `json:"roles"`
	Permisos []string `json:"permisos"`
}

// Validador valida tokens del servicio de identidad con su JWKS.
type Validador struct {
	jwks keyfunc.Keyfunc
}

// NuevoValidador descarga el JWKS de AUTH_JWKS_URL y lo mantiene actualizado en segundo plano.
func NuevoValidador(ctx context.Context, jwksURL string) (*Validador, error) {
	k, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el JWKS: %w", err)
	}
	return &Validador{jwks: k}, nil
}

// Validar comprueba firma RS256, iss, aud y exp, y que el token traiga el permiso pedido.
func (v *Validador) Validar(token, permiso string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, v.jwks.Keyfunc,
		jwt.WithValidMethods([]string{"RS256"}), // rechaza alg none y HS256
		jwt.WithIssuer("pingeso-identidad"),
		jwt.WithAudience("pingeso"),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("token inválido: %w", err) // responder 401
	}
	if !slices.Contains(claims.Permisos, permiso) {
		return nil, errors.New("falta el permiso " + permiso) // responder 403
	}
	return claims, nil
}
```

Cómo leer el token en tu middleware: primero `Authorization: Bearer <token>`, y si no viene, la
cookie `sesion`. Cuando tu servicio llama a otro, reenvía el mismo token en `Authorization`.

## 6. Errores que vas a ver

Todos los errores son `application/problem+json` (RFC 9457):

| Estado | Cuándo                                                                                               |
| ------ | ---------------------------------------------------------------------------------------------------- |
| 401    | Sin token, token inválido o vencido, o credenciales incorrectas en el login                          |
| 403    | Token válido sin el permiso del endpoint, o login de una cuenta bloqueada o inactiva                 |
| 404    | Recurso inexistente, o `/api/dev/tokens` fuera de `dev`                                              |
| 409    | Correo o nombre de rol repetido, rol asignado que se quiere borrar, o quitar al último administrador |
| 422    | El body o los parámetros no cumplen el contrato                                                      |

## 7. Qué está simulado en el Hito 1

- Los datos viven en memoria y se pierden al reiniciar. PostgreSQL llega en la Tanda 4.
- `POST /api/sesiones/actual/renovacion` vuelve a emitir el token con los mismos roles y permisos.
- `PUT /api/sesiones/actual/contrasena` valida el formato, pero no cambia la contraseña.
- Si se revoca un rol o se bloquea a un usuario, sus tokens ya emitidos siguen valiendo hasta su
  `exp` (15 minutos como máximo). Es una ventana aceptada.
