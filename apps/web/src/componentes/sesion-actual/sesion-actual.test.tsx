import { render, screen, waitFor, within } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { http, HttpResponse } from "msw"
import { describe, expect, it, vi } from "vitest"
import { cuentas, reiniciarSesionSimulada } from "@/mocks/handlers"
import { servidor } from "@/mocks/servidor"
import { URL_API } from "@/pruebas/configuracion"
import { SesionActual } from "./sesion-actual"
import type { UsuarioActual } from "./sesion-actual.api"

const enUnRato = () => new Date(Date.now() + 10 * 60 * 1000).toISOString()

function usuarioDe(correo: string): UsuarioActual {
  const cuenta = cuentas[correo]
  if (!cuenta) throw new Error(`No existe la cuenta ${correo}`)
  return { ...cuenta.usuario, expira_en: enUnRato() }
}

function montar(props: Partial<Parameters<typeof SesionActual>[0]> = {}) {
  const alSalir = vi.fn()
  const alNoHaySesion = vi.fn()
  render(
    <SesionActual
      urlBaseApi={URL_API}
      alSalir={alSalir}
      alNoHaySesion={alNoHaySesion}
      {...props}
    />,
  )
  return { alSalir, alNoHaySesion, usuario: userEvent.setup() }
}

describe("sesion-actual", () => {
  it("muestra el usuario, sus roles y sus permisos", async () => {
    reiniciarSesionSimulada(usuarioDe("admin@dev.local"))
    montar()

    expect(
      await screen.findByRole("heading", { name: "admin@dev.local" }),
    ).toBeInTheDocument()
    expect(screen.getByText("Cuenta activa")).toBeInTheDocument()
    const roles = screen.getByRole("region", { name: "Roles" })
    expect(within(roles).getByText("administrador")).toBeInTheDocument()
    const permisos = screen.getByRole("region", { name: "Permisos" })
    expect(within(permisos).getAllByRole("listitem")).toHaveLength(3)
    expect(screen.getByText(/La sesión vence a las/)).toBeInTheDocument()
  })

  it("usa el usuario inicial sin esperar a la API", () => {
    montar({ usuarioInicial: usuarioDe("operador@dev.local") })
    expect(
      screen.getByRole("heading", { name: "operador@dev.local" }),
    ).toBeInTheDocument()
  })

  it("explica cuando la cuenta no tiene roles ni permisos", () => {
    montar({
      usuarioInicial: { ...usuarioDe("usuario@dev.local"), roles: null, permisos: null },
    })
    expect(screen.getByText("Esta cuenta no tiene roles asignados.")).toBeInTheDocument()
    expect(screen.getByText(/Sin permisos adicionales/)).toBeInTheDocument()
  })

  it("muestra la carga y luego el estado vacío cuando no hay sesión", async () => {
    const { alNoHaySesion } = montar()

    expect(screen.getByRole("status")).toHaveTextContent("Cargando tu sesión")
    expect(await screen.findByText("No hay una sesión activa")).toBeInTheDocument()
    expect(alNoHaySesion).toHaveBeenCalledTimes(1)
  })

  it("cierra la sesión y avisa al host", async () => {
    reiniciarSesionSimulada(usuarioDe("admin@dev.local"))
    const { alSalir, usuario } = montar()

    await usuario.click(await screen.findByRole("button", { name: "Cerrar sesión" }))
    await waitFor(() => expect(alSalir).toHaveBeenCalledTimes(1))
  })

  it("si la sesión ya había vencido al cerrarla, igual avisa al host", async () => {
    const { alSalir, usuario } = montar({ usuarioInicial: usuarioDe("admin@dev.local") })

    // Los mocks no tienen sesión: el DELETE responde 401.
    await usuario.click(screen.getByRole("button", { name: "Cerrar sesión" }))
    await waitFor(() => expect(alSalir).toHaveBeenCalledTimes(1))
  })

  it("si no puede cerrar la sesión, lo dice y no avisa al host", async () => {
    servidor.use(
      http.delete("*/api/sesiones/actual", () =>
        HttpResponse.json({ type: "about:blank", status: 500 }, { status: 500 }),
      ),
    )
    const { alSalir, usuario } = montar({ usuarioInicial: usuarioDe("admin@dev.local") })

    await usuario.click(screen.getByRole("button", { name: "Cerrar sesión" }))
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "No pudimos cerrar la sesión",
    )
    expect(alSalir).not.toHaveBeenCalled()
  })

  it("ante una falla del servicio permite reintentar", async () => {
    let fallar = true
    servidor.use(
      http.get("*/api/sesiones/actual", () =>
        fallar
          ? HttpResponse.json({ type: "about:blank", status: 503 }, { status: 503 })
          : HttpResponse.json(usuarioDe("admin@dev.local")),
      ),
    )
    const { usuario } = montar()

    expect(await screen.findByRole("alert", {}, { timeout: 5000 })).toHaveTextContent(
      "Algo falló de nuestro lado",
    )
    fallar = false
    await usuario.click(screen.getByRole("button", { name: "Reintentar" }))
    expect(
      await screen.findByRole("heading", { name: "admin@dev.local" }),
    ).toBeInTheDocument()
  })
})
