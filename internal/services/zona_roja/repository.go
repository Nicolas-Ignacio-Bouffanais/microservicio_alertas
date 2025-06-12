package zona_roja

import (
	"fmt"
	"log"

	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/database"
	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/models"
)

func GetInterseccionesZonaRoja(fecha string) ([]models.Evento, error) {
	// Accedemos a la configuración y la BD a través del paquete database
	if database.Cfg == nil || database.Cfg.TableNames.ConcentradorGPS == "" || database.Cfg.TableNames.Geocercas == "" {
		return nil, fmt.Errorf("configuración de nombres de tabla no cargada completamente")
	}
	if database.DB == nil {
		return nil, fmt.Errorf("conexión a la base de datos no inicializada")
	}

	query := fmt.Sprintf(`
		SELECT
			c.patente,
			g.id::text AS id_geocerca,
			ST_AsText(c.coordenadas) AS coordenadas_wkt,
			c.velocidad,
			c.orientacion,
			c.fecha_hora_gps,
			c.insert_timestamp AS fecha_hora_insert
		FROM
			%s c
		JOIN
			%s g ON ST_Intersects(c.coordenadas, g.geometria)
		WHERE
			c.velocidad = 0
			AND g.tipo_geocerca = 'Zona roja'
			AND DATE(c.fecha_hora_gps) = $1;`,
		database.Cfg.TableNames.ConcentradorGPS,
		database.Cfg.TableNames.Geocercas,
	)

	log.Printf("Buscando intersecciones de camiones detenidos en zonas rojas para la fecha: %s", fecha)
	rows, err := database.DB.Query(query, fecha) // Usamos database.DB
	if err != nil {
		return nil, fmt.Errorf("error al obtener intersecciones de zona roja: %w", err)
	}
	defer rows.Close()

	var eventosDetectados []models.Evento
	for rows.Next() {
		var e models.Evento
		err := rows.Scan(&e.Patente, &e.IDGeocerca, &e.CoordenadasWKT, &e.Velocidad, &e.Orientacion, &e.FechaHoraGps, &e.FechaHoraInsert)
		if err != nil {
			return nil, fmt.Errorf("error al escanear evento detectado: %w", err)
		}
		eventosDetectados = append(eventosDetectados, e)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error después de iterar filas de intersecciones: %w", err)
	}

	return eventosDetectados, nil
}

func InsertarEventoZonaRoja(e models.Evento) error {
	if database.Cfg == nil || database.Cfg.TableNames.PreEventosZonaRoja == "" {
		return fmt.Errorf("nombre de tabla EventosZonaRoja no configurado")
	}
	if database.DB == nil {
		return fmt.Errorf("conexión a la base de datos no inicializada")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (patente, id_geocerca, coordenadas, velocidad, orientacion, fecha_hora_gps, fecha_hora_insert)
		VALUES ($1, $2, $3, $4, $5, $6, $7);`,
		database.Cfg.TableNames.PreEventosZonaRoja,
	)
	_, err := database.DB.Exec(query, // Usamos database.DB
		e.Patente,
		e.IDGeocerca,
		e.CoordenadasWKT,
		e.Velocidad,
		e.Orientacion,
		e.FechaHoraGps,
		e.FechaHoraInsert,
	)
	if err != nil {
		return fmt.Errorf("error al insertar evento de zona roja para patente %s: %w", e.Patente, err)
	}
	return nil
}
