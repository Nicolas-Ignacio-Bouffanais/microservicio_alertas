package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/config"
	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/database"
	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/services"
)

func runService(ctx context.Context,
	wg *sync.WaitGroup,
	serviceName string,
	serviceFunc func(),
	frequency time.Duration,
	executeOnStart bool) {
	defer wg.Done()
	log.Printf("Iniciando servicio: %s (frecuencia: %v)", serviceName, frequency)
	ticker := time.NewTicker(frequency)
	defer ticker.Stop()

	if executeOnStart {
		log.Printf("Ejecucion inicial de %s ...", serviceName)
	}
	for {
		select {
		case <-ticker.C:
			log.Printf("Ticker activado para %s", serviceName)
			go serviceFunc()

		case <-ctx.Done():
			log.Printf("Contexto cancelado. Deteniendo servicio: %s", serviceName)
			return
		}
	}
}

func main() {
	log.Println("Iniciando servicio de generaracion de eventos...")
	appConfig, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error critico al cargar la configuracion: %v", err)
	}
	log.Println("Configuracion cargada")

	if err := database.ConnectDB(appConfig); err != nil {
		log.Fatalf("Error crítico al conectar con la(s) base(s) de datos: %v", err)
	}
	log.Println("Conexión(es) a base de datos establecida(s).")
	if database.DB != nil {
		defer func() {
			log.Println("Cerrando conexión a PostgreSQL...")
			database.DB.Close()
		}()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	var wgServices sync.WaitGroup

	// wgServices.Add(1)
	// go runService(ctx, &wgServices, "Sobreestadia", services.ProcesarSobreestadia, 1*time.Minute, true)

	// wgServices.Add(1)
	// go runService(ctx, &wgServices, "ZonaRoja", services.ProcesarZonaRoja, 1*time.Minute, true)

	wgServices.Add(1)
	go runService(ctx, &wgServices, "ZonaOrigen", services.ProcesarZonaOrigen, 1*time.Minute, true)

	log.Println("Todos los monitores de servicios han sido iniciados y están corriendo.")

	<-signalChan // Bloquea hasta que se reciba una señal (Ctrl+C, SIGTERM)
	log.Println("Señal de interrupción recibida, enviando cancelación a los servicios...")
	cancel() // Notifica a todos los servicios (a través del contexto) que deben detenerse

	log.Println("Esperando que todos los servicios finalicen ordenadamente...")
	wgServices.Wait() // Espera a que todas las goroutines de runService (y por ende los tickers) terminen

	log.Println("Todos los servicios han finalizado. La aplicación se cerrará.")
}
