import type { ComponentProps } from "react"
import { cn } from "@/lib/utils"

type Props = ComponentProps<"section"> & {
  /** Sobre abierto: la banda con el patrón se recoge (lo usa sesion-actual). */
  abierto?: boolean
}

/**
 * Contenedor de los microcomponentes de sesión: una tarjeta con la banda del patrón de
 * seguridad arriba. La banda es decorativa (aria-hidden) y nunca lleva texto encima.
 */
export function Sobre({ abierto = false, className, children, ...props }: Props) {
  return (
    <section
      data-sobre={abierto ? "abierto" : "cerrado"}
      className={cn(
        "w-full max-w-[26.25rem] overflow-hidden rounded-lg bg-card text-card-foreground shadow-md",
        className,
      )}
      {...props}
    >
      <div
        aria-hidden="true"
        className={cn("patron-sobre h-10", abierto && "sobre-abierto")}
      />
      <div
        className={cn(
          "flex flex-col gap-6 p-6 sm:p-8",
          // La solapa recogida deja aire propio: se compensa arriba.
          abierto && "contenido-sobre pt-2 sm:pt-3",
        )}
      >
        {children}
      </div>
    </section>
  )
}
