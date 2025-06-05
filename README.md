# Microservicios de Go

```bash
microservicio_alertas/
├── go.mod
├── go.sum
├── .gitignore
├── .env.example           # Ejemplo de variables de entorno

├── cmd/                    # Puntos de entrada/entry points (main.go) para cada microservicio
│   ├── detencion_zona_roja
│   │   ├── main.go
│   │   └── Dockerfile     # Dockerfile específico para este microservicio
│   ├── exceso_velocidad
│   │   ├── main.go
│   │   └── Dockerfile
│   ├── detencion_no_autorizada
│   │    ├── main.go
│   │    └── Dockerfile
├── internal/              # Código privado del proyecto 
│   ├── config/            # Lógica de carga de configuración (compartida)
│   │   └── config.go
│   ├── database/          # Conexión a BD, helpers genéricos de queries (compartido)
│   │   └── database.go
│   ├── models/            # Estructuras de datos compartidas
│   │   ├── common_alert.go  # Estructura para la tabla alertas_alertas
│   │   └── camiones.go
│   │   └── geocercas.go
│   │   
│   ├── services/          # Lógica de negocio específica para cada tipo de alerta
│   │   ├── detencion_zona_roja/
│   │   │   ├── calculo.go # Lógica de cálculo de pre-eventos para zonas rojas
│   │   │   └── alertas.go   # Lógica de agrupación de pre-eventos en alertas para zonas
│   │   │   └── models.go    # Modelos específicos para esta alerta (ej. PreEventoZonaRoja)
│   │   ├── exceso_velocidad/
│   │   │   ├── calculo.go
│   │   │   └── alertas.go
│   │   │   └── models.go
│   │   ├── detencion_zona_roja/
│   │   │   ├── calculo.go
│   │   │   └── alertas.go
│   │   │   └── models.go

├── README.md
├── docker-compose.yml     # Para desarrollo local, levantar todos los servicios y BD
└── Dockerfile

```

```bash
docker-compose up -d        #Correr la imagen de docker
```