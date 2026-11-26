import "@testing-library/jest-dom/vitest"
import { cleanup } from "@testing-library/react"
import { afterAll, afterEach, beforeAll } from "vitest"
import { reiniciarSesionSimulada } from "@/mocks/handlers"
import { servidor } from "@/mocks/servidor"

// Una llamada sin handler es un error de la prueba, no algo que se deje pasar a la red.
beforeAll(() => servidor.listen({ onUnhandledRequest: "error" }))
afterEach(() => {
  cleanup()
  servidor.resetHandlers()
  reiniciarSesionSimulada()
})
afterAll(() => servidor.close())

/** Origen que usan las pruebas: en Node, fetch necesita una URL absoluta. */
export const URL_API = "http://localhost"
