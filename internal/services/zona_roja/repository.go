package zona_roja

import (
	"context"
	"fmt"
	"time"

	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/database"
	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/models"
)

func GetZonaRoja(batchID string) ([]models.PreEventoZonaRoja, error) {
	if database.Cfg == nil {
		return nil, fmt.Errorf("la configuración no ha sido inicializada")
	}
	query := fmt.Sprintf(`
		SELECT
			c.id AS id_concentrador,
			c.patente,
			g.id::text AS id_geocerca,
			ST_AsText(c.coordenadas) AS coordenadas_wkt,
			c.velocidad,
			c.orientacion,
			c.fecha_hora_gps,
			c.fecha_hora_registro,
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
			AND (tm.id_concentrador IS NULL OR tm.zona_roja = '%s')
		LIMIT 10;
		`,
		database.Cfg.TableNames.ConcentradorGPS,
		database.Cfg.TableNames.Geocercas,
		database.Cfg.TableNames.TablaMarcado,
		models.NoProcesado,
	)

	rows, err := database.Pool.Query(context.Background(), query, batchID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener intersecciones no marcadas: %w", err)
	}
	defer rows.Close()

	var eventosDetectados []models.PreEventoZonaRoja
	for rows.Next() {
		var e models.PreEventoZonaRoja
		e.FechaHoraCalc = time.Now()
		err := rows.Scan(
			&e.IDConcentrador,
			&e.Patente,
			&e.IDGeocerca,
			&e.CoordenadasWKT,
			&e.Orientacion,
			&e.FechaHoraGps,
			&e.FechaHoraRegistro,
			&e.FechaHoraInsert,
		)
		if err != nil {
			return nil, fmt.Errorf("error al escanear evento no marcado: %w", err)
		}
		eventosDetectados = append(eventosDetectados, e)
	}
	return eventosDetectados, rows.Err()
}

func ActualizarEstadoZonaRoja(idConcentrador []int64, estado models.EstadoProcesamiento) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id_concentrador, zona_roja)
		VALUES ($1, $2)
		ON CONFLICT (id_concentrador) DO UPDATE
		SET zona_roja = EXCLUDED.zona_roja,
			fecha_marcado = NOW();
		`,
		database.Cfg.TableNames.TablaMarcado,
	)

	_, err := database.Pool.Exec(context.Background(), query, idConcentrador, estado)
	if err != nil {
		return fmt.Errorf("error al marcar estado de zona roja para id_concentrador %d: %w", idConcentrador, err)
	}
	return nil
}

func InsertarPreEventoZonaRoja(e models.PreEventoZonaRoja) error {
	query := fmt.Sprintf(`
        INSERT INTO %s (patente, id_geocerca, coordenadas, orientacion, fecha_hora_gps, fecha_hora_registro, fecha_hora_insert, fecha_hora_calc)
        VALUES ($1, $2, ST_GeomFromText($3, 4326), $4, $5, $6, $7, $8, NOW());`,
		database.Cfg.TableNames.PreEventosZonaRoja)

	_, err := database.Pool.Exec(context.Background(), query, e.Patente, e.IDGeocerca, e.CoordenadasWKT, e.Orientacion, e.FechaHoraGps, e.FechaHoraRegistro, e.FechaHoraInsert)
	if err != nil {
		return fmt.Errorf("error al insertar pre-evento de zona roja: %w", err)
	}
	return nil
}
