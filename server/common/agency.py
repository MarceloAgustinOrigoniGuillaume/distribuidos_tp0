from .server_protocol import ServerProtocol 

import logging
import threading

from . import utils 
from queue import Queue
ALL_OK = 0
ERROR_CODE = 1
WINNERS_EOF = -1

class Agency:
    def __init__(self, connection: ServerProtocol):
        self.conn = connection
        self.winners = Queue()
        self.bets_received = threading.Event()
        self.agency_id = None


    def recv_bets(self, bets_store_lock):
        agency_id = ""
        count = 0
        try:
            agency_id = self.conn.recv_str()
            self.agency_id = int(agency_id)
        except Exception as e:
            logging.error(f"action: receive_bet_count | result: fail | error: {e}")
            return False
        total = 0
        received = 0

        try:
            count = self.conn._recv_int32() # Count of bets in batch
            while count > 0:
                total+= count

                logging.info(f"action: client_batch_recv_init | result: success | agency_id: {agency_id} | count_bets: {count}")

                res =[]
                received = 0
                for i in range(count):
                    bet = self.conn.recv_bet()
                    res.append(utils.Bet(agency_id, bet.first_name, bet.last_name, str(bet.document), bet.birthdate, str(bet.number)))
                    received+=1

                with bets_store_lock:
                    utils.store_bets(res)

                #logging.info(f"action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}")
                logging.info(f"action: apuesta_recibida | result: success | cantidad: {count}")
                self.conn.send_int32(ALL_OK) 
                count = self.conn._recv_int32() # Count of bets in batch

            self.bets_received.set()
            logging.info(f"action: recv_agency_bets | result: success | agency_id: {agency_id} | count_bets: {total}")
            return True

        except Exception as e:
            logging.error(f"action: apuesta_recibida | result: fail | cantidad: {count}")
            logging.error(f"total bets recv {total} , batch recv {received} of {count} error: {e}")

            #Send response, possible IO error handled by invoker
            self.conn.send_int32(ERROR_CODE)
            self.conn.send_str(f"{e}")

            return False

    def notify_winner(self, bet):
        self.winners.put(bet)
    def send_winner(self, bet):
        # None bet If server shutdown or winners EOF
        if bet != None:
            self.conn.send_int32(bet.number)
            self.conn.send_str(bet.document)
            return True
        
        self.conn.send_int32(WINNERS_EOF)
        return False

    def finished_winners(self):
        self.winners.put(None)


    def close(self):
        self.conn.close()
        self.winners.put(None)