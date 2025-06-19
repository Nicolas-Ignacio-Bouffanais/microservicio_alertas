package det_no_autorizada

import (
	"log"
	"sync"

	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/models"
)

const tipoServicioActual = "DETENCION_ILEGAL"

func ProcesarEventos(batchID string) {
	log.Printf("[%s] Notificación recibida. Iniciando procesamiento para batch_id: %s", tipoServicioActual, batchID)

	eventosParaProcesar, err := GetDetencionesIlegales(batchID)
	if err != nil {
		log.Printf("[%s] Error al obtener eventos de detenciones para batch_id %s: %v", tipoServicioActual, batchID, err)
		return
	}

	if len(eventosParaProcesar) == 0 {
		log.Printf("[%s] No se encontraron nuevos eventos para procesar en el batch_id %s.", tipoServicioActual, batchID)
		return
	}

	log.Printf("[%s] Se encontraron [%d] nuevos eventos para procesar en el batch_id %s.", tipoServicioActual, len(eventosParaProcesar), batchID)

	// 2. Marcar todos los eventos encontrados como "marcado" antes de procesarlos.
	ids := make([]int64, len(eventosParaProcesar))
	for i, e := range eventosParaProcesar {
		ids[i] = e.IDConcentrador
	}
	if err := ActualizarEstadoDetNoAut(ids, models.Marcado); err != nil {
		log.Printf("[%s] FALLO CRÍTICO al marcar lote como 'marcado' para batch_id %s: %v. Abortando ciclo.", tipoServicioActual, batchID, err)
		return
	}
	log.Printf("[%s] Lote de %d eventos marcado como 'marcado'. Iniciando procesamiento concurrente...", tipoServicioActual, len(eventosParaProcesar))

	// 3. Procesar cada evento de forma concurrente.
	var wg sync.WaitGroup
	for _, evento := range eventosParaProcesar {
		wg.Add(1)
		go func(e models.PreEventoDetNoAutorizada) {
			defer wg.Done()

			// Insertar el pre-evento en la tabla de resultados.
			if err := InsertarPreEventoDetNoAut(e); err != nil {
				log.Printf("[%s] FALLO al insertar pre-evento para ID %d: %v.", tipoServicioActual, e.IDConcentrador, err)
				// Si falla la inserción, marcar el registro como "Error".
				if errMarcar := ActualizarEstadoDetNoAut([]int64{e.IDConcentrador}, models.Error); errMarcar != nil {
					log.Printf("[%s] FALLO CRÍTICO al intentar marcar como 'Error' el registro %d: %v", tipoServicioActual, e.IDConcentrador, errMarcar)
				}
			} else {
				// Si la inserción es exitosa, marcar como "Procesado".
				if errMarcar := ActualizarEstadoDetNoAut([]int64{e.IDConcentrador}, models.Procesado); errMarcar != nil {
					log.Printf("[%s] FALLO al marcar como 'Procesado' el registro %d: %v", tipoServicioActual, e.IDConcentrador, errMarcar)
				} else {
					log.Printf("[%s] EVENTO PROCESADO: Patente %s, Criticidad %s",
						tipoServicioActual, e.Patente, e.Criticidad)
				}
			}
		}(evento)
	}

	wg.Wait()
	log.Printf("[%s] Procesamiento finalizado para el batch_id: %s", tipoServicioActual, batchID)
}
