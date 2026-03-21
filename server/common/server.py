import socket
import logging
import signal

from common.protocol import receive_batch, send_confirmation
from common.utils import Bet, store_bets


class Server:
    def __init__(self, port, listen_backlog):
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True
        signal.signal(signal.SIGTERM, self.__handle_sigterm)

    def __handle_sigterm(self, sig, frame):
        logging.info('action: receive_sigterm | result: success')
        self._running = False
        self._server_socket.close()
        logging.info('action: close_server_socket | result: success')

    def run(self):
        while self._running:
            try:
                client_sock = self.__accept_new_connection()
                self.__handle_client_connection(client_sock)
            except OSError as e:
                if self._running:
                    logging.error(f'action: accept_connections | result: fail | error: {e}')

        logging.info('action: server_shutdown | result: success')

    def __handle_client_connection(self, client_sock):
        try:
            fields_list = receive_batch(client_sock)
            bets = [Bet(f[0], f[1], f[2], f[3], f[4], f[5]) for f in fields_list]
            store_bets(bets)
            logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
            send_confirmation(client_sock, 'success')
        except OSError as e:
            logging.error(f'action: apuesta_recibida | result: fail | cantidad: 0 | error: {e}')
            try:
                send_confirmation(client_sock, 'fail')
            except OSError:
                pass
        finally:
            client_sock.close()
            logging.info('action: close_client_socket | result: success')

    def __accept_new_connection(self):
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
