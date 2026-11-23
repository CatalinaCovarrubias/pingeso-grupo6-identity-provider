# Tanda 2 — Frontend React, microcomponentes de sesión y Compose completo

Rama `feat/tanda-2-front`, creada desde `feat/tanda-1-contrato-mocks` porque el PR de la
Tanda 1 todavía no está mergeado. Cuando lo esté, se hace rebase sobre `main` y recién ahí
push y PR. Sin tags ni releases. Fecha: 23 de noviembre de 2026.

## Objetivo

Entregar el frontend del servicio de identidad según el DAS: dos microcomponentes React
autocontenidos (`inicio-sesion` y `sesion-actual`), cada uno con su cliente tipado desde el
contrato, una app de desarrollo que los monta con mocks MSW o contra el backend real, la
imagen `web` con su `spa-server`, el servicio `frontend` en Compose y el CI y Sonar al día.

## Qué se hizo

| Commit                                                 | Qué incluye                                                                                                                                                                     |
| ------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `chore(web): crea el workspace web con las versiones…` | `apps/web` con React, TypeScript y Vite en el catalog; el resto fijo en `apps/web/package.json`; `components.json` de shadcn (base-nova); Prettier con el plugin de Tailwind    |
| `feat(web): genera tipos y ejemplos de mocks…`         | `pnpm codegen` genera `src/api/esquema.d.ts` (openapi-typescript) y `src/mocks/ejemplos.gen.ts` (un ejemplo por esquema, sacado de los `examples` del contrato)                 |
| `feat(web): tokens de diseño y componentes base`       | `src/estilos/tema.css` (único CSS de tokens), `src/estilos/fuentes.css` (opcional), `src/components/ui` (shadcn) y los wrappers `sobre`, `boton`, `campo`, `aviso` e `insignia` |
| `feat(web): microcomponente inicio-sesion`             | Vista, `inicio-sesion.api.ts`, `use-inicio-sesion.ts`, más `lib/problemas.ts` (errores RFC 9457) y `lib/consultas.tsx` (`ConConsultas`)                                         |
| `feat(web): microcomponente sesion-actual`             | Vista, API y hook, con los estados cargando, con sesión, sin sesión y error                                                                                                     |
| `feat(web): app de desarrollo con mocks MSW…`          | `src/app-desarrollo.tsx`, handlers MSW, `public/mockServiceWorker.js`, `.env.development` (mocks) y `.env.backend` (backend real)                                               |
| `test(web): pruebas de los microcomponentes…`          | 23 pruebas de vitest con MSW, con cobertura lcov en `apps/web/coverage/lcov.info`                                                                                               |
| `chore: reglas de oxlint…`                             | `.oxlintrc.json`: `fetch` y `openapi-fetch` solo en `*.api.ts`; `localStorage` y `sessionStorage` prohibidos; un microcomponente no importa a otro                              |
| `fix(web): sube vite a 8.1.6 y vitest a 4.1.12…`       | Corrige el hallazgo HIGH de `pnpm audit` (ver desviaciones)                                                                                                                     |
| `feat(web): spa-server en Go…`                         | `apps/web/spa-server/` con `go.mod` propio y solo stdlib, más sus pruebas                                                                                                       |
| `build(web): imagen web en tres etapas…`               | `apps/web/Dockerfile` y el servicio `frontend` en `docker-compose.yaml`                                                                                                         |
| `test(web): prueba de humo con Playwright`             | `apps/web/e2e/humo.spec.ts` y `playwright.config.ts`                                                                                                                            |
| `ci: build, pruebas, audit y escaneo de la imagen web` | `ci.yml` y `sonar-project.properties`                                                                                                                                           |
| `style(web): ajustes de /audit y /polish…`             | Correcciones del pulido (ver más abajo)                                                                                                                                         |
| `docs: …`                                              | Este documento, `docs/diseno/PRODUCT.md`, `docs/diseno/DESIGN.md` y las capturas de `docs/tandas/tanda-2/`                                                                      |

### Estructura de `apps/web`

```
apps/web/
├── Dockerfile  components.json  index.html  package.json  playwright.config.ts  vite.config.ts
├── .env.development  .env.backend        # VITE_USAR_MOCKS (no es secreto)
├── e2e/humo.spec.ts
├── public/mockServiceWorker.js           # generado por msw init
├── scripts/generar-mocks.mjs
├── spa-server/  go.mod  main.go  main_test.go
└── src/
    ├── main.tsx  app-desarrollo.tsx      # app de desarrollo, no se publica
    ├── api/esquema.d.ts                  # generado
    ├── componentes/inicio-sesion/        # inicio-sesion.tsx, .api.ts, use-inicio-sesion.ts, index.ts, test
    ├── componentes/sesion-actual/        # ídem
    ├── components/ui/                    # generado por shadcn, no se edita
    ├── components/                       # wrappers: sobre, boton, campo, aviso, insignia
    ├── estilos/tema.css  fuentes.css
    ├── lib/                              # problemas.ts, consultas.tsx, utils.ts
    ├── mocks/                            # ejemplos.gen.ts (generado), handlers, navegador, servidor
    └── pruebas/configuracion.ts
```

## Props de cada microcomponente

`urlBaseApi` es el origen de la API **sin** `/api` (las rutas del contrato ya lo incluyen).
Con `""` o `window.location.origin` usa el mismo origen, que es lo recomendado: el
`spa-server`, o el proxy del host, reenvía `/api` y la cookie funciona sin CORS.

```tsx
import { InicioSesion, type UsuarioActual } from "./componentes/inicio-sesion"
import { SesionActual } from "./componentes/sesion-actual"

type PropsInicioSesion = {
  urlBaseApi: string
  alEntrar: (usuario: UsuarioActual) => void // el servicio aceptó y fijó la cookie
  correoInicial?: string
  titulo?: string // por defecto "Iniciar sesión"
  clienteConsultas?: QueryClient // opcional: compartir caché con el host
}

type PropsSesionActual = {
  urlBaseApi: string
  alSalir: () => void // la sesión se cerró (o ya no existía)
  usuarioInicial?: UsuarioActual // lo que entregó alEntrar: evita otro GET
  alNoHaySesion?: () => void // GET /api/sesiones/actual respondió 401
  clienteConsultas?: QueryClient
}
```

**`inicio-sesion`**

- Valida igual que el contrato: correo obligatorio, con formato email y máximo 254;
  contraseña de 1 a 72 caracteres. Además valida que la contraseña no pase de **72 bytes**
  (con `TextEncoder`), y desde 60 bytes muestra cuánto lleva.
- Mensajes de error:
  - 401: "El correo o la contraseña no son correctos". Es genérico, no revela si el correo
    existe. Además limpia la contraseña y le devuelve el foco.
  - 403: "Tu cuenta no está activa".
  - 422: "Revisa los datos marcados", y marca cada campo según `errors[].location`.
  - 5xx: "Algo falló de nuestro lado".
  - Sin respuesta: "No pudimos conectar con el servicio".
- Botón para mostrar y ocultar la contraseña, con `aria-pressed`.

**`sesion-actual`**

- Muestra el correo, el estado de la cuenta, los roles, los permisos (en monoespaciada), el
  vencimiento de la sesión en reloj de 24 horas y el ingreso anterior.
- Estados: cargando (skeleton con la misma caja), sin sesión (estado vacío), error (con
  "Reintentar") y con sesión.
- Un 401 manda aunque haya datos en caché. Si el `DELETE` responde 401, igual llama a
  `alSalir`, porque el resultado es el mismo: no hay sesión.

Ninguno de los dos depende de rutas, layout ni estado global, ni importa al otro. Los
coordina el host con los callbacks, como hace `app-desarrollo.tsx`.

## Decisiones aplicadas

| #   | Decisión                                                                                                                                                                                                                                                      | Dónde                                 |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------- |
| F1  | **react-query sin provider global**: cada microcomponente se envuelve en `ConConsultas`, que usa el `QueryClient` del host si llega por props o uno propio. Un provider anidado no pisa el del host                                                           | `src/lib/consultas.tsx`               |
| F2  | `fetch` solo en `*.api.ts`, vía `openapi-fetch` con `credentials: "include"`. Los errores se convierten en `ErrorApi` (estado + problem+json)                                                                                                                 | `*.api.ts`, `src/lib/problemas.ts`    |
| F3  | Mocks MSW **generados desde el contrato**: `scripts/generar-mocks.mjs` lee los `examples` y escribe `ejemplos.gen.ts`; los handlers solo agregan los escenarios (seed, 401, 403, 422, 500)                                                                    | `src/mocks/`                          |
| F4  | `VITE_USAR_MOCKS` decide entre MSW y el backend real. En `pnpm dev` vale `true`; en `pnpm dev:backend` y en la imagen, `false`. MSW se importa de forma dinámica y no entra al bundle de producción                                                           | `src/main.tsx`, `.env.*`              |
| F5  | **Tokens en un solo CSS** con `@theme` de Tailwind 4: colores OKLCH (escalas + tokens semánticos), escala tipográfica cerrada, dos pesos, radios, sombras de dos capas y duración del gesto. Todos son variables en `:root` que un cliente puede sobrescribir | `src/estilos/tema.css`                |
| F6  | **La tipografía es un token**: los microcomponentes usan `var(--fuente-texto)` y `var(--fuente-mono)` con fallback a `system-ui` y `ui-monospace`. Las fuentes Atkinson se cargan solo en la app de desarrollo y en el CSS opcional `fuentes.css`             | `tema.css`, `fuentes.css`, `main.tsx` |
| F7  | **Gesto "abrir el sobre"**: 180 ms, solo `transform` y `opacity` en CSS (`@starting-style`), sin animación con `prefers-reduced-motion`. El tramado es una máscara SVG teñida con tokens y nunca lleva texto encima                                           | `tema.css`, `components/sobre.tsx`    |
| F8  | El `spa-server` vive en `apps/web/spa-server/` (solo stdlib). Sirve `/srv` con `os.Root` (no se puede salir de la raíz), aplica fallback SPA, cachea `/assets/` como inmutable, responde `/healthz` y reenvía `/api` y `/.well-known` a `BACKEND_URL`         | `apps/web/spa-server/main.go`         |
| F9  | Healthcheck sin shell: `spa-server salud` consulta `/healthz`. Se usa en el `HEALTHCHECK` de la imagen y en Compose                                                                                                                                           | `Dockerfile`, `docker-compose.yaml`   |
| F10 | El proxy **conserva el `Host` del navegador**. Sin esto, el backend publicaba `https://backend:8080/schemas/…` en los `$schema` de sus respuestas                                                                                                             | `main.go`                             |
| F11 | Además de `nosniff` y `Referrer-Policy`, el `spa-server` envía `Content-Security-Policy: frame-ancestors 'none'` (contra el clickjacking en el login), propaga `X-Request-Id` y escribe logs JSON con `metodo`, `ruta`, `estado` y `duracion_ms`              | `main.go`                             |
| F12 | Las horas se muestran en reloj de 24 horas (`hourCycle: "h23"`): sin eso, `es-CL` daba "02:35 p. m."                                                                                                                                                          | `use-sesion-actual.ts`                |

### Dependencias fuera de la tabla del DAS

Todas van con versión exacta, en `apps/web/package.json` salvo que se indique otra cosa.

| Paquete                                                                   | Versión                          | Para qué                                                      |
| ------------------------------------------------------------------------- | -------------------------------- | ------------------------------------------------------------- |
| `@fontsource-variable/atkinson-hyperlegible-next` / `-mono`               | 5.3.0                            | Fuentes de la app de desarrollo y de `fuentes.css` (opcional) |
| `@vitest/coverage-v8`                                                     | 4.1.12                           | Cobertura lcov para Sonar                                     |
| `@testing-library/react` / `dom` / `user-event` / `jest-dom`              | 16.3.3 / 10.4.2 / 14.6.8 / 7.0.1 | Pruebas de componentes                                        |
| `jsdom`                                                                   | 30.1.3                           | Entorno DOM de vitest                                         |
| `@types/react` / `@types/react-dom` / `@types/node`                       | 19.2.18 / 19.2.7 / 22.20.5       | Tipos                                                         |
| `yaml`                                                                    | 2.9.1                            | Leer el contrato en `generar-mocks.mjs`                       |
| `class-variance-authority` / `clsx` / `tailwind-merge` / `tw-animate-css` | 0.7.1 / 2.1.1 / 3.7.0 / 1.4.0    | Dependencias que pide shadcn                                  |
| `lucide-react`                                                            | 1.70.0                           | Íconos (los de shadcn)                                        |
| CLI `shadcn` (no se instala; se corre con `pnpm dlx shadcn@4.23.0`)       | 4.23.0                           | Generar `src/components/ui`                                   |

## Cómo probarlo

```bash
pnpm install
pnpm exec lefthook install        # .npmrc tiene ignore-scripts: los hooks no se instalan solos

# App de desarrollo con mocks MSW (http://localhost:5160)
pnpm --filter web dev
# Contra el backend local en 127.0.0.1:5161 (proxy de Vite para /api y /.well-known)
pnpm --filter web dev:backend

# Verificaciones
pnpm verify
pnpm typecheck
pnpm test                          # vitest (con lcov) + go test de la API + go test del spa-server
pnpm build:web
pnpm codegen && git diff --exit-code
pnpm audit --audit-level high

# Prueba de humo (la primera vez: pnpm --filter web exec playwright install chromium)
pnpm test:e2e:web                                   # contra la app con MSW
CAPTURAS=1 pnpm test:e2e:web                        # además regenera docs/tandas/tanda-2/*.png
E2E_URL=http://127.0.0.1:5160 pnpm test:e2e:web     # contra el Compose (backend real)

# Imagen y Compose
docker build -f apps/web/Dockerfile -t web:local .
docker compose up -d --build       # frontend en 127.0.0.1:5160, backend en 127.0.0.1:5161
docker compose down
```

En PowerShell, las variables se pasan como `$env:E2E_URL='http://127.0.0.1:5160'; pnpm test:e2e:web`.

Cuentas de prueba (también aparecen en la app de desarrollo): `admin@dev.local` /
`admin-dev-local`, `operador@dev.local` / `operador-dev-local` y `usuario@dev.local` /
`usuario-dev-local`. Solo en MSW existen además `bloqueado@dev.local` /
`bloqueado-dev-local` (403) y `falla@dev.local` (500).

### Resultados de esta tanda

| Verificación                                                 | Resultado                                                                                                                                                                                  |
| ------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `pnpm install --frozen-lockfile`                             | ✅                                                                                                                                                                                         |
| `pnpm verify` (Prettier, oxlint, go vet de API y spa-server) | ✅ (ver la nota sobre Prettier en pendientes)                                                                                                                                              |
| `pnpm typecheck`                                             | ✅                                                                                                                                                                                         |
| `pnpm test`                                                  | ✅ 23 pruebas web (98,8 % de líneas), API y spa-server                                                                                                                                     |
| Lógica de fechas con `TZ=UTC` y `TZ=Pacific/Auckland`        | ✅                                                                                                                                                                                         |
| `pnpm build:web`                                             | ✅                                                                                                                                                                                         |
| `pnpm codegen` + `git diff --exit-code`                      | ✅ sin diferencias (corrido dos veces: es determinista)                                                                                                                                    |
| `go test ./...` de la API                                    | ✅ sin cambios en el backend                                                                                                                                                               |
| `pnpm audit --audit-level high`                              | ✅ después de subir vite y vitest                                                                                                                                                          |
| Reglas de oxlint                                             | ✅ probadas con archivos de ejemplo: `fetch`, `globalThis.fetch`, `localStorage`, `window.sessionStorage`, `openapi-fetch` fuera de `*.api.ts`, e imports `../otro` y `@/componentes/otro` |
| Playwright con MSW                                           | ✅ 5/5                                                                                                                                                                                     |
| Hadolint v2.15.1 sobre `apps/web/Dockerfile`                 | ✅ solo DL3066 (info), igual que la API                                                                                                                                                    |
| Trivy v0.76.1 `image web:local` y `fs`                       | ✅ 0 HIGH/CRITICAL                                                                                                                                                                         |
| govulncheck del spa-server                                   | ✅ sin vulnerabilidades                                                                                                                                                                    |

### Docker Compose (23 de noviembre)

| Prueba contra `127.0.0.1:5160`                          | Resultado                                                                                                                                       |
| ------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| `docker compose up -d --build`                          | ✅ postgres, migrate, backend y frontend; frontend `healthy` con `spa-server salud`                                                             |
| `GET /healthz` y `/` con cabeceras                      | ✅ `nosniff`, `Referrer-Policy`, `frame-ancestors 'none'`, `index.html` con `no-cache`                                                          |
| Login con `admin@dev.local` a través del spa-server     | ✅ 200 y `Set-Cookie: sesion=…; HttpOnly; Secure; SameSite=Lax`                                                                                 |
| Sesión actual con la cookie, JWKS, 401 y logout         | ✅ 200, 200, 401 genérico y 204                                                                                                                 |
| Playwright en el navegador (`E2E_URL`)                  | ✅ 4/4 (la cuenta bloqueada se omite porque solo existe en MSW); la cookie llega `HttpOnly`, `Secure` y `Lax`, y la sesión sobrevive a recargar |
| 403: `PATCH /api/usuarios/{id}` con `estado: bloqueado` | ✅ el login de `usuario@dev.local` da 403 y la UI muestra "Tu cuenta no está activa"; con una contraseña mala sigue siendo 401                  |
| `docker compose down`                                   | ✅                                                                                                                                              |

## Checklist del Hito 1

| Ítem                                                                                            | Estado   | Nota                                                                                              |
| ----------------------------------------------------------------------------------------------- | -------- | ------------------------------------------------------------------------------------------------- |
| `openapi.yaml` y **tipos TypeScript** generados con `pnpm codegen` y commiteados                | ✅       | `apps/web/src/api/esquema.d.ts`                                                                   |
| `docker compose up -d --build` levanta postgres, migrate, backend y **frontend** en 5160 y 5161 | ✅       | Cierra el pendiente de la Tanda 1                                                                 |
| Estructura de repositorio igual a la plantilla                                                  | ⚠️       | `apps/web` ya existe (con `src/componentes/` según DAS §3). Falta `internal/repository` (Tanda 4) |
| Consumidor contra `prism mock` o handlers MSW                                                   | ✅       | Los microcomponentes trabajan contra handlers MSW generados desde el contrato                     |
| Jobs `calidad`, `seguridad` e `imagen` en verde                                                 | ⏳       | Se verá al abrir el PR (después del rebase)                                                       |
| Paso de contrato: `pnpm codegen` sin diferencias                                                | ✅ local | Ahora también cubre los tipos TS y los ejemplos MSW                                               |
| Contrato publicado para los grupos dependientes                                                 | ⚠️       | Sigue pendiente de la Tanda 1 (merge, release `hito-1` y aviso)                                   |

## Desviaciones del DAS

Hay que pedirle al jefe de proyecto que las apruebe o que actualice el DAS.

1. **vite 8.1.6 y vitest / @vitest/coverage-v8 4.1.12**, en vez de 8.1.5 y 4.1.10. Con las
   versiones del DAS, `pnpm audit --audit-level high` falla por GHSA-587h-7wmj-h5pm (high,
   `launch-editor` del servidor de desarrollo en Windows), más 6 moderados y 1 bajo. Ninguno
   llega a la imagen. El DAS pide que subir versiones sea un PR propio; el equipo decidió
   subirlas aquí para que el CI pase, y hay que actualizar la tabla del DAS.
2. **Dependencias en `apps/web/package.json`**: §9 dice "versiones en el catalog, nunca
   sueltas", pero la tabla de §3 las fija en `apps/web/package.json`. Se siguió la tabla: en
   el catalog solo quedaron react, react-dom, typescript y vite.
3. El `spa-server` también reenvía **`/.well-known`** (el DAS menciona solo `/api`), para que
   el JWKS quede en el mismo origen.
4. **Sin Prism**: está en la tabla de herramientas, pero los mocks del front son MSW.
   **Sin react-router**: la app de desarrollo no tiene rutas.
5. **Imagen de Go del spa-server**: `golang:1.26.11-alpine3.23`, la misma que ya usa la API
   desde la Tanda 1 (el DAS dice 1.26.5).
6. **Import generado por shadcn**: la CLI 4.23.0 escribió `import { cn } from "cn"` (y agregó
   un paquete npm `cn`) en vez de `@/lib/utils`. Se quitó el paquete y se corrigió solo esa
   línea en los archivos generados de `src/components/ui`. Si se regeneran, hay que repetir
   el reemplazo.
7. **Dockle no corre sobre `web:ci`**: el workflow del DAS lo aplica solo a `api:ci`. Se dejó
   igual.

## Qué cambió con /audit y /polish

La auditoría técnica del skill (detector y contraste calculado sobre los tokens) dio
**17/20** antes del pulido: accesibilidad 3, rendimiento 4, theming 3, responsive 4 e
integridad 3. El detector automático no encontró problemas. Los textos pasan AA con
holgura (principal 17,7:1, secundario 6,8:1, botón 7,3:1, avisos sobre 8:1).

| Prioridad | Hallazgo                                                                                                      | Cambio                                                                |
| --------- | ------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| P1        | El borde de los inputs tenía 2,57:1 contra la tarjeta; WCAG 1.4.11 pide 3:1 para límites de controles         | `--input` pasó de neutro-400 a neutro-500 (4,75:1)                    |
| P2        | Con el sobre abierto, la solapa recogida dejaba una franja celeste vacía de unos 26 px                        | La banda abierta queda transparente y el padding superior se compensa |
| P2        | El skeleton seguía pulsando con `prefers-reduced-motion`                                                      | `motion-reduce:animate-none` en `sesion-actual`                       |
| P2        | Dos valores salían directo de la escala (`--tinta-700`, `--tinta-100`) y el cliente no los podía sobrescribir | Nuevos tokens semánticos `--primary-hover` y `--accent-border`        |
| P3        | La app de desarrollo no tenía `h1`                                                                            | El título de la cabecera pasó a `h1`                                  |
| P3        | Un `<p role="status">` (lo marcó oxlint)                                                                      | Se cambió por `<output>`                                              |

Después del pulido: **19/20**. Queda abierto (P3) que el spinner del botón sigue girando con
movimiento reducido, porque es el único indicador de progreso. El modo oscuro no existe
todavía, por decisión: los tokens semánticos ya están separados de las escalas para
agregarlo. Se revisaron en el navegador los estados de carga, vacío, error, credenciales
malas, cuenta bloqueada y con sesión, en escritorio y en 390 px. Las capturas están en
`docs/tandas/tanda-2/`.

## Pendientes para las tandas siguientes

- **Tanda 4 (backend):** el contrato limita la contraseña a 72 **caracteres**
  (`maxLength: 72`), pero bcrypt y el service cuentan 72 **bytes**
  (`apps/api/internal/service/comunes.go`, `len(contrasena)`). Una contraseña nueva con
  tildes o ñ puede pasar el contrato y recibir 422 del service. El front ya valida los bytes.
- **Tanda 4 (backend):** agregar al seed un usuario bloqueado para probar el 403 sin `PATCH`.
- **Tanda 3:** publicar el paquete npm de los microcomponentes (`index.ts` de cada
  microcomponente ya exporta lo público) y declarar qué versión del contrato consume.
- **LPS:** microcomponentes de gestión de usuarios y de asignación de roles.
- **Jefe de proyecto:** aprobar las desviaciones de arriba y actualizar la tabla de versiones
  del DAS (vite 8.1.6, vitest 4.1.12).
- **Prettier en local:** `prettier --check .` revisa todo lo que hay en disco. Si alguien
  tiene carpetas locales excluidas con `.git/info/exclude` (no con `.gitignore`), `pnpm verify`
  falla en su máquina aunque en CI pase. Conviene excluirlas también en `.prettierignore` o
  moverlas fuera del repo.
- **Publicación de esta rama:** rebase sobre `main` cuando se mergee la Tanda 1, push, PR y
  revisar `gh pr checks`. El CI nuevo (Playwright, audit, `web:ci`) todavía no corrió en
  GitHub.

## Preguntas abiertas

1. ¿El jefe de proyecto acepta subir vite y vitest en este PR, o hay que separarlo en un PR
   propio antes del merge?
2. ¿Los grupos cliente van a montar los microcomponentes en el mismo origen que el servicio
   (con un proxy como el `spa-server`)? Si no, hay que definir CORS con credenciales en el
   backend y revisar `SameSite=Lax` para el caso entre sitios.
3. ¿Se quiere Dockle también sobre `web:ci`?
