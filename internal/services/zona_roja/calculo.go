package zona_roja

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/database"
	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/models"
)

const tipoServicioActual = "DET_ZONA_ROJA"
const loteSize = 100

func ProcesarEventosZonaRoja() {
	log.Println("Iniciando procesamiento del servicio Zona Roja...")

	geocercas, err := database.GetZonasRojas()
	if err != nil {
		log.Printf("Error al obtener geocercas: %v", err)
		return
	}

	if len(geocercas) == 0 {
		log.Println("No se encontraron geocercas para procesar.")
	} else {
		log.Printf("Geocercas obtenidas: %d", len(geocercas))
		for i, geo := range geocercas {
			if i < 3 {
				geoJSON, err := json.MarshalIndent(geo, "", "  ")
				if err != nil {
					log.Printf("Error al convertir geocerca a JSON (ID: %d): %v", geo.IDGeocerca, err)
				} else {
					fmt.Printf("\n--- Geocerca %d ---\n%s\n", i+1, string(geoJSON))
				}
			} else {
				break
			}
		}
		if len(geocercas) > 3 {
			fmt.Printf("... y %d geocercas más.\n", len(geocercas)-3)
		}
	}

	camionesParaProcesar, err := database.GetCamionesDetenidos()
	if err != nil {
		log.Printf("Error al obtener camiones detenidos: %v", err)
		return
	}

	if len(camionesParaProcesar) == 0 {
		log.Println("No se encontraron camiones detenidos.")
	}
	log.Printf("[%d] camiones detenidos para procesar.", len(camionesParaProcesar))

	var wg sync.WaitGroup

	for i := 0; i < len(camionesParaProcesar); i += loteSize {
		fin := i + loteSize
		if fin > len(camionesParaProcesar) {
			fin = len(camionesParaProcesar)
		}
		lote := camionesParaProcesar[i:fin]

		wg.Add(1)
		// Lanzamos una goroutine para procesar cada lote de camiones
		go func(loteCamiones []models.Camion) {
			defer wg.Done()
			for _, camion := range loteCamiones {
				// La lógica de marcar, comparar y generar eventos para un solo camión
				procesarUnCamion(camion, geocercas)
			}
		}(lote)
	}

	wg.Wait() // Esperamos a que todas las goroutines de lotes terminen
	log.Printf("[%s] Ciclo de procesamiento finalizado.", tipoServicioActual)
}

func procesarUnCamion(camion models.Camion, geocercas []models.Geocerca) {
	// A FUTURO: Aquí es donde se reincorporaría la lógica de marcado

	// 2. Verificar intersección
	if camion.CoordenadasWKT == nil {
		return
	}

	intersecta, idGeo, errInt := database.CamionIntersectaAlgunaGeocerca(*camion.CoordenadasWKT, geocercas) //
	if errInt != nil {
		log.Printf("[%s] Error de intersección para camión ID %d: %v", tipoServicioActual, camion.Id, errInt)
		return
	}

	if intersecta { // FechaHora es la fecha en la que se termino de calcular y se inserto en la tabla de preeventos
		log.Printf("¡INTERSECCIÓN ENCONTRADA! Camión %s en Geocerca ID %s. Insertando evento...", camion.Patente, idGeo)
		e := models.Evento{
			Patente:        camion.Patente,
			IDGeocerca:     &idGeo,
			CoordenadasWKT: camion.CoordenadasWKT,
			Velocidad:      camion.Velocidad,
			Orientacion:    camion.Orientacion,
		}
		err := database.InsertarEventoZonaRoja(e) //
		if err != nil {
			log.Printf("[%s] FALLO al insertar evento para Camión %s: %v", tipoServicioActual, camion.Patente, err)
		}
	}
}
