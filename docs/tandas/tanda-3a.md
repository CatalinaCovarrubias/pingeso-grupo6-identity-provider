# Tanda 3A — Preparación del cierre del Hito 1

Rama `chore/tanda-3a-cierre-hito-1`, creada desde `main` (`2d5eccc`). No depende del PR #27 y no
comparte archivos con él. Sin tags, releases ni issue de entrega. Fecha: 25 de noviembre de 2026.

## Objetivo

Dejar listo todo lo del cierre del Hito 1 que no necesita el pipeline en verde: registrar las
decisiones, ordenar las dependencias de Go, completar los modos de falla, renombrar el módulo
tras el traspaso y preparar los textos de la entrega.

## Qué se hizo

| Commit                                                            | Qué incluye                                                                                                            |
| ----------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `chore(api): renombra el módulo Go…`                              | El módulo pasa a `github.com/PINGESO-CIVIL/usuarios-roles/apps/api`, con sus imports (16 archivos `.go`)               |
| `build(migrate): compila goose desde un módulo propio…`           | `apps/api/herramientas/goose/` con goose v3.24.1; sale del `go.mod` de la API; `Dockerfile.migrate` multi-arquitectura |
| `fix(docker): declara la sintaxis 1.7…`                           | `# syntax=docker/dockerfile:1.7` en `apps/api/Dockerfile`                                                              |
| `docs: actualiza el nombre del repositorio…`                      | Comandos `gh` de `docs/auditoria.md`                                                                                   |
| `docs: completa los modos de falla…` y `docs: corrige las filas…` | `docs/modos-de-falla.md`: verificación y estado, 15 fallas nuevas y la sección del frontend                            |
| `docs: registra las decisiones D1–D13…`                           | `docs/decisiones.md`                                                                                                   |
| `docs: borradores de la entrega del hito 1`                       | `docs/entregas/hito-1.md`: notas de la release, issue y aviso a los grupos                                             |
| `docs: resumen de la tanda 3A`                                    | Este documento                                                                                                         |

### Go y goose

- El `go.mod` de la API quedó con 5 dependencias directas y 5 indirectas: se fueron 58
  indirectas que entraban por goose, y `go.sum` pasó de 401 a 34 líneas.
- El binario de goose del módulo nuevo tiene las **mismas 58 dependencias en las mismas
  versiones** que el que se construía desde la API.
- `golang.org/x/crypto` se queda en **v0.58.0**: con v0.54.0 compila y pasa las pruebas, pero
  Trivy marca CVE-2026-56854 (HIGH, corregido en 0.55.0).
- `go install …@v3.24.1` se descartó porque la imagen quedaba con 28 vulnerabilidades
  HIGH/CRITICAL. El detalle está en [decisiones.md](../decisiones.md).

## Cómo probarlo

```bash
cd apps/api
gofmt -l .                      # no debe imprimir nada
go vet ./... && go test ./...
go mod verify && (cd herramientas/goose && go mod verify)
cd ../..

pnpm install --frozen-lockfile
pnpm verify
pnpm codegen && git diff --exit-code

docker build -f apps/api/Dockerfile -t api:ci .
docker build -f apps/api/Dockerfile.migrate -t migrate:ci .
docker build --platform linux/arm64 -f apps/api/Dockerfile.migrate -t migrate:arm64 .
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock aquasec/trivy:0.76.1 image \
  --severity HIGH,CRITICAL --ignore-unfixed --exit-code 1 migrate:ci

# Con una base vacía, sin tocar el volumen habitual
docker compose -p tanda3a up -d --build
docker compose -p tanda3a logs migrate     # "successfully migrated database to version: 1"
docker compose -p tanda3a down -v          # -v borra solo el volumen de este proyecto
```

En Git Bash, la ruta del socket se escribe `//var/run/docker.sock`.

### Resultados

| Verificación                                  | Resultado                                                                    |
| --------------------------------------------- | ---------------------------------------------------------------------------- |
| `gofmt`, `go vet`, `go test ./...`            | ✅                                                                           |
| `go mod verify` (API y goose) y `go mod tidy` | ✅ sin cambios                                                               |
| `pnpm verify`                                 | ✅                                                                           |
| `pnpm codegen` y oasdiff contra `main`        | ✅ sin diferencias                                                           |
| `docker build` de `api` y `migrate`           | ✅, también `migrate` para `linux/arm64`                                     |
| Trivy v0.76.1 sobre `migrate:ci` y `api:ci`   | ✅ 0 HIGH/CRITICAL                                                           |
| Trivy `fs` (incluye el `go.mod` de goose)     | ✅ 0 HIGH/CRITICAL                                                           |
| Hadolint, Dockle (`api`), govulncheck (API)   | ✅                                                                           |
| Gitleaks sobre los commits de la rama         | ✅ sin hallazgos                                                             |
| `docker compose up -d --build` desde cero     | ✅ `migrate` aplica la versión 1; repetido, no hay nada que migrar           |
| `/healthz`, `/readyz`, `/api/docs`, JWKS      | ✅ 200                                                                       |
| Login con usuario seed y token de prueba      | ✅ 200                                                                       |
| Unión con el PR #27 (`git merge-tree`)        | ✅ sin conflictos; compilan y pasan las pruebas de la API y del `spa-server` |

## Pendientes para la Tanda 3B

Con `main` en verde:

- [ ] **`ci.yml`:** agregar `go mod verify` y el escaneo del módulo `apps/api/herramientas/goose`.
      No se tocó ahora porque el #27 también cambia ese archivo. Ojo: tal como se compila hoy,
      `govulncheck` marca **3 vulnerabilidades alcanzables** en goose (brotli GO-2026-6888,
      klauspost/compress GO-2026-6865 y ch-go GO-2025-3603), así que ese paso quedaría en rojo.
      Compilando con las etiquetas
      `no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 no_vertica no_ydb` sale limpio:
      habría que aprobar ese cambio en `Dockerfile.migrate` y escanear con las mismas
      etiquetas. Trivy `fs` ya revisa su `go.mod`.
- [ ] Evaluar `trivy image` sobre `migrate` en el job `imagen` (hoy solo escanea `api` y, con el
      #27, `web`).
- [ ] **`spa-server`:** renombrar su módulo a
      `github.com/PINGESO-CIVIL/usuarios-roles/apps/web/spa-server` después del merge del #27.
- [ ] **`modos-de-falla.md`:** pasar de **I (#27)** a **P** las filas MF-50 a MF-57 al mergear.
- [ ] Publicar: tag `hito-1`, release con `openapi.yaml`, issue de entrega y aviso a los grupos,
      con los borradores de [hito-1.md](../entregas/hito-1.md). Revisar el checklist contra el
      commit etiquetado y completar los `⟨…⟩`.
- [ ] `sonar-project.properties` y el enlace de SonarQube, cuando el proyecto se vincule al
      repositorio de la organización.
- [ ] Actualizar el texto del issue #11, que habla de los grupos 3, 4 y 7: el aviso va a todos.
- [ ] Probar en Linux sin Docker Desktop la fila `host.docker.internal` de
      `docs/jwt-de-prueba.md`: el puerto se publica solo en `127.0.0.1`, así que probablemente
      no funcione.
- [ ] Cuando termine el congelamiento de Dependabot, decidir si vigila también el módulo de
      goose.

Para la Tanda 4:

- [ ] **Login y 72 bytes (MF-42):** bcrypt solo compara los primeros 72 bytes. Si una
      contraseña guardada ocupa exactamente 72, el login acepta cualquier texto agregado
      después. Con el contrato actual solo ocurre con letras de más de un byte. Se comprobó
      con `x/crypto` v0.58.0; hay que rechazar en el service las contraseñas de más de 72 bytes.
- [ ] Prueba en Go del límite de 72 bytes al crear una contraseña (MF-41).
- [ ] Probar que en `prod` no existen las cuentas seed (MF-46).

## Notas

- `docs/tandas/tanda-1.md` dice que `x/crypto` podría volver a v0.54.0 al sacar goose. No es
  así, por la vulnerabilidad citada; manda [decisiones.md](../decisiones.md).
- No se modificaron `README.md`, `sonar-project.properties`, `ci.yml`, `docker-compose.yaml` ni
  el contrato.
