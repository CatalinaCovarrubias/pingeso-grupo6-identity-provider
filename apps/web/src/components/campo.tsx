import { useId, type ComponentProps, type ReactNode } from "react"
import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"

type Props = ComponentProps<"input"> & {
  etiqueta: string
  /** Mensaje de error del campo; si existe, el input queda marcado como inválido. */
  error?: string
  /** Ayuda bajo el campo, por ejemplo un límite que se está por alcanzar. */
  ayuda?: ReactNode
  /** Contenido a la derecha dentro del input, por ejemplo el botón de mostrar contraseña. */
  adorno?: ReactNode
}

/** Campo de formulario con label visible, ayuda y error enlazados por aria-describedby. */
export function Campo({
  etiqueta,
  error,
  ayuda,
  adorno,
  id,
  className,
  ...props
}: Props) {
  const generado = useId()
  const idInput = id ?? generado
  const idAyuda = `${idInput}-ayuda`
  const idError = `${idInput}-error`
  const describedBy = [ayuda ? idAyuda : null, error ? idError : null]
    .filter(Boolean)
    .join(" ")

  return (
    <Field data-invalid={error ? true : undefined} className="gap-2">
      <FieldLabel htmlFor={idInput} className="text-sm font-semibold text-foreground">
        {etiqueta}
      </FieldLabel>
      <div className="relative">
        <Input
          id={idInput}
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy || undefined}
          className={cn(
            "h-10 bg-card px-3 text-base md:text-base",
            adorno ? "pr-12" : null,
            className,
          )}
          {...props}
        />
        {adorno ? (
          <div className="absolute inset-y-0 right-0 flex items-center pr-1">
            {adorno}
          </div>
        ) : null}
      </div>
      {ayuda ? (
        <FieldDescription id={idAyuda} className="text-sm text-muted-foreground">
          {ayuda}
        </FieldDescription>
      ) : null}
      {error ? (
        <FieldError id={idError} className="text-sm text-peligro-texto">
          {error}
        </FieldError>
      ) : null}
    </Field>
  )
}
