import { StrictMode } from "react"
import { createRoot } from "react-dom/client"
import "@/estilos/tema.css"
// Las fuentes se cargan solo en la app de desarrollo: los microcomponentes no las fuerzan.
import "@/estilos/fuentes.css"
import { AppDesarrollo } from "./app-desarrollo"

async function arrancar() {
  if (import.meta.env.VITE_USAR_MOCKS === "true") {
    const { trabajador } = await import("@/mocks/navegador")
    await trabajador.start({ onUnhandledRequest: "bypass", quiet: true })
  }

  const raiz = document.getElementById("raiz")
  if (!raiz) throw new Error("Falta el elemento #raiz en index.html")
  createRoot(raiz).render(
    <StrictMode>
      <AppDesarrollo />
    </StrictMode>,
  )
}

void arrancar()
