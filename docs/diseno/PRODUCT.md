# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

Personas que entran a alguno de los productos de los grupos cliente del curso PINGESO
(sección Civil). Según el producto, puede ser personal interno de una organización o
público general. No sabemos cuál en cada caso, así que el diseño no supone ni lo uno ni lo
otro: no asume alguien experto ni que use la herramienta a diario, pero tampoco a alguien que
la use por primera vez.

Su tarea es corta y siempre la misma: entrar con correo y contraseña, ver con qué cuenta y
qué permisos entró, y salir.

Los consumidores directos son los equipos de los grupos cliente, que montan estos
microcomponentes en sus propios productos y les ajustan el aspecto con tokens.

## Product Purpose

Servicio de identidad común (Grupo 6): emite la sesión en un JWT firmado con RS256 que viaja
en una cookie `HttpOnly`, y entrega dos microcomponentes React, `inicio-sesion` y
`sesion-actual`, para que cada producto no tenga que construir su propio login. Funcionan si
alguien entra a la primera, entiende por qué no pudo entrar cuando falla y sabe con qué cuenta
quedó la sesión.

## Positioning

Es la puerta común de varios productos distintos. Tiene que sentirse propia en cualquiera de
ellos (por eso la marca la ponen los tokens del cliente) sin dejar de transmitir que es un
lugar seguro donde escribir la contraseña.

## Operating Context

- Se usa en navegadores de escritorio y de celular.
- Va dentro de un producto ajeno, sin layout ni rutas propias: el microcomponente recibe por
  props la URL de la API y avisa con callbacks (`alEntrar`, `alSalir`).
- Errores RFC 9457: 401 genérico para credenciales malas (no se dice si el correo existe),
  403 para una cuenta no activa y 422 para datos que no cumplen el contrato.

## Capabilities and Constraints

- Login con correo (máximo 254) y contraseña (1 a 72 caracteres, y como máximo 72 bytes).
- La sesión actual muestra el correo, el estado, los roles, los permisos y cuándo vence.
- Cerrar sesión borra la cookie.
- Prohibidos `localStorage` y `sessionStorage`. Nada secreto en variables `VITE_*`.
- Textos en español de Chile.
- Solo modo claro por ahora, con los tokens preparados para agregar modo oscuro después.
- Sin decidir: registro de usuarios, recuperación de contraseña y gestión de usuarios y
  roles (van en otra tanda). Ninguna interfaz debe ofrecerlos hasta que existan.

## Brand Commitments

Sin marca propia ni del curso. Base neutra pero con criterio: fresca, moderna, confiable y
sobria. Todos los valores visuales están expuestos como variables CSS que un grupo cliente
puede sobrescribir.

## Evidence on Hand

- Usuarios seed de desarrollo: `admin@dev.local`, `operador@dev.local` y `usuario@dev.local`,
  con roles y permisos reales del contrato (`docs/jwt-de-prueba.md`).
- Contrato en `contracts/openapi.yaml`.
- No hay logotipos, fotografías ni testimonios, y no se deben inventar.

## Product Principles

1. Seguridad visible y calmada: el usuario siente que está en un lugar serio sin que la
   interfaz grite "seguridad".
2. Cada error dice qué pasó y qué hacer, sin filtrar información (el 401 no revela si el
   correo existe).
3. Se adapta al producto anfitrión: los tokens mandan y el componente no impone identidad.
4. Nada de funciones que no existen todavía: sin enlaces de "olvidé mi contraseña" ni de
   registro.

## Accessibility & Inclusion

WCAG 2.2 AA: labels visibles, foco visible, contraste AA, navegable con teclado, errores
anunciados a lectores de pantalla y uso cómodo en pantallas pequeñas.
