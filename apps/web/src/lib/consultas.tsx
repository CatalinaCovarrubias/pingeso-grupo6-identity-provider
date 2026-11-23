import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { useState, type ReactNode } from "react"

type Props = {
  /** Cliente del host, si quiere compartir caché. Si falta, se crea uno propio. */
  cliente?: QueryClient
  children: ReactNode
}

function crearClientePropio() {
  return new QueryClient({
    defaultOptions: {
      queries: { refetchOnWindowFocus: false, staleTime: 30_000 },
      mutations: { retry: 0 },
    },
  })
}

/**
 * Da a un microcomponente su propio contexto de react-query, sin depender de un provider
 * global del host. Un provider anidado solo afecta a su subárbol, así que no interfiere
 * con el del host si este tiene uno.
 */
export function ConConsultas({ cliente, children }: Props) {
  // Crear un QueryClient es barato; tenerlo siempre listo evita casos raros si el host
  // empieza o deja de pasar el suyo.
  const [propio] = useState(crearClientePropio)
  return <QueryClientProvider client={cliente ?? propio}>{children}</QueryClientProvider>
}
