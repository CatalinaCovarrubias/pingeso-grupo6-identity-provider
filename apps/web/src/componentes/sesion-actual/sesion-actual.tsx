import type { QueryClient } from "@tanstack/react-query"
import { CircleCheckIcon, KeyRoundIcon, LogOutIcon, RotateCcwIcon } from "lucide-react"
import { useId } from "react"
import { Aviso } from "@/components/aviso"
import { Boton } from "@/components/boton"
import { Insignia } from "@/components/insignia"
import { Sobre } from "@/components/sobre"
import { Skeleton } from "@/components/ui/skeleton"
import { ConConsultas } from "@/lib/consultas"
import type { UsuarioActual } from "./sesion-actual.api"
import {
  describirVencimiento,
  formatearMomento,
  useSesionActual,
} from "./use-sesion-actual"

export type PropsSesionActual = {
  /** Origen de la API, sin /api (las rutas del contrato ya lo incluyen). "" = mismo origen. */
  urlBaseApi: string
  /** Se llama cuando la sesión se cerró (o ya no existía al intentar cerrarla). */
  alSalir: () => void
  /** El usuario que entregó alEntrar de inicio-sesion: evita pedirlo de nuevo. */
  usuarioInicial?: UsuarioActual
  /** Se llama una vez cuando el servicio responde que no hay sesión (401). */
  alNoHaySesion?: () => void
  /** QueryClient del host, si quiere compartir caché. Si falta, se usa uno propio. */
  clienteConsultas?: QueryClient
}

/** Muestra con qué cuenta, roles y permisos está abierta la sesión, y permite cerrarla. */
export function SesionActual({ clienteConsultas, ...props }: PropsSesionActual) {
  return (
    <ConConsultas cliente={clienteConsultas}>
      <VistaSesionActual {...props} />
    </ConConsultas>
  )
}

const nombresEstado = {
  activo: "Cuenta activa",
  bloqueado: "Cuenta bloqueada",
  inactivo: "Cuenta inactiva",
} as const

function VistaSesionActual(props: Omit<PropsSesionActual, "clienteConsultas">) {
  const idTitulo = useId()
  const { estado, reintentar, cerrarSesion, cerrando, errorAlSalir } =
    useSesionActual(props)

  if (estado.tipo === "cargando") {
    return (
      <Sobre aria-labelledby={idTitulo} aria-busy="true">
        <h2 id={idTitulo} className="sr-only">
          Tu sesión
        </h2>
        <output className="sr-only">Cargando tu sesión…</output>
        <div aria-hidden="true" className="flex flex-col gap-6">
          <div className="flex flex-col gap-3">
            <Skeleton className="h-7 w-3/4" />
            <Skeleton className="h-5 w-1/2" />
          </div>
          <div className="flex flex-col gap-2">
            <Skeleton className="h-4 w-16" />
            <Skeleton className="h-6 w-32" />
          </div>
          <div className="flex flex-col gap-2">
            <Skeleton className="h-4 w-20" />
            <Skeleton className="h-5 w-40" />
            <Skeleton className="h-5 w-36" />
          </div>
          <Skeleton className="h-10 w-full" />
        </div>
      </Sobre>
    )
  }

  if (estado.tipo === "sin-sesion") {
    return (
      <Sobre aria-labelledby={idTitulo}>
        <div className="flex flex-col items-start gap-4">
          <span className="flex size-12 items-center justify-center rounded-full bg-accent text-accent-foreground">
            <KeyRoundIcon aria-hidden="true" className="size-6" />
          </span>
          <div className="flex flex-col gap-1">
            <h2 id={idTitulo} className="text-xl font-semibold text-foreground">
              No hay una sesión activa
            </h2>
            <p className="text-base text-pretty text-muted-foreground">
              Inicia sesión para ver tu cuenta, tus roles y tus permisos.
            </p>
          </div>
        </div>
      </Sobre>
    )
  }

  if (estado.tipo === "error") {
    return (
      <Sobre aria-labelledby={idTitulo}>
        <h2 id={idTitulo} className="text-xl font-semibold text-foreground">
          Tu sesión
        </h2>
        <Aviso
          tono="peligro"
          titulo={estado.aviso.titulo}
          accion={
            <Boton variante="secundario" onClick={reintentar}>
              <RotateCcwIcon aria-hidden="true" />
              Reintentar
            </Boton>
          }
        >
          {estado.aviso.detalle}
        </Aviso>
      </Sobre>
    )
  }

  const { usuario } = estado
  const roles = usuario.roles ?? []
  const permisos = usuario.permisos ?? []
  const vencimiento = describirVencimiento(usuario.expira_en)

  return (
    <Sobre abierto aria-labelledby={idTitulo}>
      <header className="flex flex-col gap-3">
        <div className="flex flex-col gap-1">
          <p className="text-sm text-muted-foreground">Sesión iniciada como</p>
          <h2
            id={idTitulo}
            className="text-xl font-semibold break-all text-foreground sm:text-2xl"
          >
            {usuario.correo}
          </h2>
        </div>
        <div className="flex flex-wrap items-center gap-x-3 gap-y-2">
          <Insignia tono={usuario.estado === "activo" ? "exito" : "advertencia"}>
            {usuario.estado === "activo" ? <CircleCheckIcon aria-hidden="true" /> : null}
            {nombresEstado[usuario.estado]}
          </Insignia>
          <p className="text-sm text-muted-foreground">
            {vencimiento.frase}{" "}
            <time
              dateTime={usuario.expira_en}
              className="font-mono text-foreground tabular-nums"
            >
              {vencimiento.momento}
            </time>
          </p>
        </div>
      </header>

      <section aria-labelledby={`${idTitulo}-roles`} className="flex flex-col gap-2">
        <h3 id={`${idTitulo}-roles`} className="text-sm font-semibold text-foreground">
          Roles
        </h3>
        {roles.length > 0 ? (
          <ul className="flex flex-wrap gap-2">
            {roles.map((rol) => (
              <li key={rol.id}>
                <Insignia tono="tinta">{rol.nombre}</Insignia>
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-sm text-muted-foreground">
            Esta cuenta no tiene roles asignados.
          </p>
        )}
      </section>

      <section aria-labelledby={`${idTitulo}-permisos`} className="flex flex-col gap-2">
        <h3 id={`${idTitulo}-permisos`} className="text-sm font-semibold text-foreground">
          Permisos
        </h3>
        {permisos.length > 0 ? (
          <ul className="flex flex-col gap-1.5">
            {permisos.map((permiso) => (
              <li key={permiso} className="font-mono text-sm break-all text-foreground">
                {permiso}
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-sm text-pretty text-muted-foreground">
            Sin permisos adicionales: esta cuenta solo puede ver su sesión y su perfil.
          </p>
        )}
      </section>

      {usuario.ultimo_acceso ? (
        <p className="text-sm text-muted-foreground">
          Ingreso anterior:{" "}
          <time dateTime={usuario.ultimo_acceso} className="tabular-nums">
            {formatearMomento(usuario.ultimo_acceso)}
          </time>
        </p>
      ) : null}

      <footer className="flex flex-col gap-3 border-t border-border pt-6">
        {errorAlSalir ? (
          <Aviso tono="peligro" titulo="No pudimos cerrar la sesión">
            {errorAlSalir.detalle}
          </Aviso>
        ) : null}
        <Boton
          variante="secundario"
          cargando={cerrando}
          textoCargando="Cerrando sesión…"
          onClick={cerrarSesion}
          className="w-full sm:w-auto sm:self-start"
        >
          <LogOutIcon aria-hidden="true" />
          Cerrar sesión
        </Boton>
      </footer>
    </Sobre>
  )
}
