# Microservicios de Go

```bash
microservicio_alertas/
├── go.mod
├── go.sum
├── .gitignore
├── .env.example           # Ejemplo de variables de entorno

├── cmd/                    # Puntos de entrada/entry points (main.go) para cada microservicio
│   ├── det_zona_roja
│   │   ├── main.go
│   │   └── Dockerfile     # Dockerfile específico para este microservicio
│   ├── zona_roja_dos
│   │   ├── main.go
│   │   └── Dockerfile     # Dockerfile específico para este microservicio
│   ├── exc_velocidad
│   │   ├── main.go
│   │   └── Dockerfile
│   ├── det_no_autorizada
│   │    ├── main.go
│   │    └── Dockerfile

├── internal/              # Código privado del proyecto 
│   ├── config/            # Lógica de carga de configuración (compartida)
│   │   └── config.go
│   ├── database/          # Conexión a BD, helpers genéricos de queries (compartido)
│   │   └── database.go
│   ├── models/            # Estructuras de datos compartidas
│   │   ├── eventos
│   │   └── camiones.go
│   │   └── geocercas.go
│   │   
│   ├── services/          # Lógica de negocio específica para cada tipo de alerta
│   │   ├── exceso_velocidad/
│   │   │   ├── calculo.go
│   │   ├── zona_roja/
│   │   │   ├── calculo.go
│   │   ├── zona_roja_geohash/
│   │   │   ├── calculo.go

├── README.md
├── docker-compose.yml     # Para desarrollo local, levantar todos los servicios y BD
└── Dockerfile

```
### Índice Espacial (GiST): Esencial para ST_Intersects
```sql
CREATE INDEX idx_geocercas_geometria ON geocercas USING GIST (geometria);
CREATE INDEX idx_concentrador_gps_coordenadas ON concentrador_gps USING GIST (coordenadas);
```
### Índices B-Tree: Para las cláusulas WHERE
```sql
CREATE INDEX idx_geocercas_geometria ON geocercas USING GIST (geometria);
CREATE INDEX idx_concentrador_gps_coordenadas ON concentrador_gps USING GIST (coordenadas);
```