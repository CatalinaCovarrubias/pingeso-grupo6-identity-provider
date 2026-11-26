// Genera src/mocks/ejemplos.gen.ts con un ejemplo de cada esquema del contrato.
// Usa los `examples` de contracts/openapi.yaml; si un campo no tiene, pone un valor fijo
// según su tipo y formato, para que el archivo sea determinista y `git diff` no cambie.
import { readFileSync, writeFileSync } from "node:fs"
import { fileURLToPath } from "node:url"
import { parse } from "yaml"

const rutaContrato = fileURLToPath(
  new URL("../../../contracts/openapi.yaml", import.meta.url),
)
const rutaSalida = fileURLToPath(new URL("../src/mocks/ejemplos.gen.ts", import.meta.url))

const contrato = parse(readFileSync(rutaContrato, "utf8"))
const esquemas = contrato.components.schemas

const porFormato = {
  "date-time": "2026-01-15T12:00:00Z",
  date: "2026-01-15",
  uuid: "00000000-0000-4000-8000-000000000000",
  email: "persona@ejemplo.cl",
  uri: "https://ejemplo.cl",
}

function resolver(esquema) {
  if (esquema?.$ref) return esquemas[esquema.$ref.split("/").pop()]
  return esquema
}

function tipoDe(esquema) {
  const tipo = esquema.type
  if (Array.isArray(tipo)) return tipo.find((t) => t !== "null")
  if (tipo) return tipo
  if (esquema.properties) return "object"
  return undefined
}

function ejemplo(esquemaCrudo, profundidad = 0) {
  const esquema = resolver(esquemaCrudo)
  if (!esquema || profundidad > 6) return null
  if (Array.isArray(esquema.examples) && esquema.examples.length > 0)
    return esquema.examples[0]
  if (esquema.default !== undefined) return esquema.default
  if (Array.isArray(esquema.enum)) return esquema.enum[0]

  switch (tipoDe(esquema)) {
    case "object": {
      const objeto = {}
      const requeridos = new Set(esquema.required ?? [])
      for (const [nombre, propiedad] of Object.entries(esquema.properties ?? {})) {
        if (propiedad.readOnly) continue
        const tieneEjemplo = Array.isArray(resolver(propiedad)?.examples)
        if (requeridos.has(nombre) || tieneEjemplo) {
          objeto[nombre] = ejemplo(propiedad, profundidad + 1)
        }
      }
      return objeto
    }
    case "array":
      return esquema.items ? [ejemplo(esquema.items, profundidad + 1)] : []
    case "string":
      return porFormato[esquema.format] ?? "texto"
    case "integer":
    case "number":
      return esquema.minimum ?? 0
    case "boolean":
      return false
    default:
      return null
  }
}

const nombres = Object.keys(esquemas).sort()
const cuerpo = nombres
  .map(
    (nombre) =>
      `  ${nombre}: ${JSON.stringify(ejemplo(esquemas[nombre]), null, 2).replace(/\n/g, "\n  ")},`,
  )
  .join("\n")

const salida = `// Generado por scripts/generar-mocks.mjs desde contracts/openapi.yaml. No editar a mano:
// se regenera con \`pnpm codegen\`.
import type { components } from "@/api/esquema"

type Esquemas = components["schemas"]

export const ejemplos = {
${cuerpo}
} satisfies { [K in keyof Esquemas]: Esquemas[K] }
`

writeFileSync(rutaSalida, salida)
console.log(`ejemplos de ${nombres.length} esquemas → src/mocks/ejemplos.gen.ts`)
