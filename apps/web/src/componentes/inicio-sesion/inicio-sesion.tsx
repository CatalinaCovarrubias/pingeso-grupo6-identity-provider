import type { QueryClient } from "@tanstack/react-query"
import { EyeIcon, EyeOffIcon } from "lucide-react"
import { useId } from "react"
import { Aviso } from "@/components/aviso"
import { Boton } from "@/components/boton"
import { Campo } from "@/components/campo"
import { Sobre } from "@/components/sobre"
import { Button } from "@/components/ui/button"
import { ConConsultas } from "@/lib/consultas"
import type { UsuarioActual } from "./inicio-sesion.api"
import { BYTES_MAXIMOS_CONTRASENA, reglas, useInicioSesion } from "./use-inicio-sesion"

export type PropsInicioSesion = {
  /** Origen de la API, sin /api (las rutas del contrato ya lo incluyen). "" = mismo origen. */
  urlBaseApi: string
  /** Se llama con el usuario cuando el servicio acepta las credenciales y fija la cookie. */
  alEntrar: (usuario: UsuarioActual) => void
  /** Correo con el que se rellena el formulario. */
  correoInicial?: string
  /** Título del formulario. Por defecto "Iniciar sesión". */
  titulo?: string
  /** QueryClient del host, si quiere compartir caché. Si falta, se usa uno propio. */
  clienteConsultas?: QueryClient
}

/** Formulario de inicio de sesión con correo y contraseña. */
export function InicioSesion({ clienteConsultas, ...props }: PropsInicioSesion) {
  return (
    <ConConsultas cliente={clienteConsultas}>
      <VistaInicioSesion {...props} />
    </ConConsultas>
  )
}

function VistaInicioSesion({
  urlBaseApi,
  alEntrar,
  correoInicial,
  titulo = "Iniciar sesión",
}: Omit<PropsInicioSesion, "clienteConsultas">) {
  const idTitulo = useId()
  const {
    formulario,
    enviar,
    enviando,
    aviso,
    contrasenaVisible,
    alternarContrasena,
    bytesContrasena,
    cercaDelLimite,
  } = useInicioSesion({ urlBaseApi, alEntrar, correoInicial })
  const { register, formState } = formulario
  const errores = formState.errors

  return (
    <Sobre aria-labelledby={idTitulo}>
      <header className="flex flex-col gap-1">
        <h2 id={idTitulo} className="text-2xl font-semibold text-balance text-foreground">
          {titulo}
        </h2>
        <p className="text-base text-muted-foreground">Usa tu correo y tu contraseña.</p>
      </header>

      <form
        noValidate
        onSubmit={enviar}
        className="flex flex-col gap-5"
        aria-busy={enviando}
      >
        {aviso ? (
          <Aviso tono="peligro" titulo={aviso.titulo}>
            {aviso.detalle}
          </Aviso>
        ) : null}

        <Campo
          etiqueta="Correo"
          type="email"
          autoComplete="username"
          inputMode="email"
          autoCapitalize="none"
          spellCheck={false}
          error={errores.correo?.message}
          {...register("correo", reglas.correo)}
        />

        <Campo
          etiqueta="Contraseña"
          type={contrasenaVisible ? "text" : "password"}
          autoComplete="current-password"
          error={errores.contrasena?.message}
          ayuda={
            cercaDelLimite && !errores.contrasena
              ? `Llevas ${bytesContrasena} de ${BYTES_MAXIMOS_CONTRASENA} bytes permitidos (las letras con tilde y la ñ cuentan doble).`
              : undefined
          }
          adorno={
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="size-9 text-muted-foreground hover:text-foreground"
              aria-label={contrasenaVisible ? "Ocultar contraseña" : "Mostrar contraseña"}
              aria-pressed={contrasenaVisible}
              onClick={alternarContrasena}
            >
              {contrasenaVisible ? (
                <EyeOffIcon aria-hidden="true" />
              ) : (
                <EyeIcon aria-hidden="true" />
              )}
            </Button>
          }
          {...register("contrasena", reglas.contrasena)}
        />

        <Boton
          type="submit"
          cargando={enviando}
          textoCargando="Entrando…"
          className="mt-1 w-full"
        >
          Entrar
        </Boton>
      </form>
    </Sobre>
  )
}
