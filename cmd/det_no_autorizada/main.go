package main

import (
	"log"
	"time"

	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/config"
	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/database"
	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/services/det_no_autorizada"
	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/pkg/runner"
)

func main() {
	log.Println("Iniciando servicio de alerta: Zona Roja...")
	appConfig, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error al cargar config: %v", err)
	}

	if err := database.ConnectDB(appConfig); err != nil {
		log.Fatalf("Error al conectar a DB: %v", err)
	}
	defer database.DB.Close()

	frecuencia := 30 * time.Second
	runner.RunService("DetNoAutorizada", det_no_autorizada.ProcesarEventos, frecuencia, true)
}
