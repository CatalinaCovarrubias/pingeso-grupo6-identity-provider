# Hito 1 — borradores de la entrega

Tres textos listos para la Tanda 3B: las notas de la release, el issue de entrega y el aviso a
los otros grupos. **Nada de esto está publicado**: no hay tag, release ni issue de entrega.

Antes de publicar:

1. El PR #27 (frontend) y el PR de la Tanda 3A están en `main`, con el pipeline en verde.
2. Reemplazar cada `⟨…⟩`.
3. Revisar el checklist contra el commit que se va a etiquetar.

Los enlaces con `/hito-1/` funcionan recién cuando exista el tag.

---

## a) Notas de la release `hito-1`

**Título:** Hito 1 — contrato, mocks y frontend de sesión

Primera entrega del servicio de identidad: el contrato OpenAPI está publicado, todos los
endpoints responden con datos simulados y emite JWT reales para que los demás servicios
integren antes de que exista la persistencia.

### Qué incluye

- **Contrato** OpenAPI 3.1, versión 0.2.0: 24 operaciones, cada una con su permiso
  (`x-permiso`), entradas validadas y errores RFC 9457.
- **Sesión:** login, logout, sesión actual, renovación y cambio de contraseña (estos dos,
  simulados), con la cookie `sesion` (`HttpOnly`, `Secure`, `SameSite=Lax`) o
  `Authorization: Bearer`.
- **JWT RS256** con JWKS público en `/.well-known/jwks.json`.
- **Usuarios, roles, permisos y perfiles** con datos en memoria y usuarios de prueba.
- **Tokens de prueba** para otros servicios con `POST /api/dev/tokens`, solo en desarrollo.
- **Frontend** (PR #27): microcomponentes `inicio-sesion` y `sesion-actual`, app de desarrollo
  con mocks MSW, imagen `web` y servicio `frontend` en Docker Compose.
- **Preparación del cierre:** módulo Go renombrado a `github.com/PINGESO-CIVIL/usuarios-roles`,
  goose en un módulo propio, modos de falla y registro de decisiones.

### Todavía no incluye (Hito 2)

Persistencia en PostgreSQL, auditoría de cambios, matriz de acceso por rol, logs por petición y
pruebas e2e. Los datos se pierden al reiniciar el servicio.

### Enlaces

| Qué                       | Enlace                                                                                                 |
| ------------------------- | ------------------------------------------------------------------------------------------------------ |
| Pipeline de este commit   | ⟨URL del run en verde⟩                                                                                 |
| Contrato                  | <https://github.com/PINGESO-CIVIL/usuarios-roles/blob/hito-1/contracts/openapi.yaml> (también adjunto) |
| Guía de JWT de prueba     | <https://github.com/PINGESO-CIVIL/usuarios-roles/blob/hito-1/docs/jwt-de-prueba.md>                    |
| Modos de falla            | <https://github.com/PINGESO-CIVIL/usuarios-roles/blob/hito-1/docs/modos-de-falla.md>                   |
| Decisiones y desviaciones | <https://github.com/PINGESO-CIVIL/usuarios-roles/blob/hito-1/docs/decisiones.md>                       |
| SonarQube                 | ⟨URL del proyecto una vez vinculado al repositorio de la organización⟩                                 |

El Hito 1 no pide SonarQube como evidencia; el enlace va como referencia.

**Archivo adjunto:** `contracts/openapi.yaml`.

---

## b) Issue "Entrega Hito 1"

No hay plantilla de este issue en el repositorio ni en la organización: se usa el checklist del
documento de hitos. Los casilleros reflejan el estado al **25 de noviembre** (`main` en
`2d5eccc` más la Tanda 3A). Lo que depende del PR #27 o del pipeline queda sin marcar.

**Título:** Entrega Hito 1

Release: ⟨URL de la release `hito-1`⟩ · Commit: ⟨SHA⟩ · Pipeline: ⟨URL del run en verde⟩

### Contrato API · DAS §2

- [x] Structs de request y response en `internal/api/`, con tags `json`, validaciones y ejemplos.
- [ ] `contracts/openapi.yaml` en OpenAPI 3.1 y los tipos TypeScript, generados con
      `pnpm codegen` y commiteados. _El contrato ya está en `main`; los tipos TypeScript llegan
      con el PR #27._
- [x] Endpoints en español y plural, como `/api/usuarios`, con identificadores UUID.
- [x] Toda entrada validada en el contrato: tipos, largos, máximos y patrones.
- [x] Cada endpoint declara su permiso, por ejemplo `<recurso>:operar`. _24 de 24 operaciones._
- [ ] Contrato publicado para los grupos que dependen del servicio, según el documento de
      componentes de la sección. _Falta la release y el aviso._

### Endpoints con datos simulados · DAS §2, §4 y §8

- [x] Todos los endpoints del contrato responden con datos simulados que cumplen el contrato.
- [x] Documentación interactiva en `/api/docs` con `API_ENV=dev`.
- [x] `GET /healthz` y `GET /readyz` responden.
- [ ] `docker compose up -d --build` levanta postgres, migrate, backend y frontend en los
      puertos `51N0` y `51N1` del servicio. _postgres, migrate y backend (5161) están en `main`;
      el frontend (5160) llega con el PR #27._

### Consumo de otros servicios · DAS §2

- [ ] Contrato de cada proveedor copiado en `contracts/externos/<servicio>.yaml`, con su cliente
      generado desde ahí. **No aplica:** el servicio de identidad no depende de ningún otro.
- [ ] Mientras el proveedor no tenga lógica, el consumidor trabaja contra `prism mock` o
      handlers MSW. **No aplica:** el servicio de identidad no depende de ningún otro.

### Arquitectura, revisión con nota · DAS §2 y §3

- [x] Capas Handler, Service y Store separadas: ningún handler llama al store.
- [x] Dueño único por dato: el servicio no escribe datos de otro ni se conecta a su base.
- [x] Dependencias en un solo sentido, sin ciclos y coherentes con el documento de componentes.
      _Identidad no depende de otros servicios; los demás dependen de él._
- [ ] Estructura de repositorio igual a la de la plantilla. _`apps/web` llega con el PR #27;
      `internal/repository` (sqlc) es del Hito 2._
- [x] `docs/modos-de-falla.md` con cada falla posible, qué la impide y cómo se verifica.

### Pipeline · DAS §6 y §7

- [ ] Jobs `calidad`, `seguridad` e `imagen` en verde sobre el commit entregado. _Bloqueado
      desde el traspaso: licencia de gitleaks y vínculo de SonarQube._
- [ ] Paso de contrato en verde: `pnpm codegen` sin diferencias y oasdiff sin cambios
      incompatibles contra `main`. _Sin diferencias en local; falta el run sobre el commit
      entregado._

---

## c) Aviso a los grupos del curso

> **Asunto:** Servicio de identidad (grupo 6): contrato y tokens de prueba disponibles
>
> Hola a todos los grupos:
>
> El servicio de identidad es la base común del curso y ya pueden integrarse con él. Los tokens
> son reales (JWT RS256, validables con el JWKS); lo simulado es el almacenamiento, que vive en
> memoria.
>
> **1. Contrato**
> <https://github.com/PINGESO-CIVIL/usuarios-roles/blob/hito-1/contracts/openapi.yaml>
> Cópienlo en su repositorio como `contracts/externos/identidad.yaml` y generen desde ahí su
> cliente.
>
> **2. Levantarlo en su máquina**
>
> ```bash
> git clone https://github.com/PINGESO-CIVIL/usuarios-roles.git
> cd usuarios-roles
> cp .env.example .env   # completen POSTGRES_PASSWORD y DATABASE_URL
>                        # y dejen vacía JWT_LLAVE_PRIVADA_ARCHIVO
> docker compose up -d --build
> ```
>
> La API queda en `http://localhost:5161` y la documentación en
> `http://localhost:5161/api/docs`.
>
> **3. Pedir un token de prueba** (solo con `API_ENV=dev`)
>
> ```bash
> curl -s -H 'Content-Type: application/json' \
>   -d '{"roles":["operador"],"permisos_extra":["reservas:operar"],"duracion":"8h"}' \
>   http://localhost:5161/api/dev/tokens
> ```
>
> En `permisos_extra` pongan los permisos de **su** servicio, con la forma `<recurso>:<accion>`.
> Usen el token en `Authorization: Bearer <token>`. También hay usuarios de prueba para el login
> (`admin@dev.local`, `operador@dev.local` y `usuario@dev.local`); las contraseñas están en la
> guía.
>
> **4. Validar los tokens: `AUTH_JWKS_URL`**
>
> | Su servicio corre en                              | `AUTH_JWKS_URL`                                          |
> | ------------------------------------------------- | -------------------------------------------------------- |
> | Su máquina, sin Docker                            | `http://localhost:5161/.well-known/jwks.json`            |
> | Docker Desktop, con identidad en la misma máquina | `http://host.docker.internal:5161/.well-known/jwks.json` |
> | La misma red de Compose que identidad             | `http://backend:8080/.well-known/jwks.json`              |
>
> Con Docker en Linux sin Docker Desktop, la segunda fila no debería funcionar, porque el puerto
> se publica solo en `127.0.0.1`: usen la misma red de Compose.
>
> Comprueben firma RS256, `iss` = `pingeso-identidad`, `aud` = `pingeso` y `exp`. **Autoricen
> por permiso, no por nombre de rol.**
>
> **Para tener en cuenta**
>
> - Sin una llave fija, cada reinicio cambia la llave y los tokens anteriores dejan de valer:
>   pidan uno nuevo.
> - El token del login dura 15 minutos; los de prueba, hasta 24 horas.
>
> Guía completa, con un ejemplo en Go:
> <https://github.com/PINGESO-CIVIL/usuarios-roles/blob/hito-1/docs/jwt-de-prueba.md>
>
> Cualquier problema, abran un issue en el repositorio.
>
> Grupo 6
