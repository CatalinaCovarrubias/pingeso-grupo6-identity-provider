import { describe, expect, it } from "vitest"
import {
  describirVencimiento,
  formatearMomento,
} from "@/componentes/sesion-actual/use-sesion-actual"
import { ErrorApi, erroresPorCampo, estadoDe, SIN_RESPUESTA } from "./problemas"

describe("problemas", () => {
  it("toma el estado de un ErrorApi y usa 0 para cualquier otro error", () => {
    expect(estadoDe(new ErrorApi(403))).toBe(403)
    expect(estadoDe(new TypeError("fetch failed"))).toBe(SIN_RESPUESTA)
  })

  it("extrae los errores por campo de un 422, solo del body y el primero de cada uno", () => {
    const error = new ErrorApi(422, {
      type: "about:blank",
      errors: [
        { location: "body.correo", message: "primero" },
        { location: "body.correo", message: "segundo" },
        { location: "query.pagina", message: "no es del body" },
        { message: "sin ubicación" },
      ],
    })
    expect(erroresPorCampo(error)).toEqual({ correo: "primero" })
    expect(erroresPorCampo(new ErrorApi(401))).toEqual({})
  })

  it("usa el detail del problema como mensaje del error", () => {
    expect(new ErrorApi(401, { type: "about:blank", detail: "No" }).message).toBe("No")
    expect(new ErrorApi(500).message).toBe("La API respondió 500")
  })
})

describe("fechas de la sesión", () => {
  const ahora = new Date(2026, 0, 15, 12, 0)

  it("muestra solo la hora si es hoy y la fecha si es otro día", () => {
    expect(formatearMomento(new Date(2026, 0, 15, 14, 35).toISOString(), ahora)).toMatch(
      /14:35/,
    )
    expect(formatearMomento(new Date(2026, 0, 16, 9, 5).toISOString(), ahora)).toMatch(
      /16 de enero/,
    )
  })

  it("dice si la sesión vence o ya venció", () => {
    const futura = describirVencimiento(
      new Date(2026, 0, 15, 12, 15).toISOString(),
      ahora,
    )
    expect(futura).toMatchObject({ vencida: false, frase: "La sesión vence a las" })
    const pasada = describirVencimiento(new Date(2026, 0, 14, 23, 0).toISOString(), ahora)
    expect(pasada).toMatchObject({ vencida: true, frase: "La sesión venció el" })
  })
})
