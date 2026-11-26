// Generado por scripts/generar-mocks.mjs desde contracts/openapi.yaml. No editar a mano:
// se regenera con `pnpm codegen`.
import type { components } from "@/api/esquema"

type Esquemas = components["schemas"]

export const ejemplos = {
  AsignacionRoles: {
    "roles": [
      "11111111-1111-4111-8111-111111111111"
    ]
  },
  CambiosRol: {},
  CambiosUsuario: {
    "correo": "ana@ejemplo.cl"
  },
  ClaimsToken: {
    "aud": [
      "pingeso"
    ],
    "exp": 0,
    "iat": 0,
    "iss": "pingeso-identidad",
    "jti": "00000000-0000-4000-8000-000000000000",
    "nbf": 0,
    "permisos": [
      "perfiles:operar",
      "reservas:operar"
    ],
    "roles": [
      "operador"
    ],
    "sub": "00000000-0000-4000-8000-000000000000"
  },
  DatosPerfil: {
    "apellido": "Pérez",
    "foto_url": "https://boveda.example.cl/archivos/foto.jpg",
    "nombre": "Ana",
    "telefono": "+56912345678"
  },
  ErrorDetail: {},
  ErrorModel: {
    "detail": "Property foo is required but is missing.",
    "instance": "https://example.com/error-log/abc123",
    "status": 400,
    "title": "Bad Request",
    "type": "https://example.com/errors/example"
  },
  JWK: {
    "alg": "RS256",
    "e": "AQAB",
    "kid": "texto",
    "kty": "RSA",
    "n": "texto",
    "use": "sig"
  },
  JWKS: {
    "keys": [
      {
        "alg": "RS256",
        "e": "AQAB",
        "kid": "texto",
        "kty": "RSA",
        "n": "texto",
        "use": "sig"
      }
    ]
  },
  ListaPermisos: {
    "datos": [
      {
        "codigo": "usuarios:administrar",
        "descripcion": "Crear, editar, desactivar y listar usuarios"
      }
    ]
  },
  ListaRoles: {
    "datos": [
      {
        "actualizado_en": "2026-01-15T12:00:00Z",
        "creado_en": "2026-01-15T12:00:00Z",
        "descripcion": "Administra usuarios, roles y perfiles",
        "id": "11111111-1111-4111-8111-111111111111",
        "nombre": "administrador",
        "permisos": [
          "perfiles:operar",
          "roles:administrar",
          "usuarios:administrar"
        ]
      }
    ]
  },
  ListaUsuarios: {
    "datos": [
      {
        "actualizado_en": "2026-01-15T12:00:00Z",
        "correo": "admin@dev.local",
        "creado_en": "2026-01-15T12:00:00Z",
        "estado": "activo",
        "id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
        "roles": [
          {
            "id": "11111111-1111-4111-8111-111111111111",
            "nombre": "administrador"
          }
        ]
      }
    ],
    "paginacion": {
      "limite": 20,
      "pagina": 1,
      "total": 57,
      "total_paginas": 3
    }
  },
  Paginacion: {
    "limite": 20,
    "pagina": 1,
    "total": 57,
    "total_paginas": 3
  },
  Perfil: {
    "actualizado_en": "2026-01-15T12:00:00Z",
    "apellido": "Pérez",
    "atributos": {},
    "foto_url": "https://boveda.example.cl/archivos/foto.jpg",
    "nombre": "Ana",
    "telefono": "+56912345678",
    "usuario_id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
  },
  Permiso: {
    "codigo": "usuarios:administrar",
    "descripcion": "Crear, editar, desactivar y listar usuarios"
  },
  Renovacion: {
    "expira_en": "2026-01-15T12:00:00Z"
  },
  Rol: {
    "actualizado_en": "2026-01-15T12:00:00Z",
    "creado_en": "2026-01-15T12:00:00Z",
    "descripcion": "Administra usuarios, roles y perfiles",
    "id": "11111111-1111-4111-8111-111111111111",
    "nombre": "administrador",
    "permisos": [
      "perfiles:operar",
      "roles:administrar",
      "usuarios:administrar"
    ]
  },
  RolResumen: {
    "id": "11111111-1111-4111-8111-111111111111",
    "nombre": "administrador"
  },
  RolesAsignados: {
    "roles": [
      {
        "id": "11111111-1111-4111-8111-111111111111",
        "nombre": "administrador"
      }
    ]
  },
  SaludRespuestaBody: {
    "estado": "ok"
  },
  SolicitudCambioContrasena: {
    "contrasena_actual": "admin-dev-local",
    "contrasena_nueva": "una-clave-nueva-segura"
  },
  SolicitudCrearRol: {
    "descripcion": "Atiende la recepción",
    "nombre": "recepcionista",
    "permisos": [
      "usuarios:administrar"
    ]
  },
  SolicitudCrearUsuario: {
    "contrasena": "una-clave-segura",
    "correo": "ana@ejemplo.cl"
  },
  SolicitudSesion: {
    "contrasena": "admin-dev-local",
    "correo": "admin@dev.local"
  },
  SolicitudTokenPrueba: {
    "duracion": "2h",
    "roles": [
      "operador"
    ],
    "usuario_id": "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
  },
  TokenPrueba: {
    "claims": {
      "aud": [
        "pingeso"
      ],
      "exp": 0,
      "iat": 0,
      "iss": "pingeso-identidad",
      "jti": "00000000-0000-4000-8000-000000000000",
      "nbf": 0,
      "permisos": [
        "perfiles:operar",
        "reservas:operar"
      ],
      "roles": [
        "operador"
      ],
      "sub": "00000000-0000-4000-8000-000000000000"
    },
    "expira_en": "2026-01-15T12:00:00Z",
    "token": "texto"
  },
  Usuario: {
    "actualizado_en": "2026-01-15T12:00:00Z",
    "correo": "admin@dev.local",
    "creado_en": "2026-01-15T12:00:00Z",
    "estado": "activo",
    "id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
    "roles": [
      {
        "id": "11111111-1111-4111-8111-111111111111",
        "nombre": "administrador"
      }
    ]
  },
  UsuarioActual: {
    "actualizado_en": "2026-01-15T12:00:00Z",
    "correo": "admin@dev.local",
    "creado_en": "2026-01-15T12:00:00Z",
    "estado": "activo",
    "expira_en": "2026-01-15T12:00:00Z",
    "id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
    "permisos": [
      "perfiles:operar",
      "roles:administrar",
      "usuarios:administrar"
    ],
    "roles": [
      {
        "id": "11111111-1111-4111-8111-111111111111",
        "nombre": "administrador"
      }
    ]
  },
} satisfies { [K in keyof Esquemas]: Esquemas[K] }
