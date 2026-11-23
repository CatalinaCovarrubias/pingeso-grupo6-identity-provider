import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useEffect, useMemo, useRef } from "react"
import {
  avisoFallaServicio,
  avisoSinConexion,
  estadoDe,
  SIN_RESPUESTA,
  type Aviso,
} from "@/lib/problemas"
import { crearApiSesionActual, type UsuarioActual } from "./sesion-actual.api"

type Opciones = {
  urlBaseApi: string
  alSalir: () => void
  usuarioInicial?: UsuarioActual
  alNoHaySesion?: () => void
}

export type EstadoSesion =
  | { tipo: "cargando" }
  | { tipo: "sin-sesion" }
  | { tipo: "error"; aviso: Aviso }
  | { tipo: "con-sesion"; usuario: UsuarioActual }

function avisoDeError(error: unknown): Aviso {
  return estadoDe(error) === SIN_RESPUESTA ? avisoSinConexion : avisoFallaServicio
}

// Reloj de 24 horas, como se usa en Chile (sin hourCycle, es-CL puede dar "02:35 p. m.").
const hora = new Intl.DateTimeFormat("es-CL", {
  hour: "2-digit",
  minute: "2-digit",
  hourCycle: "h23",
})
const diaYHora = new Intl.DateTimeFormat("es-CL", {
  day: "numeric",
  month: "long",
  hour: "2-digit",
  minute: "2-digit",
  hourCycle: "h23",
})

const mismoDia = (a: Date, b: Date) => a.toDateString() === b.toDateString()

/** "14:35" si es hoy, o "15 de enero, 14:35" si es otro día. */
export function formatearMomento(iso: string, ahora = new Date()): string {
  const fecha = new Date(iso)
  return mismoDia(fecha, ahora) ? hora.format(fecha) : diaYHora.format(fecha)
}

/** Texto del vencimiento de la sesión y si ya pasó. */
export function describirVencimiento(expiraEn: string, ahora = new Date()) {
  const fecha = new Date(expiraEn)
  const vencida = fecha.getTime() <= ahora.getTime()
  const momento = formatearMomento(expiraEn, ahora)
  const frase = `La sesión ${vencida ? "venció" : "vence"} ${mismoDia(fecha, ahora) ? "a las" : "el"}`
  return { vencida, momento, frase }
}

const reintentar = (intentos: number, error: unknown) =>
  estadoDe(error) !== 401 && intentos < 2

/** Estado de la sesión actual y la acción de cerrarla. */
export function useSesionActual({
  urlBaseApi,
  alSalir,
  usuarioInicial,
  alNoHaySesion,
}: Opciones) {
  const api = useMemo(() => crearApiSesionActual(urlBaseApi), [urlBaseApi])
  const clienteConsultas = useQueryClient()
  const clave = useMemo(
    () => ["identidad", "sesion-actual", urlBaseApi] as const,
    [urlBaseApi],
  )

  const consulta = useQuery({
    queryKey: clave,
    queryFn: api.obtener,
    initialData: usuarioInicial,
    retry: reintentar,
  })

  const salir = useMutation({
    mutationFn: api.cerrar,
    onSuccess: () => {
      clienteConsultas.removeQueries({ queryKey: clave })
      alSalir()
    },
    onError: (error) => {
      // Si la sesión ya había vencido, el resultado es el mismo: no hay sesión.
      if (estadoDe(error) === 401) {
        clienteConsultas.removeQueries({ queryKey: clave })
        alSalir()
      }
    },
  })

  const sinSesion = consulta.isError && estadoDe(consulta.error) === 401
  const avisado = useRef(false)
  useEffect(() => {
    if (sinSesion && !avisado.current) {
      avisado.current = true
      alNoHaySesion?.()
    }
    if (!sinSesion) avisado.current = false
  }, [sinSesion, alNoHaySesion])

  // Un 401 manda aunque haya datos en caché: la sesión venció o se cerró en otra pestaña.
  let estado: EstadoSesion
  if (sinSesion) estado = { tipo: "sin-sesion" }
  else if (consulta.data) estado = { tipo: "con-sesion", usuario: consulta.data }
  else if (consulta.isError)
    estado = { tipo: "error", aviso: avisoDeError(consulta.error) }
  else estado = { tipo: "cargando" }

  const errorAlSalir =
    salir.isError && estadoDe(salir.error) !== 401 ? avisoDeError(salir.error) : null

  return {
    estado,
    reintentar: () => void consulta.refetch(),
    cerrarSesion: () => salir.mutate(),
    cerrando: salir.isPending,
    errorAlSalir,
  }
}
