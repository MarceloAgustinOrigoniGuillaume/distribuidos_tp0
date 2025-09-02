import socket
import logging
import threading
from .server_protocol import ServerProtocol 
from . import utils 

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True

        self.active_connection =None # Needed to force shutdown of current connection.

    def stop(self):
        # Is not async as it is on go or other languages. No need for synchronization.
        logging.info(f'server exiting run loop.')
        self._running = False
        
        self._server_socket.close()

        # For now the handling of active connections is not synchronized/locked since
        # at worst it closes the active connection twice. Not worth the overhead of locking.
        if self.active_connection:
            self.active_connection.close()



    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while self._running:
            try:
                client_sock = self.__accept_new_connection()
                self.__handle_client_connection(client_sock)
            except OSError as e:
                if self._running:
                    logging.error("action: listen_message | result: fail | error: {e}")


        logging.info(f'server exited run loop.')

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            agency = client_sock.recv_str()

            bet = client_sock.recv_bet()

            bet = utils.Bet(agency, bet.first_name, bet.last_name, str(bet.document), bet.birthdate, str(bet.number))
            utils.store_bets([bet])
            logging.info(f"action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}")
        except Exception as e:
            logging.error(f"action: receive_bet | result: fail | error: {e}")
        finally:
            # In the future it would be needed locking. Now its overkill
            if self._running:
                self.active_connection = None
                client_sock.close()


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

        c = ServerProtocol(c)
        self.active_connection = c

        if not self._running:
            self.active_connection = None
            c.close()
            raise OSError("Finished server after accepting connection")
        return c
