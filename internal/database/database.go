package database

import (
	"context"

	"fmt"
	"log"

	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" //driver de postgresql
)

var (
	Pool *pgxpool.Pool
	Cfg  *config.AppConfig
)

func ConnectDB(appConfig *config.AppConfig) error {
	Cfg = appConfig
	dbSettings := appConfig.Database

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s pool_max_conns=10",
		dbSettings.Host, dbSettings.Port, dbSettings.User, dbSettings.Password,
		dbSettings.DBName, dbSettings.SSLMode)

	log.Printf("Intentando conectar a PostgreSQL con DSN: host=%s port=%d user=%s dbname=%s sslmode=%s",
		dbSettings.Host, dbSettings.Port, dbSettings.User, dbSettings.DBName, dbSettings.SSLMode)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("no se pudo crear el pool de conexiones: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return fmt.Errorf("no se pudo hacer ping a la base de datos: %w", err)
	}

	Pool = pool
	log.Println("¡Conexión exitosa a PostgreSQL/PostGIS usando pgxpool!")
	return nil
}
