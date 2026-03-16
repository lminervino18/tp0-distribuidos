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
server | action: apuesta_recibida | result: success | cantidad: 150
server | action: apuesta_recibida | result: success | cantidad: 150
```

### Implementación
Se modificó el servidor para aceptar conexiones en paralelo usando un `ThreadPoolExecutor` con `max_workers=listen_backlog`. El loop principal solo acepta conexiones y las despacha al pool — no bloquea procesando mensajes.

Todo el estado compartido entre threads (`_finished_agencies`, `_lottery_done`, `_pending_queries`) está protegido por un `threading.Lock()` global. Las llamadas a `store_bets()` y `load_bets()` también se realizan dentro del lock ya que el propio código de la cátedra las declara como **not thread-safe**.

`__get_winners` no adquiere el lock internamente — siempre se llama desde código que ya lo tiene, evitando deadlocks.