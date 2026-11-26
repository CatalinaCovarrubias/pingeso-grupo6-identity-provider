/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** "true" monta los mocks MSW en la app de desarrollo. No es secreto: termina en el bundle. */
  readonly VITE_USAR_MOCKS?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
