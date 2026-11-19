-- Esquema inicial del servicio de usuarios y roles.
-- Una migración ya desplegada no se edita: los cambios van en una migración nueva.

-- +goose Up
CREATE TABLE usuarios (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    correo          text        NOT NULL UNIQUE,
    hash_contrasena text        NOT NULL,
    estado          text        NOT NULL DEFAULT 'activo' CHECK (estado IN ('activo', 'bloqueado')),
    creado_en       timestamptz NOT NULL DEFAULT now(),
    actualizado_en  timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE usuarios;
