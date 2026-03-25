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
El cliente mantiene una única conexión TCP durante toda la transmisión de batches, enviando todos los `MSG_BATCH` por esa conexión y cerrándola al terminar. Luego abre una conexión nueva para enviar `MSG_FIN` con su agency_id, y finalmente abre otra conexión para enviar `MSG_QUERY` quedando bloqueado esperando la respuesta.

El servidor loopea leyendo mensajes sobre la misma conexión mientras recibe batches. Al detectar el cierre de la conexión de batches vuelve a aceptar nuevas conexiones para atender el `MSG_FIN` y el `MSG_QUERY`.

El servidor mantiene una **cola de conexiones pendientes** (`_pending_queries`) — cuando una agencia consulta antes del sorteo, el servidor no responde ni cierra la conexión, sino que la guarda en la cola. Cuando llega el último `MSG_FIN` (la agencia N), el servidor realiza el sorteo y responde a todas las conexiones pendientes antes de aceptar nuevas. Si el sorteo ya fue realizado cuando llega el `MSG_QUERY`, el servidor responde inmediatamente sin encolar la conexión. Esto garantiza que ninguna agencia recibe información parcial.

El número de agencias esperadas es configurable mediante la variable de entorno `TOTAL_AGENCIES` que es obtenida por la cantidad de clientes.
