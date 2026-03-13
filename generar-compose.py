import sys

# Lee los argumentos que le pasó el bash
# sys.argv[0] es el nombre del script
# sys.argv[1] es el nombre del archivo de salida
# sys.argv[2] es la cant de clientes
output_file = sys.argv[1]
num_clients = int(sys.argv[2])

# La parte del server es fija
content = """name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net
"""

# Por cada cliente genera un bloque idéntico
# solo cambiando el número
for i in range(1, num_clients + 1):
    content += f"""
  client{i}:
    container_name: client{i}
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID={i}
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server
"""

# La definición de red que también es fija
content += """
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
"""

# Escribe todo el contenido en el archivo de salida
with open(output_file, 'w') as f:
    f.write(content)

print(f"Archivo {output_file} generado con {num_clients} clientes")