package models

import (
	"time"
)

type Criticidad string

const (
	Leve  Criticidad = "LEVE"
	MEDIA Criticidad = "MEDIA"
	Grave Criticidad = "GRAVE"
)

type Evento struct {
	ID                int64     `db:"id"`
	IDConcentrador    int64     `db:"id_concentrador"`
	Patente           string    `db:"patente"`
	IDGeocerca        *int      `db:"id_geocerca"`
	CoordenadasWKT    *string   `db:"coordenadas_wkt"`
	FechaHoraInsert   time.Time `db:"fecha_hora_insert"`
	FechaHoraRegistro time.Time `db:"fecha_hora_registro"`
	FechaHoraCalc     time.Time `db:"fecha_hora_calc"`
	FechaHoraGps      time.Time `db:"fecha_hora_gps"`
	Velocidad         float64   `db:"velocidad"`
	Orientacion       float64   `db:"orientacion"`
}

type PreEventoExVelocidad struct {
	ID                  int64     `db:"id"`
	IDConcentrador      int64     `db:"id_concentrador"`
	Patente             string    `db:"patente"`
	IDGeocerca          *int      `db:"id_geocerca"`
	CoordenadasWKT      *string   `db:"coordenadas_wkt"`
	FechaHoraInsert     time.Time `db:"fecha_hora_insert"`
	FechaHoraRegistro   time.Time `db:"fecha_hora_registro"`
	FechaHoraCalc       time.Time `db:"fecha_hora_calc"`
	FechaHoraGps        time.Time `db:"fecha_hora_gps"`
	VelocidadRegistrada float64   `db:"velocidad_registrada"`
	VelocidadLimite     float64   `db:"velocidad_limite"`
	ExcesoRegistrado    float64   `db:"exceso_registrado"`
	Criticidad          Criticidad
	Orientacion         float64 `db:"orientacion"`
}

type PreEventoZonaRoja struct {
	ID                int64      `db:"id"`
	IDConcentrador    int64      `db:"id_concentrador"`
	Patente           string     `db:"patente"`
	IDGeocerca        *int       `db:"id_geocerca"`
	CoordenadasWKT    *string    `db:"coordenadas_wkt"`
	FechaHoraInsert   time.Time  `db:"fecha_hora_insert"`
	FechaHoraRegistro time.Time  `db:"fecha_hora_registro"`
	FechaHoraCalc     time.Time  `db:"fecha_hora_calc"`
	FechaHoraGps      time.Time  `db:"fecha_hora_gps"`
	Velocidad         float64    `db:"velocidad"`
	Orientacion       float64    `db:"orientacion"`
	Criticidad        Criticidad `db:"criticidad"`
}

type PreEventoDetNoAutorizada struct {
	IDConcentrador    int64      `db:"id_concentrador"`
	Patente           string     `db:"patente"`
	CoordenadasWKT    *string    `db:"coordenadas_wkt"`
	FechaHoraInsert   time.Time  `db:"fecha_hora_insert"`
	FechaHoraRegistro time.Time  `db:"fecha_hora_registro"`
	FechaHoraCalc     time.Time  `db:"fecha_hora_calc"`
	FechaHoraGps      time.Time  `db:"fecha_hora_gps"`
	Orientacion       float64    `db:"orientacion"`
	Criticidad        Criticidad `db:"criticidad"`
}
