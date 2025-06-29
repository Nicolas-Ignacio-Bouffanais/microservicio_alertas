package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type TableNames struct {
	ConcentradorGPS           string
	TablaMarcado              string
	Geocercas                 string
	Rutas                     string
	Vehiculos                 string
	PreEventosZonaRoja        string
	PreEventosDetNoAutorizada string
	PreEventosSobreestadia    string
	PreEventosExVelocidad     string
}

type AppConfig struct {
	Database   DBConfig
	TableNames TableNames
}

func LoadConfig(envPath ...string) (*AppConfig, error) {
	var err error
	if len(envPath) > 0 && envPath[0] != "" {
		err = godotenv.Load(envPath[0])
	} else {
		err = godotenv.Load()
	}

	if err != nil {
		log.Printf("Advertencia: No se pudo cargar el archivo .env: %v. Se usarán variables de entorno directas o valores por defecto.", err)
	}

	dbPortStr := getEnv("DB_PORT", "5432")
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		return nil, fmt.Errorf("DB_PORT inválido '%s': %w", dbPortStr, err)
	}

	cfg := &AppConfig{
		Database: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "mydatabase"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		TableNames: TableNames{
			ConcentradorGPS:           getEnv("CONCENTRADOR_GPS", "tablagps"),
			TablaMarcado:              getEnv("TablaMarcado", "TablaMarcado"),
			Geocercas:                 getEnv("GEOCERCAS", "public.geocercas"),
			Rutas:                     getEnv("RUTAS", "public.rutas"),
			Vehiculos:                 getEnv("VEHICULOS", "vehiculos"),
			PreEventosZonaRoja:        getEnv("ZONA_ROJA", "public.ZonaRoja"),
			PreEventosDetNoAutorizada: getEnv("DET_NO_AUTORIZADA", "public.DetNoAutorizada"),
			PreEventosSobreestadia:    getEnv("SOBREESTADIA", "TablaSobreestadia"),
			PreEventosExVelocidad:     getEnv("EX_VELOCIDAD", "ex_velocidad"),
		},
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
