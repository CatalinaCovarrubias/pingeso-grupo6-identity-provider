import type { ComponentProps } from "react"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"

const tonos = {
  neutro: "border-border bg-secondary text-secondary-foreground",
  tinta: "border-accent-border bg-accent text-accent-foreground",
  exito: "border-exito-borde bg-exito-fondo text-exito-texto",
  advertencia: "border-advertencia-borde bg-advertencia-fondo text-advertencia-texto",
  peligro: "border-peligro-borde bg-peligro-fondo text-peligro-texto",
} as const

type Props = Omit<ComponentProps<typeof Badge>, "variant"> & {
  tono?: keyof typeof tonos
}

/** Insignia con fondo tintado y texto oscuro del mismo tono. */
export function Insignia({ tono = "neutro", className, ...props }: Props) {
  return (
    <Badge
      variant="outline"
      className={cn("h-6 px-2.5 text-sm font-normal", tonos[tono], className)}
      {...props}
    />
  )
}
