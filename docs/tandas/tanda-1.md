# Tanda 1 — Contrato, mocks y tokens de prueba

Rama `feat/tanda-1-contrato-mocks`, creada desde `main` (`f4a0eff`). Sin push, sin PR y sin tags.
Fecha: 23 de noviembre de 2026.

## Objetivo

Publicar el contrato completo del servicio de identidad con endpoints que responden con
datos simulados, y entregar a los otros grupos **tokens JWT reales**: firmados con RS256 y
validables con el JWKS. Así pueden avanzar en paralelo antes del 4 de diciembre (Hitos 1 y 2).

## Qué se hizo

| Commit                                                          | Qué incluye                                                                                                                                                                                                                    |
| --------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `feat(config): variables JWT y llave de firma`                  | `JWT_EMISOR`, `JWT_AUDIENCIA`, `JWT_TTL` (entre 1 min y 24 h) y `JWT_LLAVE_PRIVADA_ARCHIVO` (obligatoria salvo en dev), en config, `.env.example` y Compose. `.secretos/` y `*.pem` agregados a `.gitignore` y `.dockerignore` |
| `feat(auth): firma y validación RS256 con JWKS`                 | Paquete de tokens: carga PEM PKCS#8 o PKCS#1 de 2048 bits o más, `kid` RFC 7638, firma RS256 y validación con algoritmo fijo, `iss`, `aud`, `exp` obligatorio y 30 s de desfase. La API termina si la llave no carga           |
| `feat(store): store en memoria con usuarios y roles seed`       | Modelo de dominio, errores y puertos de almacenamiento. Store en memoria con correo único sin distinguir mayúsculas y seed con IDs fijos y hash bcrypt                                                                         |
| `feat(service): reglas de sesiones, usuarios, roles y perfiles` | Login sin revelar correos, regla del último administrador, rol administrador protegido, rol asignado no se borra, perfil propio o `perfiles:operar`, tokens de prueba                                                          |
| `feat(api): contrato y handlers de identidad`                   | Handlers Huma, middleware Bearer o cookie con `x-permiso`, errores RFC 9457 y `contracts/openapi.yaml` regenerado (versión 0.2.0)                                                                                              |
| `feat(api): emisión de tokens de prueba en desarrollo`          | `POST /api/dev/tokens`, que solo existe con `API_ENV=dev`                                                                                                                                                                      |
| `docs: guía de JWT de prueba y modos de falla`                  | `docs/jwt-de-prueba.md`, `docs/modos-de-falla.md` y enlace en el README                                                                                                                                                        |
| `feat(api): duración opcional en los tokens de prueba`          | `duracion` en `POST /api/dev/tokens` (respuesta del equipo a la pregunta 3)                                                                                                                                                    |

Archivos principales:

| Capa               | Archivos                                                                                                                                                                  |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| HTTP: solo traduce | `apps/api/internal/api/`: `servidor.go`, `autenticacion.go`, `servicios.go` (interfaces), `sesiones.go`, `usuarios.go`, `roles.go`, `perfiles.go`, `dev.go`, `comunes.go` |
| Reglas de negocio  | `apps/api/internal/service/`: `sesiones.go`, `usuarios.go`, `roles.go`, `perfiles.go`, `sesion.go`, `comunes.go`, `modelo.go`, `errores.go`, `puertos.go`                 |
| Tokens             | `apps/api/internal/service/tokens/`: `llave.go`, `emisor.go`                                                                                                              |
| Store simulado     | `apps/api/internal/store/memoria/`: `memoria.go`, `seed.go`                                                                                                               |
| Armado             | `apps/api/cmd/api/main.go`, `apps/api/internal/config/config.go`                                                                                                          |
| Contrato           | `contracts/openapi.yaml`                                                                                                                                                  |

Capas: los handlers solo dependen de interfaces del service (`servicios.go`) y el service
depende de los puertos (`puertos.go`). Ningún archivo no-test de `internal/api` importa el
store. En la Tanda 4 se cambia `memoria` por PostgreSQL sin tocar los handlers.

## Endpoints

| Método y ruta                          | Uso                                                                             | `x-permiso`                        | Estados                    |
| -------------------------------------- | ------------------------------------------------------------------------------- | ---------------------------------- | -------------------------- |
| `POST /api/sesiones`                   | Login: fija la cookie `sesion` y devuelve el usuario (sin token en el body)     | `publico`                          | 200, 401, 403, 422         |
| `GET /api/sesiones/actual`             | Usuario de la sesión con sus permisos y `expira_en`                             | `autenticado`                      | 200, 401                   |
| `DELETE /api/sesiones/actual`          | Logout: borra la cookie                                                         | `autenticado`                      | 204, 401                   |
| `POST /api/sesiones/actual/renovacion` | Renovación (simulada): token nuevo, mismos permisos                             | `autenticado`                      | 200, 401                   |
| `PUT /api/sesiones/actual/contrasena`  | Cambio de contraseña (simulado): valida el formato y no cambia nada             | `autenticado`                      | 204, 401, 422              |
| `GET /.well-known/jwks.json`           | Llaves públicas, `Cache-Control: public, max-age=300`                           | `publico`                          | 200                        |
| `GET /api/usuarios`                    | Listado paginado (`pagina`, `limite` ≤ 100, `estado`, `correo`)                 | `usuarios:administrar`             | 200, 401, 403, 422         |
| `POST /api/usuarios`                   | Crear usuario activo                                                            | `usuarios:administrar`             | 201, 409, 422              |
| `GET /api/usuarios/{id}`               | Detalle                                                                         | `usuarios:administrar`             | 200, 404, 422              |
| `PATCH /api/usuarios/{id}`             | Cambiar correo o estado (`activo`, `bloqueado`, `inactivo`)                     | `usuarios:administrar`             | 200, 404, 409, 422         |
| `DELETE /api/usuarios/{id}`            | Desactivar (estado `inactivo`)                                                  | `usuarios:administrar`             | 204, 404, 409              |
| `GET /api/usuarios/{id}/roles`         | Roles asignados                                                                 | `roles:administrar`                | 200, 404                   |
| `PUT /api/usuarios/{id}/roles`         | Reemplazar roles                                                                | `roles:administrar`                | 200, 404, 409, 422         |
| `GET /api/usuarios/{id}/perfil`        | Perfil genérico                                                                 | `perfiles:operar` o propio usuario | 200, 403, 404              |
| `PUT /api/usuarios/{id}/perfil`        | Crear o reemplazar perfil                                                       | `perfiles:operar` o propio usuario | 200, 403, 404, 422         |
| `GET /api/roles`                       | Listar roles                                                                    | `roles:administrar`                | 200                        |
| `POST /api/roles`                      | Crear rol con permisos `<recurso>:<accion>` de cualquier servicio               | `roles:administrar`                | 201, 409, 422              |
| `GET /api/roles/{id}`                  | Detalle                                                                         | `roles:administrar`                | 200, 404                   |
| `PATCH /api/roles/{id}`                | Cambiar descripción o permisos                                                  | `roles:administrar`                | 200, 404, 409, 422         |
| `DELETE /api/roles/{id}`               | Eliminar rol no asignado                                                        | `roles:administrar`                | 204, 404, 409              |
| `GET /api/permisos`                    | Catálogo de permisos propios                                                    | `autenticado`                      | 200, 401                   |
| `POST /api/dev/tokens`                 | Token de prueba con `roles`, `permisos_extra` y `duracion` (solo `API_ENV=dev`) | `publico` en dev                   | 200, 422; 404 fuera de dev |
| `GET /healthz`, `GET /readyz`          | Salud (sin cambios)                                                             | `publico`                          | 200, 503                   |

Cambios respecto de la tabla del prompt:

- `GET /api/usuarios` tiene los filtros `estado` y `correo`.
- `POST /api/dev/tokens` tiene `permisos_extra` y, tras la respuesta del equipo, `duracion`.
- Las reglas 409 (último administrador, rol asignado y rol administrador protegido) quedaron implementadas.

## Decisiones aplicadas

| #          | Decisión                                                                                                                       | Dónde quedó                                              |
| ---------- | ------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------- |
| D1         | Rutas `/api/...` en español y plural; JWKS en `/.well-known/jwks.json`                                                         | Todos los handlers                                       |
| D2         | RS256 con `golang-jwt/jwt/v5 v5.3.1` y llave RSA de 2048 bits o más                                                            | `service/tokens`                                         |
| D3         | Cookie `sesion` `HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/`; también `Authorization: Bearer`; el login no devuelve el token | `api/autenticacion.go`, `api/sesiones.go`                |
| D4         | `JWT_TTL=15m`; renovación simulada                                                                                             | `config`, `service/sesiones.go`                          |
| D5         | `GET` y `PUT /api/usuarios/{id}/perfil`                                                                                        | `api/perfiles.go`                                        |
| D9         | Usuarios seed en memoria y `POST /api/dev/tokens` solo con `API_ENV=dev`                                                       | `store/memoria/seed.go`, `api/dev.go`, `cmd/api/main.go` |
| D12        | `usuarios:administrar`, `roles:administrar`, `perfiles:operar`; roles `administrador`, `operador`, `usuario`                   | `service/modelo.go`, `store/memoria/seed.go`             |
| D13        | Claims `iss`, `sub`, `aud`, `exp`, `iat`, `nbf`, `jti`, `roles` y `permisos`; header con `kid`                                 | `service/tokens/emisor.go`                               |
| Llave      | `JWT_LLAVE_PRIVADA_ARCHIVO`; en dev sin llave se genera una en memoria con aviso en el log                                     | `cmd/api/main.go` (`obtenerLlave`)                       |
| Pregunta 1 | `permisos_extra` en tokens de prueba; los roles aceptan cualquier permiso con el patrón                                        | `api/dev.go`, `service/roles.go`                         |
| Pregunta 2 | Estado `inactivo` en el contrato; `DELETE` desactiva                                                                           | `service/usuarios.go`                                    |
| Pregunta 4 | Renovación y cambio de contraseña simulados                                                                                    | `service/sesiones.go`                                    |
| Pregunta 5 | `/api/dev/tokens` en el contrato con la etiqueta `Desarrollo`                                                                  | `cmd/api/main.go` (`imprimirContrato`)                   |
| Pregunta 6 | `JWT_EMISOR=pingeso-identidad` y `JWT_AUDIENCIA=pingeso` fijos                                                                 | `.env.example`                                           |

### Desviaciones del DAS para pedir al jefe de proyecto

| Componente               | DAS                 | Repo                | Motivo                                                                                                                                                              |
| ------------------------ | ------------------- | ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `golang.org/x/crypto`    | v0.54.0             | **v0.58.0**         | `google.golang.org/grpc v1.83.2` (que entra por goose como `tool`) exige ≥ v0.55.0, y `x/net v0.62.0` sube a v0.58.0. No se puede bajar sin sacar goose de `go.mod` |
| `github.com/google/uuid` | no está en la tabla | **v1.6.0**, directa | `jti` e IDs. Ya estaba en el grafo, porque Huma lo requiere                                                                                                         |

El equipo aceptó ambas por ahora (23 de noviembre). Quedan para pedirle al jefe de proyecto que
las agregue a la tabla del DAS (§3, Backend). La salida de goose de `go.mod` va en la Tanda 3.

### Decisiones propias, confirmadas por el equipo

- **Seed solo en dev y test.** En `prod` el store en memoria arranca vacío, con un aviso en el
  log, para no publicar credenciales conocidas en un ambiente productivo.
- **Rol `administrador` protegido.** No se puede editar ni eliminar (409), para que nadie le
  quite permisos y deje el sistema sin administración (MF-29).
- **Cuenta no activa.** El login de una cuenta `bloqueado` o `inactivo` da 403, pero solo con la
  contraseña correcta. Con la contraseña mala da el mismo 401 genérico.
- **Tokens de prueba con `usuario_id` inventado.** Sirven para los otros servicios, pero
  `GET /api/sesiones/actual` los rechaza con 401 porque el usuario no existe aquí.

### Respuestas del equipo a las preguntas de la primera entrega

| Pregunta                         | Respuesta                    | Qué se hizo                                                                                                                                                                                                                                                                        |
| -------------------------------- | ---------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Desviaciones del DAS             | Aceptadas por ahora          | Anotadas arriba; goose sale de `go.mod` en la Tanda 3                                                                                                                                                                                                                              |
| Decisiones propias               | Confirmadas las cuatro       | —                                                                                                                                                                                                                                                                                  |
| Duración de los tokens de prueba | Sí, agregar `duracion`       | `POST /api/dev/tokens` acepta `duracion` de `1m` a `1440m` o de `1h` a `24h`; por defecto `JWT_TTL` (15m). El patrón del contrato acota el máximo y el service lo vuelve a validar. Pruebas: `TestTokenDePruebaConDuracion`, `TestTokenPruebaConDuracion`, `TestEmitirConDuracion` |
| Correo en los claims             | No                           | Menos datos personales en el token: quien lo necesite lo pide a `GET /api/sesiones/actual`. Anotado en `docs/jwt-de-prueba.md`                                                                                                                                                     |
| Renombrar "usuarios-roles"       | Sí, en la Tanda 3            | Junto con el traspaso a `PINGESO-CIVIL`                                                                                                                                                                                                                                            |
| Publicar el contrato             | Push y PR los hace el equipo | Cuando esté en `main` se comparte el enlace a `contracts/openapi.yaml` y `docs/jwt-de-prueba.md`                                                                                                                                                                                   |

## Cómo probarlo

```bash
cd apps/api
gofmt -l .
go vet ./...
go test ./...
cd ../..
pnpm codegen && git diff --exit-code
pnpm verify
```

Levantar la API sin Docker, en PowerShell (Git Bash no le pasa `TZ` a los programas de Windows):

```powershell
cd apps/api
$env:API_PORT='5161'; $env:API_ENV='dev'; $env:DATABASE_URL='postgres://x:y@localhost:5432/z'
$env:FRONTEND_URL='http://localhost:5160'; $env:TZ='America/Santiago'
$env:JWT_EMISOR='pingeso-identidad'; $env:JWT_AUDIENCIA='pingeso'; $env:JWT_TTL='15m'
go run ./cmd/api
```

Probarla con curl:

```bash
curl -i -c cookies.txt -H 'Content-Type: application/json' \
  -d '{"correo":"admin@dev.local","contrasena":"admin-dev-local"}' http://localhost:5161/api/sesiones
curl -b cookies.txt http://localhost:5161/api/usuarios
curl -s -H 'Content-Type: application/json' \
  -d '{"roles":["operador"],"permisos_extra":["reservas:operar"]}' http://localhost:5161/api/dev/tokens
curl http://localhost:5161/.well-known/jwks.json
```

Documentación interactiva: `http://localhost:5161/api/docs`.

### Resultados de esta tanda

| Verificación                                                         | Resultado                                                                                                                                                                                |
| -------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `gofmt -l .`                                                         | ✅ sin archivos                                                                                                                                                                          |
| `go vet ./...`                                                       | ✅                                                                                                                                                                                       |
| `go test ./...`                                                      | ✅ cobertura: cmd/api 79.7 %, api 96.1 %, config 100 %, service 91.6 %, tokens 92.8 %, memoria 95.9 %                                                                                    |
| `pnpm codegen` + `git diff --exit-code`                              | ✅ sin diferencias                                                                                                                                                                       |
| `pnpm verify`                                                        | ✅ Tuve que pasar Prettier por `docs/auditoria.md`, que no está commiteado                                                                                                               |
| oasdiff v1.35.0 `breaking` contra `main`                             | ✅ "No breaking changes to report"                                                                                                                                                       |
| Humo con el servidor real (`ejecutar()` en el puerto 5161, modo dev) | ✅ Login 200 con cookie `HttpOnly; Secure; SameSite=Lax`; sesión actual 200; token de prueba 200; operador en `/api/usuarios` 403; JWKS 200; `/api/docs` 200; logout 204 con `Max-Age=0` |
| Ejemplo de `docs/jwt-de-prueba.md` con keyfunc v3.9.1                | ✅ Compilado en un módulo aparte; valida un token real y rechaza la falta de permiso y el token alterado                                                                                 |
| `docker compose up -d --build` (detalle abajo)                       | ✅ Corrido dos veces, con permiso del equipo; la última sobre `665d92e`                                                                                                                  |
| `go test -race`                                                      | ⏸️ No se pudo: en esta máquina no hay gcc para cgo. La concurrencia del store se prueba sin `-race`                                                                                      |

Nota: Windows Defender bloqueó el `.exe` compilado en `%TEMP%` (falso positivo común en binarios
Go sin firma). Por eso la prueba de humo se hizo con una prueba Go temporal, que no quedó en el repo.

### Docker Compose (23 de noviembre)

Docker 29.7.2 y Compose v5.3.1, con un `.env` local creado desde `.env.example`: contraseña de
PostgreSQL aleatoria y `JWT_LLAVE_PRIVADA_ARCHIVO` vacía. El `.env` no se commitea. Se corrió primero
sobre `2c7101d` y luego sobre `665d92e`, con el mismo resultado. Al final, `docker compose down`.

| Verificación en `127.0.0.1:5161`           | Resultado                                                                                                                               |
| ------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------- |
| Orden de arranque                          | ✅ `postgres` llega a `healthy`, `migrate` aplica `00001_esquema_inicial.sql` y termina con código 0, y después arranca `backend`       |
| Puertos                                    | ✅ `backend` publicado solo en `127.0.0.1:5161->8080`; `postgres` sin puerto publicado                                                  |
| Log del backend                            | ✅ JSON con el aviso de llave temporal, `kid` y "servidor escuchando" en el puerto 8080 con ambiente `dev`                              |
| `GET /healthz`                             | ✅ 200 `{"estado":"ok"}`                                                                                                                |
| `GET /readyz`                              | ✅ 200 con PostgreSQL arriba. Con `docker compose stop postgres`, 503 `application/problem+json` "la base de datos no responde" (MF-01) |
| `GET /.well-known/jwks.json`               | ✅ 200, `Content-Type: application/json`, `Cache-Control: public, max-age=300`, una llave RS256                                         |
| `POST /api/sesiones` con `admin@dev.local` | ✅ 200 y `Set-Cookie: sesion=<jwt>; Path=/; Max-Age=899; HttpOnly; Secure; SameSite=Lax`                                                |
| `POST /api/dev/tokens`                     | ✅ 200 con `iss=pingeso-identidad` y `aud=["pingeso"]`. Con `"duracion":"8h"`, `exp - iat = 28800` s; con `"25h"`, 422                  |
| `docker compose down`                      | ✅ Contenedores y red eliminados; el volumen `datos_postgres` se conserva                                                               |

## Checklist del Hito 1

| Ítem                                                                         | Estado | Comentario                                                                                                   |
| ---------------------------------------------------------------------------- | ------ | ------------------------------------------------------------------------------------------------------------ |
| Structs en `internal/api/` con `json`, validaciones y ejemplos               | ✅     |                                                                                                              |
| `contracts/openapi.yaml` 3.1 y tipos TypeScript con `pnpm codegen`           | ⚠️     | El contrato sí; los tipos TypeScript van en la Tanda 2                                                       |
| Endpoints en español y plural, con UUID                                      | ✅     | `/.well-known/jwks.json`, `/healthz` y `/readyz` siguen sus estándares                                       |
| Toda entrada validada en el contrato                                         | ✅     | `format`, `minLength`, `maxLength`, `pattern`, `maximum`, `maxItems`, `uniqueItems`, `minProperties`, `enum` |
| Cada endpoint declara su permiso                                             | ✅     | `x-permiso` y `security`; lo verifica `TestCadaOperacionDeclaraSuPermiso`                                    |
| Contrato publicado para los grupos dependientes                              | ⚠️     | Falta el merge a `main`, la release `hito-1` y avisar a los grupos                                           |
| Todos los endpoints responden con datos simulados                            | ✅     | Store en memoria con seed                                                                                    |
| `/api/docs` con `API_ENV=dev`                                                | ✅     |                                                                                                              |
| `/healthz` y `/readyz`                                                       | ✅     |                                                                                                              |
| `docker compose up` con postgres, migrate, backend y frontend en 5160 y 5161 | ⚠️     | postgres, migrate y backend levantan y responden en 5161. Falta el frontend en 5160 (Tanda 2)                |
| Contratos externos y prism/MSW                                               | —      | No aplica: este servicio no consume a otros                                                                  |
| Capas separadas, ningún handler llama al store                               | ✅     |                                                                                                              |
| Dueño único por dato                                                         | ✅     |                                                                                                              |
| Dependencias sin ciclos                                                      | ✅     | No depende de nadie                                                                                          |
| Estructura igual a la plantilla                                              | ⚠️     | Faltan `internal/repository` (Tanda 4) y `apps/web` (Tanda 2)                                                |
| `docs/modos-de-falla.md`                                                     | ✅     | 34 fallas, con su estado                                                                                     |
| Jobs `calidad`, `seguridad` e `imagen` en verde                              | ❓     | No hubo push. Ojo: Trivy y gitleaks no corrieron localmente                                                  |
| Paso de contrato: codegen sin diferencias y oasdiff                          | ✅     | Verificado en local                                                                                          |

## Pendientes para las tandas siguientes

- **Tanda 2 (front):** tipos TypeScript en `pnpm codegen`, microcomponente de login contra la
  cookie, regla de oxlint contra `localStorage` y `sessionStorage`, y servicio `frontend` en Compose.
- **Tanda 4 (persistencia):**
  - Store PostgreSQL con sqlc detrás de los mismos puertos.
  - Migración nueva que amplía el `CHECK` de `usuarios.estado` con `inactivo` y agrega un índice
    único sobre `lower(correo)`.
  - Tablas de roles, permisos, asignaciones, perfiles y auditoría.
  - Renovación y cambio de contraseña reales.
  - Regla del último administrador dentro de una transacción (MF-24).
- **Tanda 5:** matriz completa en `acceso_test.go`, logs por request con `X-Request-Id` y una
  prueba de que no se loguean secretos (MF-31).
- **Tanda 6:** e2e con Testcontainers y golden. Candidatos: registro concurrente con el mismo
  correo, último administrador concurrente, token vencido y prueba de tiempos del login (MF-12).
- **Tanda 3 (Compose, Dockerfiles y traspaso):**
  - Sacar goose de `go.mod` (`go install` en `Dockerfile.migrate`), lo que permite volver a
    `x/crypto v0.54.0`.
  - Montar la llave de firma en el contenedor para test y prod.
  - Renombrar "usuarios-roles" a "identidad" (log, `package.json`, Compose, título del contrato y
    módulo Go) junto con el traspaso a `PINGESO-CIVIL`.
- **Sin tanda asignada:** bloqueo por fuerza bruta (MF-11), procedimiento de rotación de llave
  (MF-16) y defensa CSRF verificada (MF-34).

## Preguntas abiertas

Las seis preguntas de la primera entrega quedaron respondidas (ver "Respuestas del equipo"). No
quedan preguntas abiertas de esta tanda. Lo único pendiente es pedirle al jefe de proyecto que
actualice la tabla del DAS con las dos desviaciones.
