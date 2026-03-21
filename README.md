## Ejercicio 6

### Cómo ejecutar
Descomprimir los datasets de las agencias:
```bash
unzip .data/dataset.zip -d .data/
```

Generar el compose, construir y levantar:
```bash
./generar-compose.sh docker-compose-dev.yaml 5
make docker-compose-up
make docker-compose-logs
```

Bajar el sistema:
```bash
make docker-compose-down
```

### Cómo verificar
Se debe observar en los logs del cliente batches de 100 apuestas y un último batch menor:
```
client1 | action: apuesta_enviada | result: success | cantidad: 100
client1 | action: apuesta_enviada | result: success | cantidad: 86
client1 | action: loop_finished | result: success | client_id: 1
```

Y en los del servidor:
```
server | action: apuesta_recibida | result: success | cantidad: 100
server | action: apuesta_recibida | result: success | cantidad: 86
```

### Implementación
El protocolo de batch extiende el del ejercicio 5 agregando un header de cantidad:
```
[2 bytes: cantidad de apuestas][apuesta1][apuesta2]...[apuestaN]
```

Donde cada apuesta mantiene el mismo formato que antes:
```
[2 bytes: largo][agency|first_name|last_name|document|birthdate|number]
```

El tamaño máximo del batch es configurable desde `config.yaml` con la clave `batch.maxAmount`. El valor por defecto es 100 apuestas, lo que garantiza que los paquetes no superen los 8kB (~70 bytes por apuesta × 100 = ~7kB).

Los archivos CSV de cada agencia se inyectan como volúmenes en los containers correspondientes siguiendo la convención `.data/agency-{N}.csv`.
