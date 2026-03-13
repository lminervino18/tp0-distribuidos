## Ejercicio 2

### Cómo ejecutar
Generar el compose y levantar el sistema:
```bash
./generar-compose.sh docker-compose-dev.yaml 5
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

### Cómo verificar
Modificar algún valor en `client/config.yaml` o `server/config.ini` (por ejemplo cambiar `loop.amount` a 2) y reiniciar sin reconstruir:
```bash
make docker-compose-down
docker compose -f docker-compose-dev.yaml up -d
make docker-compose-logs
```

No hacemos make docker-compose-up ya que esto hace build internamente
Si los cambios se reflejan sin haber ejecutado `docker build`, el bind mount funciona correctamente.

### Implementación
Los archivos de configuración (`config.ini` del servidor y `config.yaml` del cliente) son inyectados en los containers mediante bind mounts definidos en el docker-compose generado. Esto permite modificar la configuración sin reconstruir las imágenes.

El `config.ini` fue excluido de la imagen del servidor mediante un `.dockerignore`. En el cliente se eliminó la línea `COPY` del config en el Dockerfile.