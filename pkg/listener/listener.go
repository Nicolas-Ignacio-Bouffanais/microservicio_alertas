package listener

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/database"
)

func Iniciar(canal string, callback func(payload string)) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChan
		log.Printf("Señal de interrupción recibida, deteniendo listener del canal '%s'...", canal)
		cancel()
	}()

	log.Printf("Iniciando listener para el canal de PostgreSQL: '%s'", canal)

	for {
		// Comprobamos si el contexto ha sido cancelado antes de intentar conectar.
		select {
		case <-ctx.Done():
			log.Printf("Contexto cancelado. El listener del canal '%s' se ha detenido.", canal)
			return
		default:
		}

		conn, err := database.Pool.Acquire(ctx)
		if err != nil {
			log.Printf("Error al adquirir conexión para el canal '%s', reintentando en 5 segundos: %v", canal, err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Ejecutamos LISTEN en la conexión recién adquirida.
		_, err = conn.Exec(ctx, "LISTEN "+canal)
		if err != nil {
			log.Printf("Error al ejecutar LISTEN en el canal '%s', reintentando en 5 segundos: %v", canal, err)
			conn.Release() // Liberamos la conexión fallida.
			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("Conexión establecida. Escuchando notificaciones en el canal: '%s'", canal)

		// Bucle anidado para esperar notificaciones.
		for {
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				log.Printf("Error esperando notificación en '%s': %v. Se intentará reconectar.", canal, err)
				break // Sale del bucle de espera, irá al conn.Release() y luego el bucle principal reintentará.
			}

			log.Printf("¡Notificación recibida en el canal '%s'! Payload: '%s'", notification.Channel, notification.Payload)
			// Lanzamos el procesamiento en una goroutine para no bloquear la recepción de la siguiente notificación.
			go callback(notification.Payload)
		}

		conn.Release()
	}
}
