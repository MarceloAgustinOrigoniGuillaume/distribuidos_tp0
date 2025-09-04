import socket
import logging
import threading
from .server_protocol import ServerProtocol 
from . import utils 
from queue import Queue
from concurrent.futures import ThreadPoolExecutor

class Server:
    def __init__(self, port, listen_backlog, agency_count):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self.agency_count = agency_count

        self.accepted_agencies = Queue()
        self.bets_store_lock = threading.Lock()

        # Using Event for running is thread safe, and allegedly idiomatic.
        self._shutdown_event = threading.Event()

    def should_run(self):
        # If not set shutdown then you should run, to save having another flag.
        return not self._shutdown_event.is_set() 

    def stop_accepter(self):
        self._server_socket.close()
        self._shutdown_event.set() 

    def stop(self):
        # Is not async as it is on go or other languages. No need for synchronization.
        logging.info(f'action: server_exiting_run_loop. | result: in_progress')
        self.stop_accepter()

    def __handle_client(self, agency):
        try:
            # Share/sync access to bets store.
            if agency.recv_bets(self.bets_store_lock):
                # Wait/consume winners
                bet = agency.winners.get()
                while agency.send_winner(bet):
                    bet = agency.winners.get()
        except OSError as e:
            if self.should_run():
                logging.error(f"action: client_handling | result: fail | error: {e}")


    def wait_for_lottery(self):

        new_agency = self.accepted_agencies.get() # Wait for an agency to finish... or shutdown signal If == None
        awaiting_agencies = {}

        while new_agency != None: # new_agency == None means should not run anymore.. == not self.should_run()            
            
            new_agency.bets_received.wait() # bets received could be triggered by accepter in case of server shutdown
            if not self.should_run():
                return

            # IF still running then agency already finished receiving bets. And has their ID
            awaiting_agencies[new_agency.agency_id] = new_agency

            if len(awaiting_agencies) < self.agency_count:
                new_agency= self.accepted_agencies.get() # Wait/Get next agency accepted
            else:
                # No need for sync since its assumed all bet receivers are closed
                winning_bets = filter(utils.has_won, utils.load_bets())

                # Since its not parallel it is not needed to group them by before sending them
                for bet in winning_bets:
                    awaiting_agencies[bet.agency].notify_winner(bet)
                
                for agency in awaiting_agencies.values():
                    agency.finished_winners()

                logging.info("action: sorteo | result: success")
                
                self._shutdown_event.set() # If not stopped already, stop it.
                return

    def run(self):

        executor = ThreadPoolExecutor(max_workers=self.agency_count+1) # All agencies + lottery/orchestrator thread
        executor.submit(self.wait_for_lottery)

        accepted_count = 0
        agencies = []
        while self.should_run() and accepted_count< self.agency_count: 
            # Enforce only up to the registered agencies, this saves the need to close the server socket from wait lottery

            try:
                agency = Agency(self.__accept_new_connection())
                accepted_count+=1
                agencies.append(agency)
                self.accepted_agencies.put(agency)
                executor.submit(self.__handle_client, agency)

            except OSError as e:
                if self.should_run():
                    logging.error(f"action: client accepter | result: fail | error: {e}")

        if self.should_run(): # If still running then close server socket, not needed any other agency.
            self._server_socket.close() 
            self._shutdown_event.wait() # Wait for the main thread to notify end of server
        else: # Forcing server shutdown while maybe not accepted all agencies.
            self.accepted_agencies.put(None)
        # Close agencies connections
        for agency in agencies:
            agency.close()

        executor.shutdown(wait=True)
        logging.info(f'action: server_exit | result: success | accepted: {accepted_count} agencies')


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
        return c
