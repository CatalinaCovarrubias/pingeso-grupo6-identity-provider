# Decisiones y desviaciones del DAS

Registro de las decisiones del equipo (D1–D13, planteadas en [auditoria.md](auditoria.md)) y de
todo lo que se aparta del documento **DAS — Desarrollo de Microservicios**. El DAS dice que
cambiar una pieza del stack requiere la aprobación del jefe de proyecto: mientras no quede
registrada, cada fila está **pendiente de aprobación**.

Última actualización: 25 de noviembre de 2026.

## Decisiones D1–D13

| ID  | Decisión                                                                                                                                                                  | Motivo                                                                                                                                          | Ante el jefe de proyecto |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------ |
| D1  | Rutas `/api/...` en español y plural, y JWKS en `/.well-known/jwks.json`                                                                                                  | Sigue los ejemplos del DAS y lo que esperan los otros grupos. Una `/api/v2` solo se crearía ante un cambio incompatible                         | Pendiente de aprobación  |
| D2  | Tokens firmados con RS256 (`golang-jwt/jwt/v5` v5.3.1) y llave RSA de 2048 bits o más                                                                                     | El DAS admite EdDSA o RS256; RS256 es el más compatible con las herramientas de los consumidores                                                | Pendiente de aprobación  |
| D3  | El token viaja en la cookie `sesion` (`HttpOnly`, `Secure`, `SameSite=Lax`) y también se acepta `Authorization: Bearer`. El login no lo devuelve en el body               | La cookie protege al navegador; entre servicios el JWT se reenvía por cabecera                                                                  | Pendiente de aprobación  |
| D4  | Token de 15 minutos (`JWT_TTL`) y renovación simulada hasta la Tanda 4. Se acepta que un token emitido siga valiendo hasta su `exp` tras un bloqueo o un cierre de sesión | Evita que los demás servicios consulten a identidad en cada petición, con una ventana corta                                                     | Pendiente de aprobación  |
| D5  | Los perfiles son parte de este servicio (`GET` y `PUT /api/usuarios/{id}/perfil`)                                                                                         | Están en el alcance del README y del borrador de contrato                                                                                       | Pendiente de aprobación  |
| D6  | Se mantiene Go 1.26.11 (`toolchain` e imagen de build) y se pide actualizar el DAS, que fija 1.26.5                                                                       | Es un parche más nuevo de la misma versión menor y el repositorio ya lo usaba                                                                   | Pendiente de aprobación  |
| D7  | goose v3.24.1 se compila desde un módulo propio, `apps/api/herramientas/goose/`, y sale del `go.mod` de la API                                                            | `go install …@v3.24.1`, como dice el DAS, deja la imagen `migrate` con vulnerabilidades (tabla de abajo)                                        | Pendiente de aprobación  |
| D8  | Frontend en React 19.2.8                                                                                                                                                  | Es lo que fija el DAS. No había frontend previo, así que no hay migración                                                                       | Pendiente de aprobación  |
| D9  | Usuarios seed en memoria y `POST /api/dev/tokens` solo con `API_ENV=dev`                                                                                                  | Los otros grupos prueban sus permisos con tokens reales antes de que haya base de datos, sin exponer el endpoint fuera de desarrollo            | Pendiente de aprobación  |
| D10 | Paquete npm `@pingeso-civil/usuarios-roles-ui` en GitHub Packages                                                                                                         | El DAS §8 pide publicar ahí los microcomponentes y el repositorio ya está en la organización `PINGESO-CIVIL`                                    | Pendiente de aprobación  |
| D11 | Los microcomponentes viven en `src/componentes/<kebab-case>/`                                                                                                             | El DAS se contradice: el árbol de §3 dice `src/features/`, y la sección de microcomponentes, `src/componentes/`. Se sigue la sección específica | Pendiente de aprobación  |
| D12 | Permisos `usuarios:administrar`, `roles:administrar` y `perfiles:operar`; roles `administrador`, `operador` y `usuario`                                                   | Usa la forma `<recurso>:administrar` y `<recurso>:operar` del DAS. Los demás servicios autorizan por permiso                                    | Pendiente de aprobación  |
| D13 | Claims `iss`, `sub`, `aud`, `exp`, `iat`, `nbf`, `jti`, `roles` y `permisos`, con `kid` en el header                                                                      | Cada servicio puede validar y autorizar solo con el token. No lleva correo ni otros datos personales                                            | Pendiente de aprobación  |

### Por qué goose tiene su propio módulo (D7)

Escaneo con Trivy v0.76.1 (`HIGH` y `CRITICAL` con corrección disponible) de la imagen
`migrate`, el 25 de noviembre:

| Cómo se obtiene goose                                        | Versión de goose | HIGH/CRITICAL            |
| ------------------------------------------------------------ | ---------------- | ------------------------ |
| Desde el `go.mod` de la API (situación anterior)             | v3.24.1          | 0                        |
| `go install` (lo que dice el DAS)                            | v3.24.1          | **28**                   |
| `go install` excluyendo con etiquetas los otros motores      | v3.24.1          | **15** (2 críticas, pgx) |
| **Módulo propio** (elegida)                                  | v3.24.1          | 0                        |
| `go install` con una versión nueva (publicada ese mismo día) | v3.30.2          | 0                        |

`go install` usa las dependencias que goose declaraba al publicarse (pgx 5.7.1, `x/crypto`
0.31.0, grpc 1.62.1). El módulo propio mantiene goose en la versión del DAS y conserva las
dependencias corregidas: el binario resultante tiene las mismas 58 dependencias, en las mismas
versiones, que el que se construía antes. No se usaron excepciones de Trivy, porque el DAS no
las admite cuando hay corrección.

## Desviaciones del DAS

| Desviación                                                                                   | Motivo                                                                                                                                                                                 | Dónde  | Ante el jefe de proyecto |
| -------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------------------------ |
| `golang.org/x/crypto` **v0.58.0** (DAS: v0.54.0)                                             | v0.54.0 tiene **CVE-2026-56854** (HIGH, corregido en 0.55.0): Trivy lo marca en `go.mod` y en la imagen de la API. Compila y pasa las pruebas, pero rompe el umbral de 0 HIGH/CRITICAL | `main` | Pendiente de aprobación  |
| `github.com/google/uuid` v1.6.0 como dependencia directa (no está en la tabla)               | Identificadores y `jti`. Ya estaba en el grafo porque Huma la requiere                                                                                                                 | `main` | Pendiente de aprobación  |
| Go **1.26.11** (DAS: 1.26.5), también en las imágenes de build (D6)                          | Parche más nuevo, ya en uso en el repositorio                                                                                                                                          | `main` | Pendiente de aprobación  |
| goose con `go build` desde un módulo propio, no con `go install` (D7)                        | Con `go install` la imagen queda con vulnerabilidades                                                                                                                                  | `main` | Pendiente de aprobación  |
| **vite 8.1.6** (DAS: 8.1.5) y **vitest** y `@vitest/coverage-v8` **4.1.12** (DAS: 4.1.10)    | Con las versiones del DAS, `pnpm audit --audit-level high` falla (GHSA-587h-7wmj-h5pm, entre otros)                                                                                    | PR #27 | Pendiente de aprobación  |
| Dependencias del frontend fuera de la tabla (lista de abajo)                                 | Las necesitan las pruebas, shadcn, la generación de mocks y las fuentes                                                                                                                | PR #27 | Pendiente de aprobación  |
| Dependencias en `apps/web/package.json` y no todas en el catalog                             | El §9 pide el catalog, pero la tabla del §3 las fija en `apps/web/package.json`. En el catalog quedan react, react-dom, typescript y vite                                              | PR #27 | Pendiente de aprobación  |
| El `spa-server` reenvía también `/.well-known` (el DAS menciona solo `/api`)                 | El JWKS queda en el mismo origen que la aplicación                                                                                                                                     | PR #27 | Pendiente de aprobación  |
| Sin Prism (`@stoplight/prism-cli`) ni react-router                                           | Los mocks del frontend son MSW, generados desde el contrato, y la app de desarrollo no tiene rutas                                                                                     | PR #27 | Pendiente de aprobación  |
| Una línea corregida en `src/components/ui/` (generado por shadcn, que el DAS pide no editar) | La CLI escribió `import { cn } from "cn"` en vez de `@/lib/utils`                                                                                                                      | PR #27 | Pendiente de aprobación  |
| _Condicional, no aplicada:_ gitleaks por su CLI en lugar de `gitleaks/gitleaks-action`       | La action exige `GITLEAKS_LICENSE` en organizaciones y falla desde el traspaso. Solo se haría si no se consigue la licencia                                                            | —      | Pendiente de aprobación  |

### Dependencias del frontend fuera de la tabla del DAS

Todas con versión exacta (PR #27).

| Paquete                                                                 | Versión                        |
| ----------------------------------------------------------------------- | ------------------------------ |
| `@fontsource-variable/atkinson-hyperlegible-next` y `-mono`             | 5.3.0                          |
| `@vitest/coverage-v8`                                                   | 4.1.12                         |
| `@testing-library/react`, `dom`, `user-event` y `jest-dom`              | 16.3.3, 10.4.2, 14.6.8 y 7.0.1 |
| `jsdom`                                                                 | 30.1.3                         |
| `@types/react`, `@types/react-dom` y `@types/node`                      | 19.2.18, 19.2.7 y 22.20.5      |
| `yaml`                                                                  | 2.9.1                          |
| `class-variance-authority`, `clsx`, `tailwind-merge` y `tw-animate-css` | 0.7.1, 2.1.1, 3.7.0 y 1.4.0    |
| `lucide-react`                                                          | 1.70.0                         |
| CLI `shadcn` (no se instala: `pnpm dlx shadcn@4.23.0`)                  | 4.23.0                         |

## Otras decisiones del equipo

| Decisión                                                                             | Motivo                                                                                              | Ante el jefe de proyecto |
| ------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------- | ------------------------ |
| **Dependabot congelado hasta el 4 de diciembre**: sus PR no se mergean ni se cierran | Cambian versiones que fija el DAS y, en el caso de Huma, el contrato generado, a días de la entrega | Pendiente de aprobación  |
