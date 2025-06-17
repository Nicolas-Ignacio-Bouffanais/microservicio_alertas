package ex_velocidad

import (
	"log"
	"sync"

	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/models"
)

const tipoServicioActual = "EX_VELOCIDAD"

func ProcesarEventos(batchID string) {
	log.Printf("[%s] Notificación recibida. Iniciando procesamiento para batch_id: %s", tipoServicioActual, batchID)

	eventosParaProcesar, err := GetExVelocidad(batchID)
	if err != nil {
		log.Printf("[%s] Error al obtener eventos para batch_id %s: %v", tipoServicioActual, batchID, err)
		return
	}

	if len(eventosParaProcesar) == 0 {
		log.Printf("[%s] No se encontraron eventos de exceso de velocidad para procesar en el batch_id %s.", tipoServicioActual, batchID)
		return
	}

	log.Printf("[%s] Se encontraron [%d] nuevos eventos para procesar en el batch_id %s.", tipoServicioActual, len(eventosParaProcesar), batchID)

	ids := make([]int64, len(eventosParaProcesar))
	for i, e := range eventosParaProcesar {
		ids[i] = e.IDConcentrador
	}
	if err := ActualizarEstadoExVelocidad(ids, models.Marcado); err != nil {
		log.Printf("[%s] FALLO CRÍTICO al marcar lote como 'marcado' para batch_id %s: %v. Abortando ciclo.", tipoServicioActual, batchID, err)
		return
	}
	log.Printf("[%s] Lote de %d eventos marcado como 'marcado'. Iniciando procesamiento concurrente...", tipoServicioActual, len(eventosParaProcesar))

	var wg sync.WaitGroup
	for _, evento := range eventosParaProcesar {
		wg.Add(1)
		go func(e models.Evento) {
			defer wg.Done()
			err := InsertarPreEventoExVelocidad(e)

			if err != nil {
				log.Printf("[%s] FALLO al insertar pre-evento para ID %d: %v.", tipoServicioActual, e.IDConcentrador, err)
				if errMarcar := ActualizarEstadoExVelocidad([]int64{e.IDConcentrador}, models.Error); errMarcar != nil {
					log.Printf("[%s] FALLO CRÍTICO al intentar marcar como 'Error' el registro %d: %v", tipoServicioActual, e.IDConcentrador, errMarcar)
				}
			} else {
				if errMarcar := ActualizarEstadoExVelocidad([]int64{e.IDConcentrador}, models.Procesado); errMarcar != nil {
					log.Printf("[%s] FALLO al marcar como 'Procesado' el registro %d: %v", tipoServicioActual, e.IDConcentrador, errMarcar)
				} else {
					log.Printf("[%s] EVENTO PROCESADO: Patente %s, Velocidad %.1f km/h, Límite %.1f km/h, Criticidad %s",
						tipoServicioActual, e.Patente, e.VelocidadRegistrada, e.VelocidadLimite, e.Criticidad)
				}
			}
		}(evento)
	}

	wg.Wait()
	log.Printf("[%s] Procesamiento finalizado para el batch_id: %s", tipoServicioActual, batchID)
}
