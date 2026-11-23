import createClient from "openapi-fetch"
import type { components, paths } from "@/api/esquema"
import { ErrorApi, SIN_RESPUESTA } from "@/lib/problemas"

export type UsuarioActual = components["schemas"]["UsuarioActual"]

/**
 * Cliente de la API para la sesión actual. Es el único lugar de este microcomponente que
 * hace llamadas HTTP. La cookie `sesion` viaja sola gracias a `credentials: "include"`.
 */
export function crearApiSesionActual(urlBaseApi: string) {
  const cliente = createClient<paths>({ baseUrl: urlBaseApi, credentials: "include" })

  return {
    async obtener(): Promise<UsuarioActual> {
      let resultado
      try {
        resultado = await cliente.GET("/api/sesiones/actual")
      } catch {
        throw new ErrorApi(SIN_RESPUESTA)
      }
      if (resultado.data) return resultado.data
      throw new ErrorApi(resultado.response.status, resultado.error)
    },

    async cerrar(): Promise<void> {
      let resultado
      try {
        resultado = await cliente.DELETE("/api/sesiones/actual")
      } catch {
        throw new ErrorApi(SIN_RESPUESTA)
      }
      if (resultado.response.ok) return
      throw new ErrorApi(resultado.response.status, resultado.error)
    },
  }
}
