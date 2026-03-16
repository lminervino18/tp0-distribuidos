import socket
import logging
import signal

from common.protocol import receive_bet, send_confirmation
from common.utils import Bet, store_bets


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True

        # Registrar handler de SIGTERM para graceful shutdown
        signal.signal(signal.SIGTERM, self.__handle_sigterm)

    def __handle_sigterm(self, sig, frame):
        """
        Handle SIGTERM signal for graceful shutdown.
        Closes the server socket and stops the main loop.
        """
        logging.info('action: receive_sigterm | result: success')
        self._running = False
        self._server_socket.close()
        logging.info('action: close_server_socket | result: success')

    def run(self):
        """
        Server loop

        Accepts new connections and handles each client.
        After client communication finishes, server starts to accept new connections again.
        """
        while self._running:
            try:
                client_sock = self.__accept_new_connection()
                self.__handle_client_connection(client_sock)
            except OSError as e:
                if self._running:
                    logging.error(f'action: accept_connections | result: fail | error: {e}')

        logging.info('action: server_shutdown | result: success')

    def __handle_client_connection(self, client_sock):
        """
        Read bet from a specific client socket, store it and send confirmation.

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            # Recibir los campos de la apuesta usando el protocolo
            fields = receive_bet(client_sock)
            bet = Bet(fields[0], fields[1], fields[2], fields[3], fields[4], fields[5])

            # Persistir la apuesta usando la función provista por la cátedra
            store_bets([bet])
            logging.info(f'action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}')

            # Enviar confirmación al cliente
            send_confirmation(client_sock, 'success')

        except OSError as e:
            logging.error(f'action: receive_message | result: fail | error: {e}')
        finally:
            client_sock.close()
            logging.info('action: close_client_socket | result: success')

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """
        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c