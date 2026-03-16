import struct
import logging

SEPARATOR = "|"
HEADER_SIZE = 2


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


def receive_bet(sock):
    """
    Recibe una apuesta del cliente.
    Formato: [2 bytes big-endian: largo][datos separados por '|']
    Retorna una lista con los campos: [agency, first_name, last_name, document, birthdate, number]
    """
    # Leer header con el largo del mensaje
    header = recv_all(sock, HEADER_SIZE)
    # Desempaquetar el largo como entero big-endian de 2 bytes
    length = struct.unpack("!H", header)[0]

    # Leer exactamente los bytes indicados en el header
    data = recv_all(sock, length)
    fields = data.decode("utf-8").split(SEPARATOR)
    return fields


def send_confirmation(sock, msg):
    """
    Envía una confirmación al cliente.
    Formato: [2 bytes big-endian: largo][mensaje]
    """
    data = msg.encode("utf-8")
    # Empaquetar el largo como entero big-endian de 2 bytes
    header = struct.pack("!H", len(data))
    send_all(sock, header + data)