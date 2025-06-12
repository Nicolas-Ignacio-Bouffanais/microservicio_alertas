package zona_roja_marcado

import (
	"log"
	"sync"

	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/models"
)

const tipoServicioActual = "DET_ZONA_ROJA_MARCADO"
const loteSize = 100

func ProcesarEventos() {
	log.Printf("[%s] Iniciando ciclo de procesamiento...", tipoServicioActual)

	eventosParaProcesar, err := GetInterseccionesNoMarcadas()
	if err != nil {
		log.Printf("[%s] Error al obtener intersecciones no marcadas: %v", tipoServicioActual, err)
		return
	}

	if len(eventosParaProcesar) == 0 {
		log.Printf("[%s] No se encontraron nuevos eventos para procesar.", tipoServicioActual)
		return
	}

	log.Printf("[%s] Se encontraron [%d] nuevos eventos para procesar.", tipoServicioActual, len(eventosParaProcesar))

	var wg sync.WaitGroup
	for i := 0; i < len(eventosParaProcesar); i += loteSize {
		lote := eventosParaProcesar[i:min(i+loteSize, len(eventosParaProcesar))]

		wg.Add(1)
		go func(loteAProcesar []models.Evento) {
			defer wg.Done()
			for _, evento := range loteAProcesar {
				// PASO A: Bloquear el registro con el estado 'marcado'.
				// Llama a la función correcta 'ActualizarEstadoZonaRoja'.
				if err := ActualizarEstadoZonaRoja(evento.IDConcentrador, models.Marcado); err != nil {
					log.Printf("[%s] FALLO al marcar como 'En Proceso' el registro %d: %v", tipoServicioActual, evento.IDConcentrador, err)
					continue
				}

				// PASO B: Insertar el pre-evento.
				if err := InsertarPreEvento(evento); err != nil {
					log.Printf("[%s] FALLO al insertar pre-evento para patente %s (ID Concentrador: %d): %v", tipoServicioActual, evento.Patente, evento.IDConcentrador, err)

					// Si falla, se marca como 'error' para no reintentar indefinidamente.
					if errMarcar := ActualizarEstadoZonaRoja(evento.IDConcentrador, models.Error); errMarcar != nil {
						log.Printf("[%s] FALLO CRÍTICO al intentar marcar como 'Error' el registro %d: %v", tipoServicioActual, evento.IDConcentrador, errMarcar)
					}
					continue
				}

				// PASO C: Marcar como 'procesado' si todo fue exitoso.
				if err := ActualizarEstadoZonaRoja(evento.IDConcentrador, models.Procesado); err != nil {
					log.Printf("[%s] FALLO al marcar como 'Procesado' el registro %d: %v", tipoServicioActual, evento.IDConcentrador, err)
				} else {
					log.Printf("[%s] ¡EVENTO PROCESADO! Patente: %s, ID Concentrador: %d", tipoServicioActual, evento.Patente, evento.IDConcentrador)
				}
			}
		}(lote)
	}

	wg.Wait()
	log.Printf("[%s] Ciclo de procesamiento finalizado.", tipoServicioActual)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
