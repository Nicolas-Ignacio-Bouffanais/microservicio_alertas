package ex_velocidad

import (
	"context"
	"fmt"
	"time"

	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/database"
	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/models"
)

func GetExVelocidad(batchID string) ([]models.PreEventoExVelocidad, error) {
	if database.Cfg == nil {
		return nil, fmt.Errorf("la configuración no ha sido inicializada")
	}
	query := fmt.Sprintf(`
		WITH limites_base AS (
			SELECT
				c.id AS id_concentrador,
				c.patente,
				c.fecha_hora_gps,
				c.coordenadas,
				c.velocidad,
				c.orientacion,
				c.fecha_hora_registro,
				c.insert_timestamp AS fecha_hora_insert,
				CASE 
					WHEN v.subgrupo_vehiculo = 'CAMION' THEN 90.0
					WHEN v.subgrupo_vehiculo = 'CAMIONETA' THEN 120.0
					ELSE 90.0
				END AS limite_base
			FROM
				%s c
			JOIN
				%s v ON c.patente = v.patente
			LEFT JOIN
				%s tm ON c.id = tm.id_concentrador
			WHERE
				c.batch_id = $1
				AND c.velocidad > 0
				AND (tm.id_concentrador IS NULL OR tm.ex_velocidad = 'no_procesado')
		),
		limites_geocerca AS (
			SELECT
				lb.*,
				g.id AS id_geocerca_control,
				g.vel_umbral AS limite_geo
			FROM
				limites_base lb
			LEFT JOIN
				%s g ON ST_Intersects(lb.coordenadas, g.geometria)
						 AND g.tipo_geocerca = 'Control de Velocidad'
		),
		eventos_potenciales AS (
			SELECT
				lg.*,
				COALESCE(lg.limite_geo, lg.limite_base) AS limite_final,
				CASE 
					WHEN COALESCE(lg.limite_geo, lg.limite_base) > 0 THEN
						(lg.velocidad / COALESCE(lg.limite_geo, lg.limite_base) - 1) * 100
					ELSE 0
				END AS porcentaje_exceso
			FROM
				limites_geocerca lg
		)
		SELECT
			ep.id_concentrador,
			ep.patente,
			ep.fecha_hora_gps,
			ep.fecha_hora_registro,
			ep.fecha_hora_insert,
			ST_AsText(ep.coordenadas),
			ep.orientacion,
			ep.velocidad,
			ep.limite_final,
			(ep.velocidad - ep.limite_final),
			CASE 
				WHEN ep.porcentaje_exceso > 6 THEN 'GRAVE'
				ELSE 'LEVE'
			END AS criticidad,
			ep.id_geocerca_control
		FROM
			eventos_potenciales ep
		WHERE
			ep.velocidad > ep.limite_final;
	`,
		database.Cfg.TableNames.ConcentradorGPS,
		database.Cfg.TableNames.Vehiculos,
		database.Cfg.TableNames.TablaMarcado,
		database.Cfg.TableNames.Geocercas,
	)

	rows, err := database.Pool.Query(context.Background(), query, batchID)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar la consulta de exceso de velocidad: %w", err)
	}
	defer rows.Close()

	var eventosDetectados []models.PreEventoExVelocidad
	for rows.Next() {
		var e models.PreEventoExVelocidad
		e.FechaHoraCalc = time.Now()
		e.IDGeocerca = nil
		err := rows.Scan(
			&e.IDConcentrador,
			&e.Patente,
			&e.CoordenadasWKT,
			&e.VelocidadRegistrada,
			&e.Orientacion,
			&e.IDGeocerca,
			&e.FechaHoraGps,
			&e.FechaHoraRegistro,
			&e.FechaHoraInsert,
			&e.ExcesoRegistrado,
			&e.Criticidad,
		)
		if err != nil {
			return nil, fmt.Errorf("error al escanear evento de detención ilegal: %w", err)
		}
		eventosDetectados = append(eventosDetectados, e)
	}
	return eventosDetectados, rows.Err()
}

func ActualizarEstadoExVelocidad(ids []int64, estado models.EstadoProcesamiento) error {
	if len(ids) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id_concentrador, ex_velocidad)
		SELECT id, $1 FROM unnest($2::bigint[]) AS id
		ON CONFLICT (id_concentrador) DO UPDATE
		SET 
			ex_velocidad = EXCLUDED.ex_velocidad;
	`, database.Cfg.TableNames.TablaMarcado)

	_, err := database.Pool.Exec(context.Background(), query, estado, ids)
	if err != nil {
		return fmt.Errorf("error al actualizar estado de exceso de velocidad para %d IDs: %w", len(ids), err)
	}
	return nil
}

func InsertarPreEventoExVelocidad(e models.PreEventoExVelocidad) error {
	query := fmt.Sprintf(`
        INSERT INTO %s (
            id_concentrador, patente, fecha_hora_gps, id_geocerca, velocidad_registrada,
            velocidad_limite, exceso_registrado, criticidad, fecha_hora_registro, fecha_hora_insert
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `, database.Cfg.TableNames.PreEventosExVelocidad)

	_, err := database.Pool.Exec(context.Background(),
		query,
		e.IDConcentrador, e.Patente, e.FechaHoraGps, e.IDGeocerca, e.VelocidadRegistrada,
		e.VelocidadLimite, e.ExcesoRegistrado, e.Criticidad, e.FechaHoraRegistro, e.FechaHoraInsert,
	)

	if err != nil {
		return fmt.Errorf("error al insertar pre-evento de exceso de velocidad para id_concentrador %d: %w", e.IDConcentrador, err)
	}
	return nil
}
