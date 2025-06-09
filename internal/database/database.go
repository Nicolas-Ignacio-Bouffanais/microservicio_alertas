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

func GetZonasRojas() ([]models.Geocerca, error) {
	if Cfg == nil || Cfg.TableNames.Geocercas == "" {
		return nil, fmt.Errorf("la configuración de nombres de tabla no está cargada o el nombre de la tabla de geocercas está vacío")
	}
	if DB == nil {
		return nil, fmt.Errorf("la conexión a la base de datos no ha sido inicializada")
	}

	query := fmt.Sprintf(`
	SELECT 
		id, 
		descripcion,
		tipo_geocerca,
		ST_AsText(geometria) AS geometria_wkt,
		bounding_xmin, 
		bounding_xmax, 
		bounding_ymin, 
		bounding_ymax
	FROM %s WHERE tipo_geocerca = 'Zona roja'`, Cfg.TableNames.Geocercas)

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error al obtener zonas rojas: %w", err)
	}
	defer rows.Close()

	var geocercas []models.Geocerca
	for rows.Next() {
		var g models.Geocerca

		err := rows.Scan(
			&g.IDGeocerca,
			&g.Descripcion,
			&g.TipoGeocerca,
			&g.Geometria,
			&g.BoundingXMin,
			&g.BoundingXMax,
			&g.BoundingYMin,
			&g.BoundingYMax,
		)
		if err != nil {
			log.Printf("Error al escanear fila de zona roja: %v", err)
			return nil, fmt.Errorf("error al escanear fila de zona roja: %w", err)
		}
		geocercas = append(geocercas, g)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error después de iterar las filas de zonas rojas: %w", err)
	}
	return geocercas, nil
}

func GetCamionesDetenidos() ([]models.Camion, error) {
	if Cfg == nil || Cfg.TableNames.ConcentradorGPS == "" || Cfg.TableNames.CategorizacionGPS == "" {
		return nil, fmt.Errorf("configuración de nombres de tabla no cargada completamente")
	}
	if DB == nil {
		return nil, fmt.Errorf("conexión a la base de datos no inicializada")
	}

	query := fmt.Sprintf(`
		SELECT c.id, c.patente, c.velocidad, c.orientacion, c.fecha_hora_gps, ST_AsText(ST_SetSRID(c.coordenadas, 4326)) AS coordenadas_wkt
		FROM %s c
		WHERE c.velocidad = 0 AND DATE(fecha_hora_gps) = '2025-05-12';`,
		Cfg.TableNames.ConcentradorGPS,
	)
	log.Printf("Obteniendo los camiones detenidos en zona roja: %s", query)
	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error al obtener camiones para zona roja: %w", err)
	}
	defer rows.Close()

	var camiones []models.Camion
	for rows.Next() {
		var c models.Camion
		err := rows.Scan(&c.Id, &c.Patente, &c.Velocidad, &c.FechaHoraGPS, &c.Orientacion, &c.CoordenadasWKT)
		if err != nil {
			return nil, fmt.Errorf("error al escanear camión no marcado para zona roja: %w", err)
		}
		camiones = append(camiones, c)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error después de iterar filas de camiones: %w", err)
	}
	return camiones, nil
}

func CamionIntersectaAlgunaGeocerca(camionCoordenadasWKT string, geocercas []models.Geocerca) (bool, string, error) {
	if DB == nil {
		return false, "", fmt.Errorf("conexión a la base de datos no inicializada")
	}

	for _, geo := range geocercas {
		if geo.Geometria == nil || *geo.Geometria == "" {
			continue
		}

		var intersecta bool
		query := `SELECT ST_Intersects(ST_SetSRID(ST_GeomFromText($1), 4326), ST_SetSRID(ST_GeomFromText($2), 4326));`
		err := DB.QueryRow(query, camionCoordenadasWKT, *geo.Geometria).Scan(&intersecta)

		if err != nil {
			log.Printf("Error en ST_Intersects para geocerca ID %d y WKT '%s': %v", geo.IDGeocerca, camionCoordenadasWKT, err)
			continue
		}
		if intersecta {
			return true, fmt.Sprintf("%d", geo.IDGeocerca), nil
		}
	}
	return false, "", nil
}

func InsertarEventoZonaRoja(e models.Evento) error {
	if Cfg == nil || Cfg.TableNames.PreEventosZonaRoja == "" {
		return fmt.Errorf("nombre de tabla EventosZonaRoja no configurado")
	}
	if DB == nil {
		return fmt.Errorf("conexión a la base de datos no inicializada")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (patente, id_geocerca, coordenadas_wkt, fecha_hora, velocidad, orientacion) 
		VALUES ($1, $2, $3, $4, $5, $6);`,
		Cfg.TableNames.PreEventosZonaRoja,
	)
	_, err := DB.Exec(query,
		e.Patente,
		e.IDGeocerca,
		e.CoordenadasWKT,
		e.Velocidad,
		e.Orientacion,
	)
	if err != nil {
		return fmt.Errorf("error al insertar evento de zona roja para patente %s: %w", e.Patente, err)
	}
	return nil
}
