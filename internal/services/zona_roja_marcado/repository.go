package zona_roja_marcado

import (
	"fmt"

	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/database"
	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/models"
)

// GetInterseccionesNoMarcadas busca camiones detenidos en zonas rojas que aún no han sido procesados.
func GetInterseccionesNoMarcadas() ([]models.Evento, error) {
	if database.Cfg == nil {
		return nil, fmt.Errorf("la configuración no ha sido inicializada")
	}
	query := fmt.Sprintf(`
		SELECT
			c.id AS id_concentrador, -- Se necesita el ID para marcarlo después
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
			%s cat ON c.id = cat.id_concentrador
		WHERE
			c.velocidad = 0
			AND g.tipo_geocerca = 'Zona roja'
			-- La condición clave: procesar solo si no está marcado o la marca es explícitamente falsa.
			AND (cat.id_concentrador IS NULL OR cat.zona_roja = false);
		`,
		database.Cfg.TableNames.ConcentradorGPS,
		database.Cfg.TableNames.Geocercas,
		database.Cfg.TableNames.TablaMarcado,
	)

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error al obtener intersecciones no marcadas: %w", err)
	}
	defer rows.Close()

	var eventosDetectados []models.Evento
	for rows.Next() {
		var e models.Evento
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

func MarcarComoProcesado(idConcentrador int64) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id_concentrador, zona_roja, fecha_marcado)
		VALUES ($1, true, NOW())
		ON CONFLICT (id_concentrador) DO UPDATE
		SET zona_roja = EXCLUDED.zona_roja,
			fecha_marcado = EXCLUDED.fecha_marcado;
		`,
		database.Cfg.TableNames.TablaMarcado,
	)

	_, err := database.DB.Exec(query, idConcentrador)
	if err != nil {
		return fmt.Errorf("error al marcar como procesado para id_concentrador %d: %w", idConcentrador, err)
	}
	return nil
}
