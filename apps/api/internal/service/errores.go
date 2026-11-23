package service

import "errors"

// Causas de error del service. La capa HTTP las traduce a códigos de estado.
var (
	ErrNoEncontrado   = errors.New("no encontrado")
	ErrConflicto      = errors.New("conflicto")
	ErrValidacion     = errors.New("datos inválidos")
	ErrCredenciales   = errors.New("credenciales inválidas")
	ErrCuentaNoActiva = errors.New("cuenta no activa")
	ErrNoAutenticado  = errors.New("no autenticado")
	ErrSinPermiso     = errors.New("sin permiso")
)

// ErrorNegocio es un error con un mensaje apto para mostrar al cliente.
// Su causa es una de las variables de arriba.
type ErrorNegocio struct {
	causa   error
	mensaje string
}

func (e *ErrorNegocio) Error() string { return e.mensaje }

// Unwrap permite usar errors.Is contra la causa.
func (e *ErrorNegocio) Unwrap() error { return e.causa }

func nuevoError(causa error, mensaje string) error {
	return &ErrorNegocio{causa: causa, mensaje: mensaje}
}
