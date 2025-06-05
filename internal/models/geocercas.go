package models

type Geocerca struct {
	IDGeocerca int `db:"id_geocerca"`
	//cut			string `db:"cut"`
	Descripcion  string  `db:"descripcion"`
	TipoGeocerca *string `db:"tipo_geocerca"`

	Geometria *string `db:"geometria"`

	BoundingXMin float64 `db:"bounding_xmin"`
	BoundingXMax float64 `db:"bounding_xmax"`
	BoundingYMin float64 `db:"bounding_ymin"`
	BoundingYMax float64 `db:"bounding_ymax"`

	VelUmbral *int `db:"vel_umbral"` // Puede ser NULL
	TiempoMax *int `db:"tiempo_Max"` // Puede ser NULL
}
