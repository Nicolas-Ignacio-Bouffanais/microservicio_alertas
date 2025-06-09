package runner

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func RunService(serviceName string, serviceFunc func(), frequency time.Duration, executeOnStart bool) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChan
		log.Printf("[%s] Señal de interrupción recibida, deteniendo...", serviceName)
		cancel()
	}()

	log.Printf("[%s] Iniciado. Frecuencia: %v", serviceName, frequency)
	ticker := time.NewTicker(frequency)
	defer ticker.Stop()

	if executeOnStart {
		log.Printf("[%s] Ejecución inicial...", serviceName)
		go serviceFunc()
	}

	for {
		select {
		case <-ticker.C:
			log.Printf("[%s] Ticker activado.", serviceName)
			go serviceFunc()
		case <-ctx.Done():
			log.Printf("[%s] Contexto cancelado. Finalizando.", serviceName)
			time.Sleep(2 * time.Second)
			return
		}
	}
}
