import createClient from "openapi-fetch"
import type { components, paths } from "@/api/esquema"
import { ErrorApi, SIN_RESPUESTA } from "@/lib/problemas"

export type UsuarioActual = components["schemas"]["UsuarioActual"]
export type SolicitudSesion = Pick<
  components["schemas"]["SolicitudSesion"],
  "correo" | "contrasena"
>

/**
 * Cliente de la API para el inicio de sesión. Es el único lugar de este microcomponente
 * que hace llamadas HTTP. `credentials: "include"` deja que el navegador guarde la cookie
 * `sesion` (HttpOnly) que responde el servicio.
 */
export function crearApiInicioSesion(urlBaseApi: string) {
  const cliente = createClient<paths>({ baseUrl: urlBaseApi, credentials: "include" })

  return {
    async iniciarSesion(datos: SolicitudSesion): Promise<UsuarioActual> {
      let resultado
      try {
        resultado = await cliente.POST("/api/sesiones", { body: datos })
      } catch {
        throw new ErrorApi(SIN_RESPUESTA)
      }
      if (resultado.data) return resultado.data
      throw new ErrorApi(resultado.response.status, resultado.error)
    },
  }
}
