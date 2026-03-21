# TP0: Docker + Comunicaciones + Concurrencia

**Nombre:** Lorenzo Minervino  
**Padrón:** 107863

## Cómo ejecutar

Cada ejercicio se encuentra en su rama correspondiente (`ej1`, `ej2`, ..., `ej8`). Para ejecutar un ejercicio específico, pararse en la rama correspondiente y seguir las instrucciones del README de esa rama.

En general, el flujo es:
```bash
./generar-compose.sh docker-compose-dev.yaml 5
make docker-compose-up
make docker-compose-logs
```

Para detener el sistema:
```bash
make docker-compose-down
```

## Estructura del repositorio

Cada rama contiene la solución incremental del ejercicio correspondiente. Las ramas son acumulativas — cada una parte de la anterior e incorpora los cambios pedidos por el enunciado.

Durante el desarrollo se detectaron y corrigieron bugs en ramas anteriores. Dichas correcciones fueron mergeadas hacia adelante, por lo que todas las ramas posteriores incorporan los fixes.

## Tests

Todos los ejercicios pasan las pruebas automáticas de caja negra provistas por la cátedra.
