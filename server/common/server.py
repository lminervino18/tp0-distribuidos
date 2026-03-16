import os
import socket
import logging
import signal

from common.protocol import (
    MSG_BATCH, MSG_FIN, MSG_QUERY,
    receive_message_type, receive_batch, receive_agency_id,
    send_confirmation, send_winners
)
from common.utils import Bet, store_bets, load_bets, has_won

# Total de agencias esperadas — configurable por variable de entorno
TOTAL_AGENCIES = int(os.getenv('TOTAL_AGENCIES', 5))


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True

        # Conjunto de agencias que notificaron fin
        self._finished_agencies = set()
        # Indica si el sorteo ya fue realizado
        self._lottery_done = False
        # Cola de conexiones pendientes que consultaron antes del sorteo
        # agency_id -> client_sock
        self._pending_queries = {}

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

        Accepts new connections and dispatches each message type.
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
        Lee el tipo de mensaje y despacha al handler correspondiente.
        Las conexiones de consulta pendientes no se cierran acá —
        se cierran cuando el sorteo esté listo.
        """
        try:
            msg_type = receive_message_type(client_sock)

            if msg_type == MSG_BATCH:
                self.__handle_batch(client_sock)
                client_sock.close()
                logging.info('action: close_client_socket | result: success')
            elif msg_type == MSG_FIN:
                self.__handle_fin(client_sock)
                client_sock.close()
                logging.info('action: close_client_socket | result: success')
            elif msg_type == MSG_QUERY:
                # Si el sorteo ya está listo responder inmediatamente
                # sino guardar la conexión para responder después
                self.__handle_query(client_sock)
            else:
                logging.error(f'action: receive_message | result: fail | error: unknown type {msg_type}')
                client_sock.close()

        except OSError as e:
            logging.error(f'action: receive_message | result: fail | error: {e}')
            client_sock.close()

    def __handle_batch(self, client_sock):
        """Recibe un batch de apuestas, las persiste y confirma al cliente"""
        fields_list = receive_batch(client_sock)
        bets = [Bet(f[0], f[1], f[2], f[3], f[4], f[5]) for f in fields_list]

        store_bets(bets)
        logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
        send_confirmation(client_sock, 'success')

    def __handle_fin(self, client_sock):
        """
        Registra que una agencia terminó de enviar apuestas.
        Cuando llegan las 5 agencias realiza el sorteo y responde
        a todas las conexiones pendientes.
        """
        agency_id = receive_agency_id(client_sock)
        self._finished_agencies.add(agency_id)
        logging.info(f'action: fin_recibido | result: success | agency_id: {agency_id} | total: {len(self._finished_agencies)}')

        # Cuando todas las agencias terminaron realizar el sorteo
        if len(self._finished_agencies) == TOTAL_AGENCIES:
            self._lottery_done = True
            logging.info('action: sorteo | result: success')

            # Responder a todas las conexiones que estaban esperando
            pending = list(self._pending_queries.items())
            self._pending_queries.clear()

            for pending_agency_id, pending_sock in pending:
                try:
                    winners = self.__get_winners(pending_agency_id)
                    send_winners(pending_sock, winners)
                    logging.info(f'action: consulta_ganadores | result: success | agency_id: {pending_agency_id} | cant_ganadores: {len(winners)}')
                except OSError as e:
                    logging.error(f'action: consulta_ganadores | result: fail | agency_id: {pending_agency_id} | error: {e}')
                finally:
                    pending_sock.close()
                    logging.info('action: close_client_socket | result: success')

    def __handle_query(self, client_sock):
        """
        Responde con los DNIs ganadores de la agencia.
        Si el sorteo no está listo guarda la conexión en la cola
        y la responde cuando llegue el último fin.
        """
        agency_id = receive_agency_id(client_sock)

        if self._lottery_done:
            # Sorteo ya realizado, responder inmediatamente
            try:
                winners = self.__get_winners(agency_id)
                send_winners(client_sock, winners)
                logging.info(f'action: consulta_ganadores | result: success | agency_id: {agency_id} | cant_ganadores: {len(winners)}')
            except OSError as e:
                logging.error(f'action: consulta_ganadores | result: fail | agency_id: {agency_id} | error: {e}')
            finally:
                client_sock.close()
                logging.info('action: close_client_socket | result: success')
        else:
            # Guardar conexión para responder cuando el sorteo esté listo
            logging.info(f'action: consulta_ganadores | result: waiting | agency_id: {agency_id}')
            self._pending_queries[agency_id] = client_sock

    def __get_winners(self, agency_id):
        """Retorna los DNIs ganadores de una agencia específica"""
        return [
            bet.document
            for bet in load_bets()
            if bet.agency == agency_id and has_won(bet)
        ]

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