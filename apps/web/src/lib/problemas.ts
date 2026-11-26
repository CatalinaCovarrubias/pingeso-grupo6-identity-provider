import type { components } from "@/api/esquema"

/** Cuerpo de un error RFC 9457 (application/problem+json) según el contrato. */
export type Problema = components["schemas"]["ErrorModel"]

/** Estado 0: no hubo respuesta (sin red, servicio caído o CORS). */
export const SIN_RESPUESTA = 0

/** Error de una llamada a la API, con el estado HTTP y el problem+json si vino. */
export class ErrorApi extends Error {
  readonly estado: number
  readonly problema: Problema | undefined

  constructor(estado: number, problema?: Problema) {
    super(problema?.detail ?? problema?.title ?? `La API respondió ${estado}`)
    this.name = "ErrorApi"
    this.estado = estado
    this.problema = problema
  }
}

export function esErrorApi(error: unknown): error is ErrorApi {
  return error instanceof ErrorApi
}

/** Estado HTTP de cualquier error: el de ErrorApi, o SIN_RESPUESTA si es otro. */
export function estadoDe(error: unknown): number {
  return esErrorApi(error) ? error.estado : SIN_RESPUESTA
}

/**
 * Errores por campo de un 422. Huma usa `location` con la forma "body.correo".
 * Devuelve { correo: "mensaje" } solo para los campos del body.
 */
export function erroresPorCampo(error: unknown): Record<string, string> {
  if (!esErrorApi(error) || error.estado !== 422) return {}
  const campos: Record<string, string> = {}
  for (const detalle of error.problema?.errors ?? []) {
    const campo = detalle.location?.startsWith("body.")
      ? detalle.location.slice(5)
      : undefined
    if (campo && detalle.message && !(campo in campos)) campos[campo] = detalle.message
  }
  return campos
}

/** Mensaje para mostrar en pantalla: qué pasó y qué hacer. */
export type Aviso = { titulo: string; detalle: string }

export const avisoSinConexion: Aviso = {
  titulo: "No pudimos conectar con el servicio",
  detalle: "Revisa tu conexión a internet e intenta de nuevo.",
}

export const avisoFallaServicio: Aviso = {
  titulo: "Algo falló de nuestro lado",
  detalle: "No es un problema con tus datos. Intenta de nuevo en unos minutos.",
}
