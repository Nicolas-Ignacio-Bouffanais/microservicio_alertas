package det_no_autorizada

import (
	"log"
	"sync"

	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/models"
)

const tipoServicioActual = "DETENCION_ILEGAL"
const loteSize = 100

func ProcesarEventos() {
	log.Printf("[%s] Iniciando ciclo de procesamiento...", tipoServicioActual)

	eventosParaProcesar, err := GetDetencionesIlegales()
	if err != nil {
		log.Printf("[%s] ERROR FATAL al obtener detenciones: %v", tipoServicioActual, err)
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
				// PASO 1: Intentar insertar el pre-evento.
				if err := InsertarPreEventoDetNoAut(evento); err != nil {
					log.Printf("[%s] FALLO al insertar pre-evento para ID %d: %v. Marcando como ERROR.", tipoServicioActual, evento.IDConcentrador, err)

					if errMarcar := ActualizarEstadoDetNoAut(evento.IDConcentrador, models.Error); errMarcar != nil {
						log.Printf("[%s] FALLO CRÍTICO al intentar marcar como 'Error' el registro %d: %v", tipoServicioActual, evento.IDConcentrador, errMarcar)
					}
					continue
				}

				// PASO 2: Si el paso anterior fue exitoso, marcar como PROCESADO.
				if err := ActualizarEstadoDetNoAut(evento.IDConcentrador, models.Procesado); err != nil {
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
