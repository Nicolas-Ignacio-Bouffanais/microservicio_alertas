package models

import "time"

type Camion struct {
	Id             int64     `json:"id" db:"id"`
	Patente        string    `json:"patente" db:"patente"`
	CoordenadasWKT *string   `json:"coordenadas_wkt,omitempty" db:"coordenadas_wkt"` // Para recibir "POINT (lon lat)"
	Velocidad      float64   `json:"velocidad" db:"velocidad"`                       // en km/h
	Orientacion    float64   `json:"orientacion" db:"orientacion"`                   // grados de 0 a 360 (mapeado desde 'orientacion')
	FechaHoraGPS   time.Time `json:"fecha_hora_gps" db:"fecha_hora_gps"`
}
