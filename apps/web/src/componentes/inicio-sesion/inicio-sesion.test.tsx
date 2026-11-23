import { QueryClient } from "@tanstack/react-query"
import { render, screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { http, HttpResponse } from "msw"
import { describe, expect, it, vi } from "vitest"
import { servidor } from "@/mocks/servidor"
import { URL_API } from "@/pruebas/configuracion"
import { InicioSesion } from "./inicio-sesion"

function montar(props: Partial<Parameters<typeof InicioSesion>[0]> = {}) {
  const alEntrar = vi.fn()
  const usuario = userEvent.setup()
  render(<InicioSesion urlBaseApi={URL_API} alEntrar={alEntrar} {...props} />)
  const correo = screen.getByLabelText("Correo")
  const contrasena = screen.getByLabelText("Contraseña")
  const entrar = screen.getByRole("button", { name: "Entrar" })
  return { alEntrar, usuario, correo, contrasena, entrar }
}

async function entrarCon(correoTexto: string, contrasenaTexto: string) {
  const elementos = montar()
  await elementos.usuario.type(elementos.correo, correoTexto)
  await elementos.usuario.type(elementos.contrasena, contrasenaTexto)
  await elementos.usuario.click(elementos.entrar)
  return elementos
}

describe("inicio-sesion", () => {
  it("entra con credenciales correctas y entrega el usuario", async () => {
    const { alEntrar } = await entrarCon("admin@dev.local", "admin-dev-local")

    await waitFor(() => expect(alEntrar).toHaveBeenCalledTimes(1))
    expect(alEntrar.mock.calls[0]?.[0]).toMatchObject({
      correo: "admin@dev.local",
      estado: "activo",
      permisos: ["perfiles:operar", "roles:administrar", "usuarios:administrar"],
    })
  })

  it("con credenciales malas muestra un error genérico y limpia la contraseña", async () => {
    const { alEntrar, contrasena } = await entrarCon("admin@dev.local", "otra-clave")

    const alerta = await screen.findByRole("alert")
    expect(alerta).toHaveTextContent("El correo o la contraseña no son correctos")
    expect(contrasena).toHaveValue("")
    expect(contrasena).toHaveFocus()
    expect(alEntrar).not.toHaveBeenCalled()
  })

  it("con una cuenta bloqueada explica que no está activa", async () => {
    const { alEntrar } = await entrarCon("bloqueado@dev.local", "bloqueado-dev-local")

    expect(await screen.findByRole("alert")).toHaveTextContent("Tu cuenta no está activa")
    expect(alEntrar).not.toHaveBeenCalled()
  })

  it("valida en el navegador antes de llamar a la API", async () => {
    const pedidos = vi.fn()
    servidor.events.on("request:start", pedidos)
    const { usuario, correo, entrar } = montar()

    await usuario.click(entrar)
    expect(await screen.findByText("Escribe tu correo.")).toBeInTheDocument()
    expect(screen.getByText("Escribe tu contraseña.")).toBeInTheDocument()
    expect(correo).toHaveAttribute("aria-invalid", "true")

    await usuario.type(correo, "sin-arroba")
    await usuario.click(entrar)
    expect(
      await screen.findByText(/debe tener la forma nombre@dominio.cl/),
    ).toBeInTheDocument()

    expect(pedidos).not.toHaveBeenCalled()
    servidor.events.removeAllListeners()
  })

  it("rechaza contraseñas de más de 72 bytes aunque tengan menos de 72 caracteres", async () => {
    const { usuario, correo, contrasena, entrar, alEntrar } = montar()
    await usuario.type(correo, "admin@dev.local")
    await usuario.type(contrasena, "ñ".repeat(37)) // 37 caracteres, 74 bytes

    expect(screen.getByText(/Llevas 74 de 72 bytes/)).toBeInTheDocument()
    await usuario.click(entrar)
    expect(await screen.findByText(/es demasiado larga/)).toBeInTheDocument()
    expect(alEntrar).not.toHaveBeenCalled()
  })

  it("marca el campo que el servicio rechaza con 422", async () => {
    servidor.use(
      http.post("*/api/sesiones", () =>
        HttpResponse.json(
          {
            type: "about:blank",
            status: 422,
            title: "Unprocessable Entity",
            errors: [
              {
                location: "body.correo",
                message: "expected string to be RFC 5322 email",
              },
            ],
          },
          { status: 422, headers: { "Content-Type": "application/problem+json" } },
        ),
      ),
    )
    await entrarCon("raro@dev.local", "x")

    expect(await screen.findByText("Revisa los datos marcados")).toBeInTheDocument()
    expect(screen.getByText("expected string to be RFC 5322 email")).toBeInTheDocument()
    expect(screen.getByLabelText("Correo")).toHaveAttribute("aria-invalid", "true")
  })

  it("distingue una falla del servicio de una falla de conexión", async () => {
    await entrarCon("falla@dev.local", "cualquiera")
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Algo falló de nuestro lado",
    )
  })

  it("avisa cuando no hay conexión con el servicio", async () => {
    servidor.use(http.post("*/api/sesiones", () => HttpResponse.error()))
    await entrarCon("admin@dev.local", "admin-dev-local")
    expect(await screen.findByRole("alert")).toHaveTextContent("No pudimos conectar")
  })

  it("muestra y oculta la contraseña", async () => {
    const { usuario, contrasena } = montar()
    expect(contrasena).toHaveAttribute("type", "password")

    await usuario.click(screen.getByRole("button", { name: "Mostrar contraseña" }))
    expect(contrasena).toHaveAttribute("type", "text")
    await usuario.click(screen.getByRole("button", { name: "Ocultar contraseña" }))
    expect(contrasena).toHaveAttribute("type", "password")
  })

  it("acepta el título, el correo inicial y un QueryClient del host", async () => {
    const alEntrar = vi.fn()
    render(
      <InicioSesion
        urlBaseApi={URL_API}
        alEntrar={alEntrar}
        titulo="Entrar al sistema"
        correoInicial="operador@dev.local"
        clienteConsultas={new QueryClient()}
      />,
    )
    expect(screen.getByRole("heading", { name: "Entrar al sistema" })).toBeInTheDocument()
    expect(screen.getByLabelText("Correo")).toHaveValue("operador@dev.local")

    const usuario = userEvent.setup()
    await usuario.type(screen.getByLabelText("Contraseña"), "operador-dev-local{Enter}")
    await waitFor(() => expect(alEntrar).toHaveBeenCalledTimes(1))
  })
})
