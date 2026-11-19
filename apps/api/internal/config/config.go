// Paquete config carga y valida la configuración del servicio desde variables de entorno.
package config

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

// Configuracion reúne todas las variables de entorno que necesita el servicio.
type Configuracion struct {
	Puerto      string // API_PORT: puerto interno, siempre 8080
	Ambiente    string // API_ENV: dev, test o prod
	DatabaseURL string // DATABASE_URL: conexión a la base propia
	FrontendURL string // FRONTEND_URL: origen permitido y base de enlaces
	ZonaHoraria string // TZ: America/Santiago
}

// ambientesValidos son los únicos valores aceptados para API_ENV.
var ambientesValidos = []string{"dev", "test", "prod"}

// Cargar lee las variables de entorno y devuelve un error si falta alguna obligatoria
// o si tiene un valor inválido. El proceso debe terminar si Cargar falla.
func Cargar() (Configuracion, error) {
	cfg := Configuracion{
		Puerto:      os.Getenv("API_PORT"),
		Ambiente:    os.Getenv("API_ENV"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		FrontendURL: os.Getenv("FRONTEND_URL"),
		ZonaHoraria: os.Getenv("TZ"),
	}

	var faltantes []string
	if cfg.Puerto == "" {
		faltantes = append(faltantes, "API_PORT")
	}
	if cfg.Ambiente == "" {
		faltantes = append(faltantes, "API_ENV")
	}
	if cfg.DatabaseURL == "" {
		faltantes = append(faltantes, "DATABASE_URL")
	}
	if cfg.FrontendURL == "" {
		faltantes = append(faltantes, "FRONTEND_URL")
	}
	if cfg.ZonaHoraria == "" {
		faltantes = append(faltantes, "TZ")
	}
	if len(faltantes) > 0 {
		return Configuracion{}, fmt.Errorf("faltan variables de entorno obligatorias: %s", strings.Join(faltantes, ", "))
	}

	if !slices.Contains(ambientesValidos, cfg.Ambiente) {
		return Configuracion{}, errors.New("API_ENV debe ser dev, test o prod")
	}
	return cfg, nil
}

// EsProduccion indica si el servicio corre en producción, donde no se publica la documentación.
func (c Configuracion) EsProduccion() bool {
	return c.Ambiente == "prod"
}
