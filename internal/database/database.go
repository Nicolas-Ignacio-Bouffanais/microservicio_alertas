package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/config"

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
