---
name: Identidad · microcomponentes de sesión
description: Base neutra y confiable para el login común de los productos del curso, pensada para que cada cliente la tiña con sus tokens.
colors:
  tinta-sobre: "oklch(0.46 0.15 264)"
  tinta-sobre-honda: "oklch(0.39 0.13 265)"
  tinta-anillo: "oklch(0.54 0.16 263)"
  tinta-velo: "oklch(0.97 0.016 262)"
  tinta-trama: "oklch(0.87 0.065 262)"
  papel-frio: "oklch(0.985 0.004 262)"
  tarjeta: "oklch(1 0 0)"
  grafito: "oklch(0.21 0.024 262)"
  grafito-suave: "oklch(0.47 0.027 262)"
  borde-control: "oklch(0.555 0.025 262)"
  borde-quieto: "oklch(0.925 0.012 262)"
  peligro-texto: "oklch(0.38 0.14 25)"
  peligro-fondo: "oklch(0.97 0.014 20)"
  advertencia-texto: "oklch(0.41 0.095 60)"
  advertencia-fondo: "oklch(0.98 0.03 92)"
  exito-texto: "oklch(0.38 0.09 155)"
  exito-fondo: "oklch(0.975 0.02 155)"
typography:
  headline:
    fontFamily: "var(--fuente-texto), system-ui, sans-serif"
    fontSize: "1.5rem"
    fontWeight: 600
    lineHeight: 1.33
  title:
    fontFamily: "var(--fuente-texto), system-ui, sans-serif"
    fontSize: "1.25rem"
    fontWeight: 600
    lineHeight: 1.4
  body:
    fontFamily: "var(--fuente-texto), system-ui, sans-serif"
    fontSize: "1rem"
    fontWeight: 400
    lineHeight: 1.5
  label:
    fontFamily: "var(--fuente-texto), system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 600
    lineHeight: 1.43
  data:
    fontFamily: "var(--fuente-mono), ui-monospace, monospace"
    fontSize: "0.875rem"
    fontWeight: 400
    lineHeight: 1.43
    fontFeature: "tnum"
rounded:
  sm: "4px"
  md: "8px"
  pill: "9999px"
spacing:
  xs: "4px"
  sm: "8px"
  md: "16px"
  lg: "24px"
  xl: "32px"
components:
  button-primary:
    backgroundColor: "{colors.tinta-sobre}"
    textColor: "{colors.tarjeta}"
    rounded: "{rounded.md}"
    height: "40px"
    padding: "0 16px"
  button-primary-hover:
    backgroundColor: "{colors.tinta-sobre-honda}"
  button-secondary:
    backgroundColor: "{colors.tarjeta}"
    textColor: "{colors.grafito}"
    rounded: "{rounded.md}"
    height: "40px"
    padding: "0 16px"
  input:
    backgroundColor: "{colors.tarjeta}"
    textColor: "{colors.grafito}"
    rounded: "{rounded.md}"
    height: "40px"
    padding: "0 12px"
  card-sobre:
    backgroundColor: "{colors.tarjeta}"
    textColor: "{colors.grafito}"
    rounded: "{rounded.md}"
    width: "420px"
    padding: "32px"
  badge-rol:
    backgroundColor: "{colors.tinta-velo}"
    textColor: "{colors.tinta-sobre-honda}"
    rounded: "{rounded.pill}"
    height: "24px"
  alert-peligro:
    backgroundColor: "{colors.peligro-fondo}"
    textColor: "{colors.peligro-texto}"
    rounded: "{rounded.md}"
---

# Design System: Identidad · microcomponentes de sesión

La fuente de verdad son los tokens de `apps/web/src/estilos/tema.css`. Este documento
explica qué significan y cómo usarlos; si algún valor no coincide, manda el CSS.

## Overview

**Idea rectora: el sobre de seguridad.** Por dentro, los sobres que llevan algo
confidencial traen impreso un tramado fino de líneas que impide leer el contenido a
contraluz. Ese tramado es el único gesto propio del sistema: aparece en la banda superior
de la tarjeta, justo donde se guarda algo privado, y en ningún otro lugar.

Todo lo demás es deliberadamente familiar: un formulario que cualquiera sabe usar, con
labels visibles, un solo botón principal y errores en frases simples. La sensación buscada
es la de un trámite bien hecho: sobrio, claro y tranquilo. Es una base neutra porque cada
grupo cliente le pone su marca sobrescribiendo variables CSS; el sistema tiene que verse
bien con la tinta por defecto y también cuando alguien la cambia por su propio color.

Se evitan a propósito: vidrio y desenfoques, neón sobre fondo oscuro, degradados en el
texto y la estética "hacker" que suele acompañar a todo lo que dice "seguridad".

## Colors

Estrategia contenida: neutros fríos teñidos con el mismo tono de la tinta, un único color
de acción y colores de estado que solo aparecen cuando significan algo.

### Primary

- **Tinta de sobre** (`oklch(0.46 0.15 264)`): el botón principal, el foco y la selección.
  Un azul tinta profundo, el color con el que se imprime el tramado de los sobres. Con texto
  blanco encima alcanza 7,3:1.
- **Tinta honda** (`oklch(0.39 0.13 265)`): el hover del botón principal y el texto de las
  insignias de rol.
- **Trama** (`oklch(0.87 0.065 262)`) sobre **velo de tinta** (`oklch(0.97 0.016 262)`): el
  patrón de la banda. Su contraste es bajo a propósito (1,4:1): es decoración y nunca lleva
  texto encima.

### Neutral

- **Papel frío** (`oklch(0.985 0.004 262)`): el fondo de la página, apenas azulado.
- **Tarjeta** (blanco): la superficie del sobre, de los inputs y del botón secundario.
- **Grafito** (`oklch(0.21 0.024 262)`): el texto principal (17,7:1 sobre la tarjeta).
- **Grafito suave** (`oklch(0.47 0.027 262)`): el texto secundario (6,8:1).
- **Borde de control** (`oklch(0.555 0.025 262)`): el contorno de inputs y botones
  secundarios. Tiene al menos 3:1 contra la tarjeta, como pide WCAG 1.4.11.
- **Borde quieto** (`oklch(0.925 0.012 262)`): los separadores que no son controles.

### Estados

Cada estado tiene un fondo tintado claro con texto oscuro del mismo tono, siempre
acompañado de un ícono y de una frase. El color nunca comunica solo: peligro (`peligro-*`),
advertencia (`advertencia-*`) y éxito (`exito-*`), todos sobre 8:1.

## Typography

Una sola familia para todo, más una monoespaciada para los datos. Los microcomponentes no
cargan fuentes: usan `var(--fuente-texto)` y `var(--fuente-mono)`, que por defecto son las
del sistema. La app de desarrollo carga Atkinson Hyperlegible Next y Atkinson Hyperlegible
Mono (en `estilos/fuentes.css`, que un cliente puede incluir si quiere). Se eligieron porque
están diseñadas para que ninguna letra se confunda con otra, algo que importa al escribir
un correo o leer un permiso.

- Escala cerrada en rem: 12 / 14 / 16 / 18 / 20 / 24 / 30. Nada intermedio.
- Dos pesos: 400 para leer y 600 para destacar. `font-medium` de shadcn también cae en 600.
- La monoespaciada se usa solo para datos que se copian o comparan: permisos
  (`roles:administrar`), horas con cifras tabulares y las credenciales de prueba. No se usa
  como adorno "técnico".

## Layout

- La tarjeta mide hasta 420 px de ancho y ocupa todo el ancho disponible en pantallas
  angostas. El microcomponente no impone layout: el host decide dónde va.
- El espaciado sale de la grilla de 4 px (4, 8, 12, 16, 24, 32). Dentro de un campo, label
  e input van a 8 px; entre campos hay 20-24 px; entre secciones, 24 px.
- Hay más espacio alrededor de un grupo que dentro de él, y más arriba de un título que
  debajo.
- El padding de la tarjeta es de 24 px en celulares y 32 px desde `sm`.

## Elevation & Depth

Una sola elevación con sombra: la tarjeta (`shadow-md`, de dos capas, con la luz desde
arriba y teñida con el grafito). La tarjeta no lleva borde: la elevación se declara una
sola vez. Los inputs se distinguen por su borde, no por una sombra. La escala tiene cuatro
pasos (`xs`, `sm`, `md`, `lg`) para cuando haya menús o diálogos.

## Shapes

Radio de 8 px en la tarjeta, los inputs, los botones y los avisos; 4 px para piezas
internas muy chicas. Las insignias (roles y estado de la cuenta) son las únicas píldoras:
son controles pequeños y de solo lectura. No se mezclan esquinas rectas con redondeadas.

## Components

- **Sobre** (`components/sobre.tsx`): la tarjeta con la banda tramada de 40 px arriba. En
  `inicio-sesion` el sobre está cerrado y la banda se ve completa. En `sesion-actual` el
  sobre está abierto: la solapa se recoge hacia arriba (`scaleY` a 0,35 en 180 ms con
  salida exponencial) y el contenido aparece con un leve ascenso. Solo usa transform y
  opacity, y con `prefers-reduced-motion` no hay animación.
- **Botón** (`components/boton.tsx`): 40 px de alto. El primario es tinta sólida y hay uno
  por vista. El secundario lleva borde de control sobre la tarjeta. Mientras carga muestra
  un spinner, cambia el texto ("Entrando…") y no se puede volver a pulsar.
- **Campo** (`components/campo.tsx`): label visible en 600, input de 40 px y texto de 16 px
  (así iOS no hace zoom). La ayuda y el error quedan enlazados con `aria-describedby`. Con
  error, el borde y el anillo pasan a peligro y el mensaje aparece debajo.
- **Aviso** (`components/aviso.tsx`): fondo tintado, ícono, título en 600, una frase que
  dice qué hacer y, si corresponde, una acción ("Reintentar").
- **Insignia** (`components/insignia.tsx`): roles en velo de tinta; el estado de la cuenta
  en éxito o advertencia, con ícono.
- **Estados**: la carga es un skeleton con la misma caja que el contenido final, para que
  nada salte. Sin sesión hay un ícono en un círculo tintado y una frase. Ante un error hay
  un aviso con la opción de reintentar.

## Do's and Don'ts

- **Sí:** cambiar la marca sobrescribiendo tokens semánticos (`--primary`,
  `--primary-hover`, `--patron-sobre-tinta`, `--fuente-texto`…) en un CSS cargado después
  de `tema.css`.
- **Sí:** mantener el tramado solo en la banda del sobre y siempre sin texto encima.
- **Sí:** escribir los errores como qué pasó y qué hacer. El 401 es siempre genérico.
- **No:** editar `src/components/ui/` (lo genera shadcn); las variantes van en los wrappers
  de `src/components/`.
- **No:** usar valores fuera de la escala (colores sueltos, `padding: 13px`, pesos 500 o
  700).
- **No:** agregar enlaces a funciones que no existen todavía, como "olvidé mi contraseña"
  o "crear cuenta".
- **No:** usar la monoespaciada para decorar ni el tramado como fondo de página.
