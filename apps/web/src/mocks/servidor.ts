import { setupServer } from "msw/node"
import { handlers } from "./handlers"

/** Servidor MSW para las pruebas de vitest. */
export const servidor = setupServer(...handlers)
