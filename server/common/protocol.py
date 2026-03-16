import struct
import logging

SEPARATOR = "|"
HEADER_SIZE = 2

# Tipos de mensaje del protocolo
MSG_BATCH = 0x01  # batch de apuestas
MSG_FIN   = 0x02  # notificación de fin de envío
MSG_QUERY = 0x03  # consulta de ganadores


def recv_all(sock, n):
    """Lee exactamente n bytes del socket evitando short-read"""
    data = b""
    while len(data) < n:
        chunk = sock.recv(n - len(data))
        if not chunk:
            raise OSError("Conexión cerrada por el cliente")
        data += chunk
    return data


def send_all(sock, data):
    """Envía exactamente todos los bytes evitando short-write"""
    total_sent = 0
    while total_sent < len(data):
        sent = sock.send(data[total_sent:])
        if sent == 0:
            raise OSError("Conexión cerrada por el cliente")
        total_sent += sent


def receive_message_type(sock):
    """
    Lee el byte de tipo de mensaje.
    Retorna el tipo como entero (MSG_BATCH, MSG_FIN o MSG_QUERY)
    """
    data = recv_all(sock, 1)
    return struct.unpack("!B", data)[0]


def receive_bet(sock):
    """
    Recibe una apuesta del cliente.
    Formato: [2 bytes: largo][datos separados por '|']
    Retorna una lista con los campos: [agency, first_name, last_name, document, birthdate, number]
    """
    # Leer header con el largo del mensaje
    header = recv_all(sock, HEADER_SIZE)
    length = struct.unpack("!H", header)[0]

    # Leer exactamente los bytes indicados en el header
    data = recv_all(sock, length)
    return data.decode("utf-8").split(SEPARATOR)


def receive_batch(sock):
    """
    Recibe un batch de apuestas del cliente.
    Formato: [2 bytes: cantidad][apuesta1][apuesta2]...[apuestaN]
    Retorna una lista de listas con los campos de cada apuesta
    """
    # Leer cantidad de apuestas
    header = recv_all(sock, HEADER_SIZE)
    count = struct.unpack("!H", header)[0]

    # Leer cada apuesta individualmente reutilizando receive_bet
    bets = []
    for _ in range(count):
        bets.append(receive_bet(sock))
    return bets


def receive_agency_id(sock):
    """
    Lee el agency_id enviado por el cliente.
    Formato: [2 bytes: agency_id]
    """
    header = recv_all(sock, HEADER_SIZE)
    return struct.unpack("!H", header)[0]


def send_confirmation(sock, msg):
    """
    Envía una confirmación al cliente.
    Formato: [2 bytes: largo][mensaje]
    """
    data = msg.encode("utf-8")
    header = struct.pack("!H", len(data))
    send_all(sock, header + data)


def send_winners(sock, winners):
    """
    Envía la lista de DNIs ganadores al cliente.
    Formato: [2 bytes: largo total][DNI1|DNI2|...]
    Si no hay ganadores envía largo 0.
    """
    if not winners:
        send_all(sock, struct.pack("!H", 0))
        return

    # Unir DNIs con separador y enviar con header de largo
    data = SEPARATOR.join(winners).encode("utf-8")
    header = struct.pack("!H", len(data))
    send_all(sock, header + data)