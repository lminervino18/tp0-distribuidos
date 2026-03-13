## Ejercicio 4

### Cómo ejecutar
```bash
./generar-compose.sh docker-compose-dev.yaml 5
make docker-compose-up
```

Para probar el graceful shutdown mientras los clientes corren:
```bash
make docker-compose-down
```

### Cómo verificar
Levantar el sistema y bajar antes de que los clientes terminen solos. Se debe observar en los logs:
```
client1 | action: receive_sigterm | result: success | client_id: 1
client1 | action: close_connection | result: success | client_id: 1
client1 exited with code 0
server  | action: receive_sigterm | result: success
server  | action: close_server_socket | result: success
server  | action: server_shutdown | result: success
server exited with code 0
```

### Implementación
**Servidor (Python):** se registra un handler para SIGTERM con `signal.signal()`. Al recibirlo, cierra el server socket y termina el loop principal. El `accept()` bloqueado lanza `OSError` que es catcheada correctamente.

**Cliente (Go):** se crea un channel `sigChan` que recibe SIGTERM. Una goroutine escucha en paralelo al loop principal. Al recibir SIGTERM cierra la conexión activa y notifica al loop mediante un `stopChan`. El loop usa `select` para poder interrumpir el `time.Sleep` inmediatamente en lugar de esperar que Docker mande SIGKILL.