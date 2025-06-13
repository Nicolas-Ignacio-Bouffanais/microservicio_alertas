package det_no_autorizada

import (
	"fmt"
	"time"

	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/database"
	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/models"
)

func GetDetencionesIlegales() ([]models.Evento, error) {
	if database.Cfg == nil {
		return nil, fmt.Errorf("la configuración no ha sido inicializada")
	}

	query := fmt.Sprintf(`
		SELECT
			c.id AS id_concentrador,
			c.patente,
			ST_AsText(c.coordenadas) AS coordenadas_wkt,
			c.velocidad,
			c.orientacion,
			c.fecha_hora_gps,
			c.fecha_hora_registro,
			c.insert_timestamp AS fecha_hora_insert
		FROM
			%s c
		LEFT JOIN
			%s tm ON c.id = tm.id_concentrador
		WHERE
			c.velocidad = 0
			AND (tm.id_concentrador IS NULL OR tm.det_no_autorizada = '%s')
			AND NOT EXISTS (
				SELECT 1
				FROM %s g
				WHERE
					g.tipo_geocerca IN ('Zona de Origen', 'Zona de Destino', 'Zona de Peaje', 'Zona de Pesaje')
					AND ST_Intersects(c.coordenadas, g.geometria)
			)
		LIMIT 10;
	`,
		database.Cfg.TableNames.ConcentradorGPS,
		database.Cfg.TableNames.TablaMarcado,
		models.NoProcesado,
		database.Cfg.TableNames.Geocercas)

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error al obtener detenciones ilegales: %w", err)
	}
	defer rows.Close()

	var eventosDetectados []models.Evento
	for rows.Next() {
		var e models.Evento
		e.FechaHoraCalc = time.Now()
		e.IDGeocerca = nil
		err := rows.Scan(
			&e.IDConcentrador,
			&e.Patente,
			&e.CoordenadasWKT,
			&e.Velocidad,
			&e.Orientacion,
			&e.FechaHoraGps,
			&e.FechaHoraRegistro,
			&e.FechaHoraInsert,
		)
		if err != nil {
			return nil, fmt.Errorf("error al escanear evento de detención ilegal: %w", err)
		}
		eventosDetectados = append(eventosDetectados, e)
	}
	return eventosDetectados, rows.Err()
}

func ActualizarEstadoDetNoAut(idConcentrador int64, estado models.EstadoProcesamiento) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id_concentrador, det_no_autorizada)
		VALUES ($1, $2)
		ON CONFLICT (id_concentrador) DO UPDATE
		SET det_no_autorizada = EXCLUDED.det_no_autorizada,
			fecha_marcado = NOW();
		`,
		database.Cfg.TableNames.TablaMarcado,
	)
	_, err := database.DB.Exec(query, idConcentrador, estado)
	return err
}

// CORREGIDO: Se estandarizó el nombre de la función.
func InsertarPreEventoDetNoAut(e models.Evento) error {
	query := fmt.Sprintf(`
        INSERT INTO %s (patente, id_geocerca, coordenadas, velocidad, orientacion, fecha_hora_gps, fecha_hora_insert, fecha_hora_calc, fecha_hora_registro)
        VALUES ($1, $2, ST_GeomFromText($3, 4326), $4, $5, $6, $7, NOW(), $8);`,
		database.Cfg.TableNames.PreEventosDetNoAutorizada)

	_, err := database.DB.Exec(query, e.Patente, e.IDGeocerca, e.CoordenadasWKT, e.Velocidad, e.Orientacion, e.FechaHoraGps, e.FechaHoraInsert, e.FechaHoraRegistro)
	return err
}
