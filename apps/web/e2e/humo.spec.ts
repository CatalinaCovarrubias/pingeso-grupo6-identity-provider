import { expect, test, type Page } from "@playwright/test"

// Prueba de humo de la app de desarrollo: login, errores, sesión actual y cierre.
// Con CAPTURAS=1 guarda capturas en docs/tandas/tanda-2/ para la documentación.
const conMocks = !process.env.E2E_URL
const capturas = process.env.CAPTURAS === "1"

async function capturar(pagina: Page, nombre: string) {
  if (!capturas) return
  await pagina.screenshot({ path: `../../docs/tandas/tanda-2/${nombre}.png` })
}

async function entrar(pagina: Page, correo: string, contrasena: string) {
  await pagina.getByLabel("Correo").fill(correo)
  await pagina.getByLabel("Contraseña", { exact: true }).fill(contrasena)
  await pagina.getByRole("button", { name: "Entrar" }).click()
}

test.beforeEach(async ({ page }) => {
  await page.emulateMedia({ reducedMotion: "reduce" })
  await page.goto("/")
  // Sin sesión, la app pasa de sesion-actual (401) al formulario.
  await expect(page.getByRole("heading", { name: "Iniciar sesión" })).toBeVisible()
})

test("se puede recorrer el formulario con el teclado", async ({ page }) => {
  await capturar(page, "01-inicio-sesion")
  await page.getByLabel("Correo").focus()
  await page.keyboard.press("Tab")
  await expect(page.getByLabel("Contraseña", { exact: true })).toBeFocused()
  await page.keyboard.press("Tab")
  await expect(page.getByRole("button", { name: "Mostrar contraseña" })).toBeFocused()
  await page.keyboard.press("Tab")
  await expect(page.getByRole("button", { name: "Entrar" })).toBeFocused()

  await page.keyboard.press("Enter")
  await expect(page.getByText("Escribe tu correo.")).toBeVisible()
  await expect(page.getByText("Escribe tu contraseña.")).toBeVisible()
  await capturar(page, "02-validaciones")
})

test("con credenciales malas muestra un error genérico", async ({ page }) => {
  await entrar(page, "admin@dev.local", "no-es-la-clave")
  await expect(page.getByRole("alert")).toContainText(
    "El correo o la contraseña no son correctos",
  )
  await expect(page.getByLabel("Contraseña", { exact: true })).toHaveValue("")
  await capturar(page, "03-credenciales-malas")
})

test("una cuenta bloqueada no puede entrar", async ({ page }) => {
  test.skip(!conMocks, "La cuenta bloqueada existe solo en los mocks MSW")
  await entrar(page, "bloqueado@dev.local", "bloqueado-dev-local")
  await expect(page.getByRole("alert")).toContainText("Tu cuenta no está activa")
  await capturar(page, "04-cuenta-bloqueada")
})

test("entra, ve su sesión y la cierra", async ({ page }) => {
  await entrar(page, "admin@dev.local", "admin-dev-local")

  await expect(page.getByRole("heading", { name: "admin@dev.local" })).toBeVisible()
  await expect(page.getByText("Cuenta activa")).toBeVisible()
  await expect(page.getByRole("region", { name: "Roles" })).toContainText("administrador")
  await expect(page.getByRole("region", { name: "Permisos" })).toContainText(
    "usuarios:administrar",
  )
  await capturar(page, "05-sesion-actual")

  if (!conMocks) {
    // Con el backend real la sesión vive en la cookie HttpOnly: sobrevive a recargar.
    const cookies = await page.context().cookies()
    const sesion = cookies.find((cookie) => cookie.name === "sesion")
    expect(sesion).toMatchObject({ httpOnly: true, secure: true, sameSite: "Lax" })
    await page.reload()
    await expect(page.getByRole("heading", { name: "admin@dev.local" })).toBeVisible()
  }

  await page.getByRole("button", { name: "Cerrar sesión" }).click()
  await expect(page.getByRole("heading", { name: "Iniciar sesión" })).toBeVisible()
})

test("se ve bien en un celular", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await expect(page.getByRole("button", { name: "Entrar" })).toBeVisible()
  const desborda = await page.evaluate(
    () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
  )
  expect(desborda).toBe(false)
  await capturar(page, "06-movil")
})
