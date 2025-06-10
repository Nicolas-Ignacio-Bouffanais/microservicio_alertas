package zona_roja

import (
	"log"
	"sync"

	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/database"
	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/models"
)

const tipoServicioActual = "DET_ZONA_ROJA"
const loteSize = 100

func ProcesarEventosZonaRoja() {
	log.Println("Iniciando procesamiento del servicio Zona Roja...")

	fechaDeProceso := "2025-05-12"

	eventosParaInsertar, err := database.GetInterseccionesZonaRoja(fechaDeProceso)
	if err != nil {
		log.Printf("Error al obtener intersecciones de zona roja: %v", err)
		return
	}

	if len(eventosParaInsertar) == 0 {
		log.Println("No se encontraron nuevas intersecciones en zonas rojas.")
		return
	}

	log.Printf("Se encontraron [%d] nuevas intersecciones para procesar en lotes de %d.", len(eventosParaInsertar), loteSize)

	var wg sync.WaitGroup

	for i := 0; i < len(eventosParaInsertar); i += loteSize {
		fin := i + loteSize
		if fin > len(eventosParaInsertar) {
			fin = len(eventosParaInsertar)
		}
		// Obtenemos el lote actual que vamos a procesar
		lote := eventosParaInsertar[i:fin]

		wg.Add(1)
		// 3. Lanzamos UNA goroutine por CADA LOTE, no por cada evento.
		go func(loteAProcesar []models.Evento) {
			defer wg.Done()

			// 4. Dentro de la goroutine, procesamos cada evento del lote secuencialmente.
			for _, eventoData := range loteAProcesar {
				err := database.InsertarEventoZonaRoja(eventoData)
				if err != nil {
					log.Printf("[%s] FALLO al insertar evento para Camión %s: %v", tipoServicioActual, eventoData.Patente, err)
				} else {
					if eventoData.IDGeocerca != nil {
						log.Printf("¡EVENTO INSERTADO! Camión %s en Geocerca ID %s.", eventoData.Patente, *eventoData.IDGeocerca)
					} else {
						log.Printf("¡EVENTO INSERTADO! Camión %s (sin geocerca asociada).", eventoData.Patente)
					}
				}
			}
		}(lote)
	}

	wg.Wait() // Esperamos a que todas las goroutines de los lotes terminen.
	log.Printf("[%s] Ciclo de procesamiento finalizado.", tipoServicioActual)
}
