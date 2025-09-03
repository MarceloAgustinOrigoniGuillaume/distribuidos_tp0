import socket
import logging
import threading
from .server_protocol import ServerProtocol 
from . import utils 

ALL_OK = 0
ERROR_CODE = 1
WINNERS_EOF = -1

class Agency:
    def __init__(self, connection):
        self.conn = connection

    def notify_winner(self, bet):
        self.conn.send_int32(bet.number)
        self.conn.send_str(bet.document)

    def finished_winners(self):
        self.conn.send_int32(WINNERS_EOF)

    def close(self):
        self.conn.close()

class Server:
    def __init__(self, port, listen_backlog, agency_count):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True
        self.agency_count = agency_count

        self.active_connection =None # Needed to force shutdown of current connection.

        # Agency is actually a number, but to make it more flexible it will be used a hashmap
        self.awaiting_agencies = {}

    def stop(self):
        # Is not async as it is on go or other languages. No need for synchronization.
        logging.info(f'action: server_exiting_run_loop. | result: in_progress')
        self._running = False
        
        self._server_socket.close()

        # For now the handling of active connections is not synchronized/locked since
        # at worst it closes the active connection twice. Not worth the overhead of locking.
        if self.active_connection:
            self.active_connection.close()

        for agency in self.awaiting_agencies:
            agency.close()



    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        awaiting_count = 0
        while self._running:
            try:
                client_sock = self.__accept_new_connection()
                
                if self.__handle_client_connection(client_sock):
                    awaiting_count+=1
                    if awaiting_count == self.agency_count:
                        winning_bets = filter(utils.has_won, utils.load_bets())

                        # Since its not parallel it is not needed to group them by before sending them
                        for bet in winning_bets:
                            self.awaiting_agencies[bet.agency].notify_winner(bet)
                        
                        
                        for agency in self.awaiting_agencies.values():
                            agency.finished_winners()
                            agency.close()

                        logging.info("action: sorteo | result: success")

                elif self._running:
                    self.active_connection = None
                    client_sock.close()

            except OSError as e:
                if self._running:
                    logging.error(f"action: client handler | result: fail | error: {e}")


        logging.info(f'action: server_exited_run_loop. | result: success')

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
            return False
        total = 0
        received = 0

        try:
            count = client_sock._recv_int32() # Count of bets in batch
            while count > 0:
                total+= count

                logging.info(f"action: client_batch_recv_init | result: success | agency: {agency} | count_bets: {count}")

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

            self.awaiting_agencies[agency] = Agency(client_sock)
            logging.info(f"action: recv_agency_bets | result: success | agency: {agency} | count_bets: {total}")
            return True

        except Exception as e:
            logging.error(f"action: apuesta_recibida | result: fail | cantidad: {count}")
            logging.error(f"total bets recv {total} , batch recv {received} of {count} error: {e}")

            #Send response, possible IO error handled by invoker
            client_sock.send_int32(ERROR_CODE)
            client_sock.send_str(f"{e}")

            return False


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
