## Ejercicio 5

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
Se debe observar en los logs del cliente:
```
client1 | action: apuesta_enviada | result: success | dni: 30904461 | numero: 7571
```

Y en los del servidor:
```
server | action: apuesta_almacenada | result: success | dni: 30904461 | numero: 7571
```

### Protocolo de comunicación
Se implementó un protocolo binario propio con el siguiente formato:
```
[2 bytes big-endian: largo del mensaje][mensaje en texto plano]
```

Los campos de la apuesta se serializan separados por `|`:
```
agency|first_name|last_name|document|birthdate|number
```

El mismo formato se usa para la confirmación del servidor. Se evitan short-read y short-write mediante loops que garantizan la lectura/escritura de exactamente N bytes.

### Implementación
La lógica se modularizó en tres capas:

**Modelo de dominio:** `bet.go` (cliente) y `utils.py` (servidor) contienen el struct/clase `Bet` sin lógica de comunicación.

**Capa de comunicación:** `protocol.go` (cliente) y `protocol.py` (servidor) contienen la serialización, `sendAll`/`recvAll` y el protocolo de longitud prefija.

**Lógica de negocio:** `client.go` y `server.py` orquestan el flujo usando las capas anteriores.