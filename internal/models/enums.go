// internal/models/enums.go
package models

type EstadoProcesamiento string // <-- Cambio de int a string

const (
	NoProcesado EstadoProcesamiento = "no_procesado"
	Marcado     EstadoProcesamiento = "marcado"
	Error       EstadoProcesamiento = "error"
	Procesado   EstadoProcesamiento = "procesado"
)
