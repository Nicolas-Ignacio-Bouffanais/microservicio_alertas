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

func GetCamionesDetenidos() ([]models.Camion, error) {
	if Cfg == nil || Cfg.TableNames.ConcentradorGPS == "" {
		return nil, fmt.Errorf("la configuración de nombres de tabla no está cargada o el nombre de la tabla de GPS está vacío")
	}
	if DB == nil {
		return nil, fmt.Errorf("la conexión a la base de datos no ha sido inicializada")
	}

	query := fmt.Sprintf(`
	SELECT patente, orientacion, velocidad, fecha_hora_gps 
	FROM %s
	WHERE velocidad = 0 AND DATE(fecha_hora_gps) = '2025-05-12'`, Cfg.TableNames.ConcentradorGPS)

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar la consulta de camiones detenidos: %w", err)
	}
	defer rows.Close()

	var camiones []models.Camion
	for rows.Next() {
		var c models.Camion
		err := rows.Scan(
			&c.Patente,
			&c.Orientacion,
			&c.Velocidad,
			&c.FechaHoraGPS,
		)
		if err != nil {
			log.Printf("Error al escanear fila de camión: %v", err)
			return nil, fmt.Errorf("error al escanear fila de camión: %w", err)
		}
		camiones = append(camiones, c)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error después de iterar las filas de camiones: %w", err)
	}

	return camiones, nil
}

func GetGeocercas() ([]models.Geocerca, error) {
	if Cfg == nil || Cfg.TableNames.Geocercas == "" {
		return nil, fmt.Errorf("la configuración de nombres de tabla no está cargada o el nombre de la tabla de geocercas está vacío")
	}
	if DB == nil {
		return nil, fmt.Errorf("la conexión a la base de datos no ha sido inicializada")
	}

	query := fmt.Sprintf(`
	SELECT 
		id_geocerca, 
		descripcion, 
		tipo_geocerca, 
		ST_AsText(geometria) AS geometria_wkt,
		bounding_xmin, 
		bounding_xmax, 
		bounding_ymin, 
		bounding_ymax,
		vel_umbral,
		estadia_max 
	FROM %s`, Cfg.TableNames.Geocercas)

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error al obtener geocercas: %w", err)
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
			&g.VelUmbral,
			&g.TiempoMax,
		)
		if err != nil {
			log.Printf("Error al escanear fila de geocerca: %v", err)
			return nil, fmt.Errorf("error al escanear fila de geocerca: %w", err)
		}
		geocercas = append(geocercas, g)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error después de iterar las filas de geocercas: %w", err)
	}
	return geocercas, nil
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
		id_geocerca, 
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

func GetZonasOrigen() ([]models.Geocerca, error) {
	if Cfg == nil || Cfg.TableNames.Geocercas == "" {
		return nil, fmt.Errorf("la configuración de nombres de tabla no está cargada o el nombre de la tabla de geocercas está vacío")
	}
	if DB == nil {
		return nil, fmt.Errorf("la conexión a la base de datos no ha sido inicializada")
	}

	query := fmt.Sprintf(`
	SELECT 
		id_geocerca, 
		descripcion,
		tipo_geocerca,
		ST_AsText(geometria) AS geometria_wkt,
		bounding_xmin, 
		bounding_xmax, 
		bounding_ymin, 
		bounding_ymax
	FROM %s WHERE tipo_geocerca = 'Zona de origen y destino'`, Cfg.TableNames.Geocercas)

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
