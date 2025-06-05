package models

import "time"

type Camion struct {
	Patente      string    `json:"patente" db:"patente"`
	Latitud      float64   `json:"latitud" db:"latitud"`         // Corresponde a ST_Y(coordenadas) en PostGIS
	Longitud     float64   `json:"longitud" db:"longitud"`       // Corresponde a ST_X(coordenadas) en PostGIS
	Velocidad    float64   `json:"velocidad" db:"velocidad"`     // en km/h
	Orientacion  float64   `json:"orientacion" db:"orientacion"` // grados de 0 a 360 (mapeado desde 'orientacion')
	FechaHoraGPS time.Time `json:"fecha_hora_gps" db:"fecha_hora_gps"`
}
