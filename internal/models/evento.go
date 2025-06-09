package models

import (
	"time"
)

type Evento struct {
	ID             int64     `db:"id"`
	Patente        string    `db:"patente"`
	IDGeocerca     *string   `db:"id_geocerca"`
	CoordenadasWKT *string   `db:"coordenadas_wkt"`
	FechaHora      time.Time `db:"fecha_hora"`
	Velocidad      float64   `db:"velocidad"`
	Orientacion    float64   `db:"orientacion"`
}
