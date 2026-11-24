# MS Usuarios y Roles

Servicio base de autenticación, usuarios, roles y perfiles. Emite los JWT que validan los demás servicios.

## Evidencia

- Pipeline: https://github.com/PINGESO-CIVIL/usuarios-roles/actions
- SonarQube: https://sonarcloud.io/project/overview?id=CatalinaCovarrubias_pingeso-grupo6-identity-provider
- Contrato: [`contracts/openapi.yaml`](contracts/openapi.yaml)
- Modos de falla: [`docs/modos-de-falla.md`](docs/modos-de-falla.md)
- JWT de prueba para los otros grupos: [`docs/jwt-de-prueba.md`](docs/jwt-de-prueba.md)

## Puertos locales

| Componente | URL                               |
| ---------- | --------------------------------- |
| API        | http://localhost:5161             |
| Docs API   | http://localhost:5161/api/docs    |
| Web        | http://localhost:5160 (pendiente) |

## Puesta en marcha

```bash
pnpm install
pnpm exec lefthook install
cp .env.example .env   # completar los valores replace-with-*
docker compose up -d --build
```

## Comandos

| Comando         | Qué hace                                               |
| --------------- | ------------------------------------------------------ |
| `pnpm codegen`  | Regenera `contracts/openapi.yaml` desde los structs Go |
| `pnpm test:api` | Pruebas de la API                                      |
| `pnpm verify`   | Formato, lint y `go vet`                               |
