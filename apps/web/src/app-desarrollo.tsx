// App de desarrollo: monta los microcomponentes como lo haría un grupo cliente. No se publica.
import { useState } from "react"
import { Insignia } from "@/components/insignia"
import { InicioSesion, type UsuarioActual } from "@/componentes/inicio-sesion"
import { SesionActual } from "@/componentes/sesion-actual"

const usaMocks = import.meta.env.VITE_USAR_MOCKS === "true"

// Mismo origen: en dev el proxy de Vite (y en la imagen el spa-server) reenvía /api.
const urlBaseApi = window.location.origin

type Pantalla = { tipo: "sesion"; usuario?: UsuarioActual } | { tipo: "entrar" }

const cuentasDePrueba = [
  ["admin@dev.local", "admin-dev-local"],
  ["operador@dev.local", "operador-dev-local"],
  ["usuario@dev.local", "usuario-dev-local"],
  ...(usaMocks
    ? [
        ["bloqueado@dev.local", "bloqueado-dev-local"],
        ["falla@dev.local", "cualquiera"],
      ]
    : []),
] as const

export function AppDesarrollo() {
  // Parte preguntando por la sesión: si ya hay cookie, se muestra directamente.
  const [pantalla, setPantalla] = useState<Pantalla>({ tipo: "sesion" })

  return (
    <div className="flex min-h-dvh flex-col">
      <header className="border-b border-border bg-card">
        <div className="mx-auto flex max-w-5xl flex-wrap items-center justify-between gap-3 px-4 py-3 sm:px-6">
          <h1 className="text-base font-semibold text-foreground">
            Identidad{" "}
            <span className="font-normal text-muted-foreground">· app de desarrollo</span>
          </h1>
          <Insignia tono={usaMocks ? "advertencia" : "tinta"}>
            {usaMocks ? "Datos simulados (MSW)" : "Backend real"}
          </Insignia>
        </div>
      </header>

      <main className="mx-auto grid w-full max-w-5xl flex-1 gap-10 px-4 py-10 sm:px-6 md:grid-cols-[minmax(0,26.25rem)_minmax(0,1fr)] md:items-start md:gap-12 md:py-16">
        <div className="flex justify-center md:justify-start">
          {pantalla.tipo === "entrar" ? (
            <InicioSesion
              urlBaseApi={urlBaseApi}
              alEntrar={(usuario) => setPantalla({ tipo: "sesion", usuario })}
            />
          ) : (
            <SesionActual
              key={pantalla.usuario?.id ?? "sin-usuario"}
              urlBaseApi={urlBaseApi}
              usuarioInicial={pantalla.usuario}
              alNoHaySesion={() => setPantalla({ tipo: "entrar" })}
              alSalir={() => setPantalla({ tipo: "entrar" })}
            />
          )}
        </div>

        <aside
          aria-labelledby="titulo-cuentas"
          className="flex max-w-prose flex-col gap-4"
        >
          <h2 id="titulo-cuentas" className="text-lg font-semibold text-foreground">
            Cuentas de prueba
          </h2>
          <p className="text-sm text-pretty text-muted-foreground">
            Existen solo con <code className="font-mono">API_ENV=dev</code>. Las
            contraseñas son públicas a propósito.
            {usaMocks
              ? " En modo simulado también hay una cuenta bloqueada (403) y una que hace fallar al servicio (500)."
              : null}
          </p>
          <table className="w-full text-left text-sm">
            <thead>
              <tr className="border-b border-border text-muted-foreground">
                <th scope="col" className="py-2 pr-4 font-semibold">
                  Correo
                </th>
                <th scope="col" className="py-2 font-semibold">
                  Contraseña
                </th>
              </tr>
            </thead>
            <tbody className="font-mono">
              {cuentasDePrueba.map(([correo, contrasena]) => (
                <tr key={correo} className="border-b border-border last:border-0">
                  <td className="py-2 pr-4 break-all text-foreground">{correo}</td>
                  <td className="py-2 break-all text-foreground">{contrasena}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </aside>
      </main>
    </div>
  )
}
