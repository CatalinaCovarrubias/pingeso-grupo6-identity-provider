// Comando api: arranque del servicio, configuración y manejo de señales.
//
// Uso:
//
//	api                     levanta el servidor HTTP
//	api openapi             imprime el contrato OpenAPI 3.1 (YAML) en la salida estándar
//	api openapi <archivo>   escribe el contrato OpenAPI 3.1 (YAML) en el archivo indicado
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/CatalinaCovarrubias/pingeso-grupo6-identity-provider/apps/api/internal/api"
	"github.com/CatalinaCovarrubias/pingeso-grupo6-identity-provider/apps/api/internal/config"
	"github.com/CatalinaCovarrubias/pingeso-grupo6-identity-provider/apps/api/internal/service"
	"github.com/CatalinaCovarrubias/pingeso-grupo6-identity-provider/apps/api/internal/service/tokens"
	"github.com/CatalinaCovarrubias/pingeso-grupo6-identity-provider/apps/api/internal/store/memoria"
)

// version se reemplaza al compilar con -ldflags "-X main.version=<sha>".
var version = "dev"

func main() {
	ctx, detener := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer detener()

	if err := ejecutar(ctx, os.Args[1:], os.Stdout); err != nil {
		slog.Error("el servicio terminó con error", "error", err)
		os.Exit(1)
	}
}

// ejecutar decide qué hacer según los argumentos. Está separada de main para poder probarla.
func ejecutar(ctx context.Context, args []string, salida io.Writer) error {
	if len(args) > 0 && args[0] == "openapi" {
		if len(args) > 1 {
			// Se escribe directo al archivo para que la terminal de Windows no cambie la codificación.
			return escribirContrato(args[1])
		}
		return imprimirContrato(salida)
	}
	return levantarServidor(ctx)
}

// imprimirContrato escribe el contrato OpenAPI generado desde los structs de internal/api.
func imprimirContrato(salida io.Writer) error {
	// El contrato incluye POST /api/dev/tokens (etiqueta Desarrollo), aunque en test y prod no exista.
	_, humaAPI := api.NuevoServidor(api.Dependencias{PublicarDocs: true, RegistrarDev: true})
	contrato, err := humaAPI.OpenAPI().YAML()
	if err != nil {
		return fmt.Errorf("no se pudo generar el contrato: %w", err)
	}
	_, err = salida.Write(contrato)
	return err
}

// escribirContrato guarda el contrato en un archivo, con permisos de lectura para todos.
func escribirContrato(ruta string) error {
	var contenido bytes.Buffer
	if err := imprimirContrato(&contenido); err != nil {
		return err
	}
	return os.WriteFile(ruta, contenido.Bytes(), 0o644)
}

// levantarServidor carga la configuración, conecta con PostgreSQL y atiende peticiones
// hasta que se cancele ctx (en main, al recibir SIGINT o SIGTERM).
func levantarServidor(ctx context.Context) error {
	cfg, err := config.Cargar()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("servicio", "usuarios-roles", "version", version)
	slog.SetDefault(logger)

	llave, err := obtenerLlave(cfg, logger)
	if err != nil {
		return err
	}
	logger.Info("llave de firma JWT lista", "kid", llave.Kid())
	emisor := tokens.NuevoEmisor(llave, cfg.JWTEmisor, cfg.JWTAudiencia, cfg.JWTTTL)

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("DATABASE_URL inválida: %w", err)
	}
	defer pool.Close()

	dep, err := armarDependencias(ctx, cfg, emisor, logger)
	if err != nil {
		return err
	}
	dep.BD = pool
	manejador, _ := api.NuevoServidor(dep)
	servidor := &http.Server{
		Addr:              ":" + cfg.Puerto,
		Handler:           manejador,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errores := make(chan error, 1)
	go func() {
		logger.Info("servidor escuchando", "puerto", cfg.Puerto, "ambiente", cfg.Ambiente)
		errores <- servidor.ListenAndServe()
	}()

	select {
	case err := <-errores:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		logger.Info("señal recibida, cerrando el servidor")
	}

	// Se parte de ctx sin su cancelación: ctx ya fue cancelado por la señal,
	// pero el cierre ordenado necesita su propio plazo de 10 segundos.
	ctxCierre, cancelar := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancelar()
	return servidor.Shutdown(ctxCierre)
}

// armarDependencias conecta los services con el store en memoria del Hito 1.
// El seed de desarrollo solo se carga en dev y test: en prod el almacén parte vacío.
func armarDependencias(ctx context.Context, cfg config.Configuracion, emisor *tokens.Emisor, logger *slog.Logger) (api.Dependencias, error) {
	almacen := memoria.Nueva()
	if cfg.EsProduccion() {
		logger.Warn("store en memoria sin datos seed: en prod nadie puede iniciar sesión hasta la Tanda 4")
	} else if err := memoria.CargarSeed(ctx, almacen, bcrypt.DefaultCost); err != nil {
		return api.Dependencias{}, err
	}

	sesiones, err := service.NuevoSesiones(almacen, almacen, emisor, bcrypt.DefaultCost, time.Now)
	if err != nil {
		return api.Dependencias{}, err
	}
	return api.Dependencias{
		PublicarDocs: !cfg.EsProduccion(),
		RegistrarDev: cfg.EsDesarrollo(),
		Sesiones:     sesiones,
		Desarrollo:   sesiones,
		Usuarios:     service.NuevoUsuarios(almacen, almacen, bcrypt.DefaultCost, time.Now),
		Roles:        service.NuevoRoles(almacen, almacen, time.Now),
		Perfiles:     service.NuevoPerfiles(almacen, almacen, time.Now),
	}, nil
}

// obtenerLlave carga la llave de firma desde JWT_LLAVE_PRIVADA_ARCHIVO. Solo en dev,
// si la variable falta, genera una llave en memoria y lo avisa en el log.
func obtenerLlave(cfg config.Configuracion, logger *slog.Logger) (*tokens.Llave, error) {
	if cfg.JWTLlavePrivadaArchivo != "" {
		return tokens.CargarLlave(cfg.JWTLlavePrivadaArchivo)
	}
	if !cfg.EsDesarrollo() {
		return nil, errors.New("JWT_LLAVE_PRIVADA_ARCHIVO es obligatoria fuera de dev")
	}
	llave, err := tokens.GenerarLlave()
	if err != nil {
		return nil, err
	}
	logger.Warn("JWT_LLAVE_PRIVADA_ARCHIVO no está definida: se generó una llave temporal en memoria; "+
		"los tokens dejan de valer al reiniciar el servicio", "kid", llave.Kid())
	return llave, nil
}
