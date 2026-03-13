## Ejercicio 3

### Cómo ejecutar
Levantar solo el servidor (sin clientes, no afectan pero no son necesarios):
```bash
./generar-compose.sh docker-compose-dev.yaml 0
make docker-compose-up
```

Ejecutar el script de validación:
```bash
./validar-echo-server.sh
```

Bajar el sistema:
```bash
make docker-compose-down
```

### Implementación
El script `validar-echo-server.sh` corre netcat dentro de un container temporal de `busybox` (que ya tiene netcat instalado) conectado a la red `tp0_testing_net`, sin exponer puertos al host. Envía un mensaje al servidor y compara la respuesta. Si son iguales imprime `result: success`, caso contrario `result: fail`.