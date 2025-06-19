package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/config"
	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/database"
	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/services/ex_velocidad"
	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/pkg/listener"
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
	defer database.Pool.Close()

	go listener.Iniciar("gps_batch_listo", ex_velocidad.ProcesarBatch)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	log.Println("Señal de apagado recibida. Cerrando conexiones...")
}
