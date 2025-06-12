package models

type Rutas struct {
	IDRuta      int     `db:"id_Ruta"`
	Descripcion string  `db:"descripcion"`
	TipoRuta    *string `db:"tipo_Ruta"`
	Geometria   *string `db:"geometria"`
	VelUmbral   *int    `db:"vel_umbral"` // Puede ser NULL
	Radio       int     `db:"radio"`
}
