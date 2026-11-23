import { useMutation } from "@tanstack/react-query"
import { useMemo, useState } from "react"
import { useForm, type RegisterOptions } from "react-hook-form"
import {
  avisoFallaServicio,
  avisoSinConexion,
  erroresPorCampo,
  estadoDe,
  SIN_RESPUESTA,
  type Aviso,
} from "@/lib/problemas"
import {
  crearApiInicioSesion,
  type SolicitudSesion,
  type UsuarioActual,
} from "./inicio-sesion.api"

/** Largos del contrato (SolicitudSesion en contracts/openapi.yaml). */
export const LARGO_MAXIMO_CORREO = 254
export const LARGO_MAXIMO_CONTRASENA = 72
/** bcrypt solo admite 72 bytes: una letra con tilde o una ñ ocupa 2. */
export const BYTES_MAXIMOS_CONTRASENA = 72
/** Desde cuántos bytes se avisa que la contraseña se acerca al límite. */
export const BYTES_AVISO_CONTRASENA = 60

const codificador = new TextEncoder()
export const bytesDe = (texto: string) => codificador.encode(texto).length

// Igual de permisivo que format: email del contrato: algo@algo.algo, sin espacios.
const PATRON_CORREO = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export const reglas = {
  correo: {
    required: "Escribe tu correo.",
    maxLength: {
      value: LARGO_MAXIMO_CORREO,
      message: `El correo no puede pasar de ${LARGO_MAXIMO_CORREO} caracteres.`,
    },
    pattern: {
      value: PATRON_CORREO,
      message: "Revisa el correo: debe tener la forma nombre@dominio.cl.",
    },
  },
  contrasena: {
    required: "Escribe tu contraseña.",
    maxLength: {
      value: LARGO_MAXIMO_CONTRASENA,
      message: `La contraseña no puede pasar de ${LARGO_MAXIMO_CONTRASENA} caracteres.`,
    },
    validate: (valor: string) =>
      bytesDe(valor) <= BYTES_MAXIMOS_CONTRASENA ||
      "La contraseña es demasiado larga: las letras con tilde y la ñ cuentan doble. Acórtala un poco.",
  },
} satisfies { [K in keyof SolicitudSesion]: RegisterOptions<SolicitudSesion, K> }

/** Qué decirle a la persona según cómo respondió el servicio. */
export function avisoDeInicioSesion(error: unknown): Aviso | null {
  if (!error) return null
  switch (estadoDe(error)) {
    case 401:
      // Genérico a propósito: no revela si el correo existe.
      return {
        titulo: "El correo o la contraseña no son correctos",
        detalle: "Revisa ambos datos e intenta de nuevo.",
      }
    case 403:
      return {
        titulo: "Tu cuenta no está activa",
        detalle: "Pide a quien administra el sistema que revise tu cuenta.",
      }
    case 422:
      return {
        titulo: "Revisa los datos marcados",
        detalle: "Hay campos que el servicio no pudo aceptar.",
      }
    case SIN_RESPUESTA:
      return avisoSinConexion
    default:
      return avisoFallaServicio
  }
}

type Opciones = {
  urlBaseApi: string
  alEntrar: (usuario: UsuarioActual) => void
  correoInicial?: string
}

/** Estado del formulario de inicio de sesión. */
export function useInicioSesion({ urlBaseApi, alEntrar, correoInicial }: Opciones) {
  const api = useMemo(() => crearApiInicioSesion(urlBaseApi), [urlBaseApi])
  const [contrasenaVisible, setContrasenaVisible] = useState(false)

  const formulario = useForm<SolicitudSesion>({
    defaultValues: { correo: correoInicial ?? "", contrasena: "" },
    mode: "onTouched",
  })

  const mutacion = useMutation({
    mutationFn: api.iniciarSesion,
    onSuccess: (usuario) => {
      formulario.resetField("contrasena")
      alEntrar(usuario)
    },
    onError: (error) => {
      const campos = erroresPorCampo(error)
      for (const campo of ["correo", "contrasena"] as const) {
        const mensaje = campos[campo]
        if (mensaje) formulario.setError(campo, { type: "servidor", message: mensaje })
      }
      // Con credenciales malas se limpia la contraseña y se vuelve a ella.
      if (estadoDe(error) === 401) {
        formulario.resetField("contrasena")
        formulario.setFocus("contrasena")
      }
    },
  })

  const enviar = formulario.handleSubmit((datos) => {
    mutacion.mutate({ correo: datos.correo.trim(), contrasena: datos.contrasena })
  })

  const bytesContrasena = bytesDe(formulario.watch("contrasena") ?? "")

  return {
    formulario,
    enviar,
    enviando: mutacion.isPending,
    aviso: avisoDeInicioSesion(mutacion.error),
    contrasenaVisible,
    alternarContrasena: () => setContrasenaVisible((visible) => !visible),
    bytesContrasena,
    cercaDelLimite: bytesContrasena >= BYTES_AVISO_CONTRASENA,
  }
}
