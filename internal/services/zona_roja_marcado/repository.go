package zona_roja_marcado

import (
	"fmt"
	"time"

	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/database"
	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/models"
)

// GetInterseccionesNoMarcadas busca camiones en zonas rojas cuyo estado en la columna 'zona_roja' es 'no_procesado'.
func GetInterseccionesNoMarcadas() ([]models.Evento, error) {
	if database.Cfg == nil {
		return nil, fmt.Errorf("la configuración no ha sido inicializada")
	}
	// La consulta ahora busca explícitamente en la columna 'zona_roja'
	query := fmt.Sprintf(`
		SELECT
			c.id AS id_concentrador,
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
		LEFT JOIN
			%s tm ON c.id = tm.id_concentrador
		WHERE
			c.velocidad = 0
			AND g.tipo_geocerca = 'Zona roja'
			AND (tm.id_concentrador IS NULL OR tm.zona_roja = '%s');
		`,
		database.Cfg.TableNames.ConcentradorGPS,
		database.Cfg.TableNames.Geocercas,
		database.Cfg.TableNames.TablaMarcado,
		models.NoProcesado,
	)

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error al obtener intersecciones no marcadas: %w", err)
	}
	defer rows.Close()

	var eventosDetectados []models.Evento
	for rows.Next() {
		var e models.Evento
		e.FechaHoraCalc = time.Now()
		err := rows.Scan(
			&e.IDConcentrador,
			&e.Patente,
			&e.IDGeocerca,
			&e.CoordenadasWKT,
			&e.Velocidad,
			&e.Orientacion,
			&e.FechaHoraGps,
			&e.FechaHoraInsert,
		)
		if err != nil {
			return nil, fmt.Errorf("error al escanear evento no marcado: %w", err)
		}
		eventosDetectados = append(eventosDetectados, e)
	}
	return eventosDetectados, rows.Err()
}

// ActualizarEstadoZonaRoja actualiza el estado EN LA COLUMNA 'zona_roja' de la tabla de marcado.
func ActualizarEstadoZonaRoja(idConcentrador int64, estado models.EstadoProcesamiento) error {
	// Esta consulta inserta o actualiza el estado en la columna específica 'zona_roja'.
	// NO hace referencia a una columna 'estado'.
	query := fmt.Sprintf(`
		INSERT INTO %s (id_concentrador, zona_roja)
		VALUES ($1, $2)
		ON CONFLICT (id_concentrador) DO UPDATE
		SET zona_roja = EXCLUDED.zona_roja,
			fecha_marcado = NOW();
		`,
		database.Cfg.TableNames.TablaMarcado,
	)

	_, err := database.DB.Exec(query, idConcentrador, estado)
	if err != nil {
		// Este es el mensaje de error que estabas viendo, pero ahora la consulta es correcta.
		return fmt.Errorf("error al marcar estado de zona roja para id_concentrador %d: %w", idConcentrador, err)
	}
	return nil
}

func InsertarPreEvento(e models.Evento) error {
	query := fmt.Sprintf(`
        INSERT INTO %s (patente, id_geocerca, coordenadas, velocidad, orientacion, fecha_hora_gps, fecha_hora_insert, fecha_hora_calc)
        VALUES ($1, $2, ST_GeomFromText($3, 4326), $4, $5, $6, $7, NOW());`,
		database.Cfg.TableNames.PreEventosZonaRoja)

	_, err := database.DB.Exec(query, e.Patente, e.IDGeocerca, e.CoordenadasWKT, e.Velocidad, e.Orientacion, e.FechaHoraGps, e.FechaHoraInsert)
	if err != nil {
		return fmt.Errorf("error al insertar pre-evento de zona roja: %w", err)
	}
	return nil
}
