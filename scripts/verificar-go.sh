#!/bin/sh
# Revisa el formato (gofmt) y ejecuta go vet sobre la API.
# Lo usa el hook pre-commit de lefthook; funciona en Linux, macOS y Windows (Git Bash).
cd apps/api || exit 1

sin_formato=$(gofmt -l .)
if [ -n "$sin_formato" ]; then
  echo "Archivos Go sin formato. Ejecuta: cd apps/api; gofmt -w ."
  echo "$sin_formato"
  exit 1
fi

go vet ./...
