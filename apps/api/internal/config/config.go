// Paquete config carga y valida la configuración del servicio desde variables de entorno.
package config

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"
)

// Límites aceptados para JWT_TTL.
const (
	ttlMinimo = time.Minute
	ttlMaximo = 24 * time.Hour
)

// Configuracion reúne todas las variables de entorno que necesita el servicio.
type Configuracion struct {
	Puerto      string // API_PORT: puerto interno, siempre 8080
	Ambiente    string // API_ENV: dev, test o prod
	DatabaseURL string // DATABASE_URL: conexión a la base propia
	FrontendURL string // FRONTEND_URL: origen permitido y base de enlaces
	ZonaHoraria string // TZ: America/Santiago

	JWTEmisor    string        // JWT_EMISOR: valor del claim iss
	JWTAudiencia string        // JWT_AUDIENCIA: valor del claim aud
	JWTTTL       time.Duration // JWT_TTL: duración de cada token, por ejemplo 15m
	// JWTLlavePrivadaArchivo es la ruta al PEM con la llave RSA de firma
	// (JWT_LLAVE_PRIVADA_ARCHIVO). Es obligatoria salvo en dev.
	JWTLlavePrivadaArchivo string
}

// ambientesValidos son los únicos valores aceptados para API_ENV.
var ambientesValidos = []string{"dev", "test", "prod"}

// Cargar lee las variables de entorno y devuelve un error si falta alguna obligatoria
// o si tiene un valor inválido. El proceso debe terminar si Cargar falla.
func Cargar() (Configuracion, error) {
	cfg := Configuracion{
		Puerto:                 os.Getenv("API_PORT"),
		Ambiente:               os.Getenv("API_ENV"),
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		FrontendURL:            os.Getenv("FRONTEND_URL"),
		ZonaHoraria:            os.Getenv("TZ"),
		JWTEmisor:              os.Getenv("JWT_EMISOR"),
		JWTAudiencia:           os.Getenv("JWT_AUDIENCIA"),
		JWTLlavePrivadaArchivo: os.Getenv("JWT_LLAVE_PRIVADA_ARCHIVO"),
	}
	ttl := os.Getenv("JWT_TTL")

	obligatorias := []struct{ nombre, valor string }{
		{"API_PORT", cfg.Puerto},
		{"API_ENV", cfg.Ambiente},
		{"DATABASE_URL", cfg.DatabaseURL},
		{"FRONTEND_URL", cfg.FrontendURL},
		{"TZ", cfg.ZonaHoraria},
		{"JWT_EMISOR", cfg.JWTEmisor},
		{"JWT_AUDIENCIA", cfg.JWTAudiencia},
		{"JWT_TTL", ttl},
	}
	var faltantes []string
	for _, v := range obligatorias {
		if v.valor == "" {
			faltantes = append(faltantes, v.nombre)
		}
	}
	// Solo en dev se permite arrancar sin llave: se genera una en memoria.
	if cfg.Ambiente != "dev" && cfg.JWTLlavePrivadaArchivo == "" {
		faltantes = append(faltantes, "JWT_LLAVE_PRIVADA_ARCHIVO")
	}
	if len(faltantes) > 0 {
		return Configuracion{}, fmt.Errorf("faltan variables de entorno obligatorias: %s", strings.Join(faltantes, ", "))
	}

	if !slices.Contains(ambientesValidos, cfg.Ambiente) {
		return Configuracion{}, errors.New("API_ENV debe ser dev, test o prod")
	}

	duracion, err := time.ParseDuration(ttl)
	if err != nil {
		return Configuracion{}, errors.New("JWT_TTL debe ser una duración, por ejemplo 15m")
	}
	if duracion < ttlMinimo || duracion > ttlMaximo {
		return Configuracion{}, fmt.Errorf("JWT_TTL debe estar entre %s y %s", ttlMinimo, ttlMaximo)
	}
	cfg.JWTTTL = duracion
	return cfg, nil
}

// EsProduccion indica si el servicio corre en producción, donde no se publica la documentación.
func (c Configuracion) EsProduccion() bool {
	return c.Ambiente == "prod"
}

// EsDesarrollo indica si el servicio corre en desarrollo local: solo ahí existe
// POST /api/dev/tokens y se permite generar la llave de firma en memoria.
func (c Configuracion) EsDesarrollo() bool {
	return c.Ambiente == "dev"
}
