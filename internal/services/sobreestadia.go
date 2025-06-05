package services

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Nicolas-Ignacio-Bouffanais/microservicio_alertas/internal/database"
)

func ProcesarZonaOrigen() {
	log.Println("Iniciando procesamiento del servicio Zona Origen...")

	geocercas, err := database.GetZonasOrigen()
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

	fmt.Println("-----------------------------------------------------")

	camionesDetenidos, err := database.GetCamionesDetenidos()
	if err != nil {
		log.Printf("Error al obtener camiones detenidos: %v", err)
		return
	}

	if len(camionesDetenidos) == 0 {
		log.Println("No se encontraron camiones detenidos.")
	} else {
		log.Printf("Camiones detenidos obtenidos: %d", len(camionesDetenidos))
		for i, camion := range camionesDetenidos {
			if i < 5 {
				camionJSON, err := json.MarshalIndent(camion, "", "  ")
				if err != nil {
					log.Printf("Error al convertir camión a JSON (Patente: %s): %v", camion.Patente, err)
				} else {
					fmt.Printf("\n--- Camión Detenido %d ---\n%s\n", i+1, string(camionJSON))
				}
			} else {
				break
			}
		}
		if len(camionesDetenidos) > 5 {
			fmt.Printf("... y %d camiones detenidos más.\n", len(camionesDetenidos)-5)
		}
	}

	log.Println("Procesamiento del servicio Zona Origen completado por ahora.")
}
