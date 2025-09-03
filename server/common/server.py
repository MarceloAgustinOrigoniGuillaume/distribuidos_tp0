import socket
import logging
import threading
from .server_protocol import ServerProtocol 
from . import utils 

ALL_OK = 0
ERROR_CODE = 1


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

                # In the future it would be needed locking. Now its overkill
                if self._running:
                    self.active_connection = None
                    client_sock.close()                
            except OSError as e:
                if self._running:
                    logging.error(f"action: client handler | result: fail | error: {e}")


        logging.info(f'server exited run loop.')

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        agency = ""
        count = 0
        try:
            agency = client_sock.recv_str()

        except Exception as e:
            logging.error(f"action: receive_bet_count | result: fail | error: {e}")
            return
        total = 0
        received = 0

        try:
            count = client_sock._recv_int32() # Count of bets in batch
            while count > 0:
                total+= count

                logging.info(f"action: client batch recv init | result: success | agency: {agency} | count bets {count}")

                res =[]
                received = 0
                for i in range(count):
                    bet = client_sock.recv_bet()
                    res.append(utils.Bet(agency, bet.first_name, bet.last_name, str(bet.document), bet.birthdate, str(bet.number)))
                    received+=1

                utils.store_bets(res)

                #logging.info(f"action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}")
                logging.info(f"action: apuesta_recibida | result: success | cantidad: {count}")
                client_sock.send_int32(ALL_OK) 
                count = client_sock._recv_int32() # Count of bets in batch
            

            logging.info(f"action: client connection finished | result: success | agency: {agency} | count bets {total}")

        except Exception as e:
            logging.error(f"action: apuesta_recibida | result: fail | cantidad: {count}")
            logging.error(f"total bets recv {total} , batch recv {received} of {count} error: {e}")

            #Send response, possible IO error handled by invoker
            client_sock.send_int32(ERROR_CODE)
            client_sock.send_str(f"{e}")

            return


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
