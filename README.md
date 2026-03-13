## Ejercicio 1

### Cómo ejecutar
Generar el archivo de Docker Compose con N clientes:
```bash
./generar-compose.sh <archivo_salida> <cantidad_clientes>
```

Por ejemplo con 5 clientes:
```bash
./generar-compose.sh docker-compose-dev.yaml 5
```

Levantar el sistema:
```bash
make docker-compose-up
```

Ver logs:
```bash
make docker-compose-logs
```

Bajar el sistema:
```bash
make docker-compose-down
```

### Implementación
El script `generar-compose.sh` invoca a `generar-compose.py` que genera dinámicamente el archivo YAML con el servidor y N clientes.