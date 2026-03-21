## Ejercicio 8

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
Se debe observar que el servidor procesa conexiones de múltiples clientes en paralelo — los logs de distintas IPs se intercalan:
```
server | action: accept_connections | result: success | ip: 172.25.125.3
server | action: accept_connections | result: success | ip: 172.25.125.4
server | action: apuesta_recibida | result: success | cantidad: 100
server | action: apuesta_recibida | result: success | cantidad: 100
```

### Implementación
Se modificó el servidor para aceptar conexiones en paralelo usando `threading.Thread`. Por cada conexión entrante el loop principal lanza un thread dedicado con `daemon=True`, no bloquea esperando que termine el procesamiento antes de aceptar la siguiente conexión.

La implementación usa `threading` directamente en lugar de `concurrent.futures.ThreadPoolExecutor`, ya que la cátedra no permite el uso de Futures/Asyncio en este TP0 por abstraer demasiado la coordinación de hilos. En un principi lo modele con `concurrent.futures.ThreadPoolExecutor` (y me funcionaba bien) pero lei en el campus que no se podia usar

Todo el estado compartido entre threads (`_finished_agencies`, `_lottery_done`, `_pending_queries`) está protegido por un `threading.Lock()` global. Las llamadas a `store_bets()` y `load_bets()` también se realizan dentro del lock ya que el propio código de la cátedra las declara como **not thread-safe**.

`__get_winners` no adquiere el lock internamente, siempre se llama desde código que ya lo tiene, evitando deadlocks.
