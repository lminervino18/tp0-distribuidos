## Ejercicio 7

### Cómo ejecutar
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
Se debe observar el sorteo y las consultas de ganadores:
```
server | action: sorteo | result: success
client1 | action: consulta_ganadores | result: success | cant_ganadores: 2
```

### Protocolo
Se agregaron dos nuevos tipos de mensaje al protocolo, discriminados por un byte de tipo al inicio de cada mensaje:
```
0x01 → MSG_BATCH   batch de apuestas
0x02 → MSG_FIN     notificación de fin de envío
0x03 → MSG_QUERY   consulta de ganadores
```

La respuesta del servidor con los ganadores usa el formato:
```
[2 bytes: largo total][DNI1|DNI2|...]
```

### Implementación
Cada cliente al terminar de enviar todos sus batches abre una conexión nueva y envía `MSG_FIN` con su agency_id. Luego abre otra conexión y envía `MSG_QUERY` quedando bloqueado esperando la respuesta.

El servidor mantiene una **cola de conexiones pendientes** (`_pending_queries`) — cuando una agencia consulta antes del sorteo, el servidor no responde ni cierra la conexión, sino que la guarda en la cola. Cuando llega el último `MSG_FIN` (la agencia N), el servidor realiza el sorteo y responde a todas las conexiones pendientes antes de aceptar nuevas. Esto garantiza que ninguna agencia recibe información parcial.

El número de agencias esperadas es configurable mediante la variable de entorno `TOTAL_AGENCIES` que es obtenida por la cantidad de clientes.