package models

import (
	"time"
)

type Evento struct {
	ID                int64     `db:"id"`
	IDConcentrador    int64     `db:"id_concentrador"`
	Patente           string    `db:"patente"`
	IDGeocerca        *string   `db:"id_geocerca"`
	CoordenadasWKT    *string   `db:"coordenadas_wkt"`
	FechaHoraInsert   time.Time `db:"fecha_hora_insert"`
	FechaHoraRegistro time.Time `db:"fecha_hora_registro"`
	FechaHoraCalc     time.Time `db:"fecha_hora_calc"`
	FechaHoraGps      time.Time `db:"fecha_hora_gps"`
	Velocidad         float64   `db:"velocidad"`
	Orientacion       float64   `db:"orientacion"`
}
