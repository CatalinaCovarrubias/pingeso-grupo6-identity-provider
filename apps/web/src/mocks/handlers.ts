// Handlers MSW de la API de sesiones. Parten de los ejemplos del contrato
// (ejemplos.gen.ts) y simulan los mismos escenarios que el backend en modo dev:
// los usuarios seed, un 401 genérico, un 403 por cuenta bloqueada y un 422 por datos inválidos.
import { http, HttpResponse } from "msw"
import type { components } from "@/api/esquema"
import { ejemplos } from "./ejemplos.gen"

type UsuarioActual = components["schemas"]["UsuarioActual"]
type Problema = components["schemas"]["ErrorModel"]

type Cuenta = {
  contrasena: string
  usuario: Omit<UsuarioActual, "expira_en">
}

const base = ejemplos.UsuarioActual

/** Mismos usuarios seed que docs/jwt-de-prueba.md, más uno bloqueado y uno que falla. */
export const cuentas: Record<string, Cuenta> = {
  "admin@dev.local": {
    contrasena: "admin-dev-local",
    usuario: { ...base, ultimo_acceso: "2026-01-14T09:30:00Z" },
  },
  "operador@dev.local": {
    contrasena: "operador-dev-local",
    usuario: {
      ...base,
      id: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
      correo: "operador@dev.local",
      roles: [{ id: "22222222-2222-4222-8222-222222222222", nombre: "operador" }],
      permisos: ["perfiles:operar"],
    },
  },
  "usuario@dev.local": {
    contrasena: "usuario-dev-local",
    usuario: {
      ...base,
      id: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
      correo: "usuario@dev.local",
      roles: [{ id: "33333333-3333-4333-8333-333333333333", nombre: "usuario" }],
      permisos: [],
    },
  },
  "bloqueado@dev.local": {
    contrasena: "bloqueado-dev-local",
    usuario: {
      ...base,
      id: "dddddddd-dddd-4ddd-8ddd-dddddddddddd",
      correo: "bloqueado@dev.local",
      estado: "bloqueado",
      roles: [],
      permisos: [],
    },
  },
}

/** Correo que hace responder 500: sirve para ver el estado de error del servicio. */
export const CORREO_FALLA = "falla@dev.local"

const DURACION_SESION_MS = 15 * 60 * 1000

let sesion: UsuarioActual | null = null

/** Deja los mocks sin sesión (las pruebas lo llaman entre casos). */
export function reiniciarSesionSimulada(usuario: UsuarioActual | null = null) {
  sesion = usuario
}

function problema(
  status: number,
  title: string,
  detail: string,
  errors?: Problema["errors"],
) {
  return HttpResponse.json<Problema>(
    { type: "about:blank", status, title, detail, ...(errors ? { errors } : {}) },
    { status, headers: { "Content-Type": "application/problem+json" } },
  )
}

const noAutorizado = () =>
  problema(401, "Unauthorized", "Correo o contraseña incorrectos")

export const handlers = [
  http.post("*/api/sesiones", async ({ request }) => {
    const cuerpo = (await request.json().catch(() => null)) as {
      correo?: unknown
      contrasena?: unknown
    } | null
    const correo =
      typeof cuerpo?.correo === "string" ? cuerpo.correo.trim().toLowerCase() : ""
    const contrasena = typeof cuerpo?.contrasena === "string" ? cuerpo.contrasena : ""

    if (!correo.includes("@") || contrasena.length === 0) {
      return problema(422, "Unprocessable Entity", "validation failed", [
        ...(!correo.includes("@")
          ? [
              {
                location: "body.correo",
                message: "El correo no tiene un formato válido.",
              },
            ]
          : []),
        ...(contrasena.length === 0
          ? [{ location: "body.contrasena", message: "La contraseña es obligatoria." }]
          : []),
      ])
    }
    if (correo === CORREO_FALLA) {
      return problema(500, "Internal Server Error", "Error interno")
    }

    const cuenta = cuentas[correo]
    if (!cuenta || cuenta.contrasena !== contrasena) return noAutorizado()
    if (cuenta.usuario.estado !== "activo") {
      return problema(403, "Forbidden", "La cuenta no está activa")
    }

    sesion = {
      ...cuenta.usuario,
      expira_en: new Date(Date.now() + DURACION_SESION_MS).toISOString(),
    }
    return HttpResponse.json(sesion)
  }),

  http.get("*/api/sesiones/actual", () => {
    if (!sesion) return problema(401, "Unauthorized", "Falta la sesión")
    return HttpResponse.json(sesion)
  }),

  http.delete("*/api/sesiones/actual", () => {
    if (!sesion) return problema(401, "Unauthorized", "Falta la sesión")
    sesion = null
    return new HttpResponse(null, { status: 204 })
  }),
]
