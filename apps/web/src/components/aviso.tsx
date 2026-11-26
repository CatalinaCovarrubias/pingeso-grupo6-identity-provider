import { CircleAlertIcon, CircleCheckIcon, TriangleAlertIcon } from "lucide-react"
import type { ReactNode } from "react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { cn } from "@/lib/utils"

const tonos = {
  peligro: {
    clases: "border-peligro-borde bg-peligro-fondo text-peligro-texto",
    Icono: CircleAlertIcon,
  },
  advertencia: {
    clases: "border-advertencia-borde bg-advertencia-fondo text-advertencia-texto",
    Icono: TriangleAlertIcon,
  },
  exito: {
    clases: "border-exito-borde bg-exito-fondo text-exito-texto",
    Icono: CircleCheckIcon,
  },
} as const

type Props = {
  tono: keyof typeof tonos
  titulo: string
  children?: ReactNode
  /** Botón u otra acción para salir del problema, debajo del texto. */
  accion?: ReactNode
  className?: string
}

/** Alerta con fondo tintado, ícono y texto: nunca comunica solo con color. */
export function Aviso({ tono, titulo, children, accion, className }: Props) {
  const { clases, Icono } = tonos[tono]
  return (
    <Alert className={cn("gap-1 px-4 py-3 text-sm", clases, className)}>
      <Icono aria-hidden="true" />
      <AlertTitle className="font-semibold">{titulo}</AlertTitle>
      {children ? (
        <AlertDescription className="text-pretty text-current">
          {children}
        </AlertDescription>
      ) : null}
      {accion ? <div className="col-start-2 mt-2">{accion}</div> : null}
    </Alert>
  )
}
