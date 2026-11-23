import { setupWorker } from "msw/browser"
import { handlers } from "./handlers"

/** Service worker de MSW para la app de desarrollo (VITE_USAR_MOCKS=true). */
export const trabajador = setupWorker(...handlers)
