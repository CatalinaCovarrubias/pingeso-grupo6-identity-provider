import { Loader2Icon } from "lucide-react"
import type { ComponentProps } from "react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

type Props = Omit<ComponentProps<typeof Button>, "variant" | "size"> & {
  /** primario: la acción principal (una por vista). secundario: con borde. */
  variante?: "primario" | "secundario"
  /** Mientras es true, el botón muestra el spinner y no se puede volver a pulsar. */
  cargando?: boolean
  /** Texto que reemplaza a children mientras carga, por ejemplo "Entrando…". */
  textoCargando?: string
}

const variantes = {
  primario: "default",
  secundario: "outline",
} as const

/** Botón de los microcomponentes: alto cómodo para el dedo (40 px) y estado de carga. */
export function Boton({
  variante = "primario",
  cargando = false,
  textoCargando,
  disabled,
  className,
  children,
  ...props
}: Props) {
  return (
    <Button
      variant={variantes[variante]}
      disabled={disabled || cargando}
      focusableWhenDisabled={cargando}
      aria-busy={cargando || undefined}
      className={cn(
        "h-10 gap-2 px-4 text-base md:text-sm",
        variante === "primario" && "hover:bg-primary-hover",
        variante === "secundario" && "border-input bg-card",
        className,
      )}
      {...props}
    >
      {cargando ? <Loader2Icon aria-hidden="true" className="animate-spin" /> : null}
      {cargando && textoCargando ? textoCargando : children}
    </Button>
  )
}
