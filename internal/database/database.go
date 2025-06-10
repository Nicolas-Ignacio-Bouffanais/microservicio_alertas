package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/config"
	models "github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/models"

	_ "github.com/jackc/pgx/v5/stdlib" //driver de postgresql
)

var (
	DB  *sql.DB
	Cfg *config.AppConfig
)

func ConnectDB(appConfig *config.AppConfig) error {
	Cfg = appConfig
	dbSettings := appConfig.Database

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbSettings.Host,
		dbSettings.Port,
		dbSettings.User,
		dbSettings.Password,
		dbSettings.DBName,
		dbSettings.SSLMode,
	)

	log.Printf("Intentando conectar a PostgreSQL con DSN: host=%s port=%d user=%s dbname=%s sslmode=%s",
		dbSettings.Host, dbSettings.Port, dbSettings.User, dbSettings.DBName, dbSettings.SSLMode)

	var err error

	DB, err = sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("error al abrir la conexión a la base de datos PostgreSQL: %w", err)
	}

	err = DB.Ping()
	if err != nil {
		DB.Close()
		return fmt.Errorf("error al hacer ping a la base de datos PostgreSQL: %w", err)
	}

	log.Println("¡Conexión exitosa a la base de datos PostgreSQL/PostGIS!")
	return nil
}

func GetInterseccionesZonaRoja(fecha string) ([]models.Evento, error) {
	if Cfg == nil || Cfg.TableNames.ConcentradorGPS == "" || Cfg.TableNames.Geocercas == "" {
		return nil, fmt.Errorf("configuración de nombres de tabla no cargada completamente")
	}
	if DB == nil {
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
		Cfg.TableNames.ConcentradorGPS,
		Cfg.TableNames.Geocercas,
	)

	log.Printf("Buscando intersecciones de camiones detenidos en zonas rojas para la fecha: %s", fecha)
	rows, err := DB.Query(query, fecha)
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
	if Cfg == nil || Cfg.TableNames.PreEventosZonaRoja == "" {
		return fmt.Errorf("nombre de tabla EventosZonaRoja no configurado")
	}
	if DB == nil {
		return fmt.Errorf("conexión a la base de datos no inicializada")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (patente, id_geocerca, coordenadas, velocidad, orientacion, fecha_hora_gps, fecha_hora_insert) 
		VALUES ($1, $2, $3, $4, $5, $6, $7);`,
		Cfg.TableNames.PreEventosZonaRoja,
	)
	_, err := DB.Exec(query,
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
