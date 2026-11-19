# Modos de falla — MS Usuarios y Roles

Cada falla posible del servicio, qué la impide y cómo se verifica. Este documento se escribe
antes de programar cada parte y se actualiza cuando aparece una falla nueva.

Estado: **P** = prevención implementada y probada, **D** = diseñada, falta implementarla.

## 1. Disponibilidad

| ID    | Falla                                    | Qué la impide                                                                          | Cómo se verifica                                                                | Estado |
| ----- | ---------------------------------------- | -------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- | ------ |
| MF-01 | PostgreSQL caído o sin red               | `/readyz` responde 503 (problem+json) si el ping no contesta en 2 s                    | `TestReadyzConBaseCaida` y prueba manual deteniendo el contenedor               | P      |
| MF-02 | Falta una variable de entorno            | La configuración se valida al arrancar y el proceso termina con un error que la nombra | `TestCargarFallaSiFaltanVariables`, `TestLevantarServidorFallaSinConfiguracion` | P      |
| MF-03 | La API arranca antes que las migraciones | Compose: `backend` espera `service_completed_successfully` de `migrate`                | `docker compose up` desde cero                                                  | P      |
| MF-04 | Consulta lenta deja la petición colgada  | Timeout de 5 s en cada llamada a la base y `ReadHeaderTimeout` en el servidor          | e2e con consulta bloqueada (Hito 2)                                             | D      |

## 2. Autenticación

| ID    | Falla                                           | Qué la impide                                                                     | Cómo se verifica                                                 | Estado |
| ----- | ----------------------------------------------- | --------------------------------------------------------------------------------- | ---------------------------------------------------------------- | ------ |
| MF-10 | Credenciales incorrectas                        | Respuesta 401 genérica: no revela si existe el correo                             | e2e: correo inexistente y contraseña mala dan la misma respuesta | D      |
| MF-11 | Ataque de fuerza bruta                          | Bloqueo temporal tras N intentos fallidos por correo                              | e2e: N+1 intentos fallidos devuelven 429                         | D      |
| MF-12 | Diferencia de tiempo revela si existe el correo | Se calcula el hash aunque el correo no exista                                     | Prueba de tiempos en el e2e de login                             | D      |
| MF-13 | Token alterado o firmado con otra llave         | Verificación de firma RS256/EdDSA con la llave pública y `kid`                    | Prueba con token modificado: 401                                 | D      |
| MF-14 | Token vencido                                   | Validación de `exp` en el middleware                                              | Prueba con token vencido: 401                                    | D      |
| MF-15 | Robo del token desde el navegador               | Cookie `HttpOnly`, `Secure`, `SameSite=Lax`. Prohibido `localStorage`             | Prueba de cabecera `Set-Cookie` en el login                      | D      |
| MF-16 | Llave privada filtrada                          | La llave entra solo por secreto. Rotación con nuevo `kid` publicado en el JWKS    | Procedimiento documentado y prueba de rotación                   | D      |
| MF-17 | Otros servicios no pueden leer el JWKS          | `/.well-known/jwks.json` sin autenticación; los consumidores lo cachean (keyfunc) | e2e del endpoint JWKS                                            | D      |

## 3. Autorización y datos

| ID    | Falla                                                | Qué la impide                                                          | Cómo se verifica                                    | Estado |
| ----- | ---------------------------------------------------- | ---------------------------------------------------------------------- | --------------------------------------------------- | ------ |
| MF-20 | Usuario sin permiso accede a un recurso              | Cada endpoint declara su permiso; el middleware responde 403           | Matriz de acceso por rol en `acceso_test.go`        | D      |
| MF-21 | Usuario se asigna a sí mismo un rol con más permisos | Asignar roles requiere `usuarios:administrar`                          | `acceso_test.go`                                    | D      |
| MF-22 | Dos registros simultáneos con el mismo correo        | Restricción `UNIQUE` en `usuarios.correo`; el segundo recibe 409       | e2e concurrente: dos POST a la vez, uno debe fallar | D      |
| MF-23 | Se elimina un rol que tienen usuarios                | El servicio rechaza con 409 si el rol está asignado                    | e2e                                                 | D      |
| MF-24 | Se elimina o bloquea al último administrador         | Regla de negocio: siempre debe quedar al menos un administrador activo | e2e                                                 | D      |
| MF-25 | Datos de entrada inválidos                           | Validación en el contrato con Huma (tipos, `minLength`, `pattern`)     | Pruebas de cada endpoint con datos inválidos: 422   | D      |

## 4. Información sensible

| ID    | Falla                                | Qué la impide                                                                   | Cómo se verifica                                    | Estado |
| ----- | ------------------------------------ | ------------------------------------------------------------------------------- | --------------------------------------------------- | ------ |
| MF-30 | Contraseñas guardadas en texto plano | Solo se guarda el hash (`golang.org/x/crypto`)                                  | Prueba: la tabla nunca contiene la contraseña       | D      |
| MF-31 | Contraseñas o tokens en los logs     | Logs con `log/slog`: nunca se registran contraseñas, tokens, cookies ni cuerpos | Revisión de código y prueba sobre la salida de logs | D      |
| MF-32 | Errores que exponen trazas o SQL     | Errores en formato RFC 9457 con mensajes genéricos                              | `TestReadyzConBaseCaida` revisa `problem+json`      | P      |
| MF-33 | Secretos subidos al repositorio      | Gitleaks en pre-commit y en el pipeline                                         | Commit de prueba con un secreto falso es bloqueado  | D      |
