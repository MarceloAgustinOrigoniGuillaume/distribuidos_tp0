package common

import (
	"sync"
	"context"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/serial"

	"github.com/op/go-logging"
	"io"
	"time"
)

var log = logging.MustGetLogger("log")
const WINNERS_EOF = int32(-1)
const CONNECT_ATTEMPTS = 3
const RETRY_DELAY      = 1 * time.Second
type BetWinner struct {
	Dni string
	number int32
}

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	BetsCSV string
	BatchSize int32
}





// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   *serial.ClientConnection
	betReader *serial.BetReader
	lock     sync.Mutex // Needed since we dont know when the sigterm signal might come.	
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {

	for attempt := 1; attempt <= CONNECT_ATTEMPTS; attempt++ {
		conn, err = serial.NewClientConnection(c.config.ServerAddress)
		if err == nil {
			// Successful connection
			c.lock.Lock()
			defer c.lock.Unlock()
			c.conn = conn
			return nil
		}

		log.Infof(
			"action: connect_retry | result: in_progress | attempt: %d/%d | client_id: %v | error: %v",
			attempt, CONNECT_ATTEMPTS, c.config.ID, err,
		)

		time.Sleep(retryDelay)
	}

	log.Errorf(
		"action: connect | result: fail | client_id: %v | error: %v",
		c.config.ID,
		err,
	)
	
	return err
}

func (c *Client) createClientReader() error {

	betReader, err := serial.NewBetReader(c.config.BatchSize, c.config.BetsCSV)

	if err != nil {
		log.Errorf("action: open_bets_file | result: fail | client_id: %v | error: %s",
				c.config.ID,
				err,
			)
		return err
	}


	c.lock.Lock()
    defer c.lock.Unlock()	
	c.betReader = betReader

	return nil
}


func (c *Client) checkContinue(ctx context.Context, msg string, err error) bool {
	
	select {
	case <-ctx.Done():
		log.Infof("action: loop_cancel after %s | result: success | client_id: %v", msg, c.config.ID)
		return false
	default: // Continue
	}	

	if err != nil {
		log.Errorf("action: %s | result: fail | client_id: %v | error: %s",
			msg,
			c.config.ID,
			err,
		)	
		return false
	}

	return true
}

func (c *Client) initConnection() error {
	builder:= serial.NewClientSerializer()
	builder.WriteStr(c.config.ID)

	err := c.conn.Send(builder)
	if err != nil {
		log.Errorf("action: connection init | result: fail | client_id: %v | error: %s",
				c.config.ID,
				err,
			)
		return err
	}

	return err
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(ctx context.Context) {
			 
		if (c.createClientSocket() != nil) { // Abort client If connection failed.
			return
		}

		defer c.StopClient(); // Stop always since we dont really check/want to check wether it was already closed.

		if (c.initConnection() != nil || c.createClientReader() != nil) { // Abort client If reader or connection start failed
			return
		}

		packetBuilder:= serial.NewPacketBuilder()

		select {
		case <-ctx.Done():
			log.Infof("action: send_cancel_at_read_first_batch | result: success | client_id: %v", c.config.ID)
			return
		default: // Continue
		}	
		total:=int32(0)
		count, err := c.betReader.YieldBatch(packetBuilder)

		for count > 0 && c.checkContinue(ctx, "Read of batch", err) {
			log.Infof("action: load_%d_bets | result: success", count)
			// We need to check for err != nil after reading batch and after sending the packet.
			err = packetBuilder.SendPacket(count, c.conn)

			
			if c.checkContinue(ctx, "Send of batch", err) {
				total+= count
				log.Infof("action: send %d bets | result: success", count)
				count, err = c.betReader.YieldBatch(packetBuilder)
			} else {
				break // Do not print message as Read of Batch error.
			}
		}

		if count == 0 {
			err= packetBuilder.SendFinish(c.conn)

			if !c.checkContinue(ctx, "Finish send", err){
				log.Infof("action: failed_finish_send_bets | result: success | total_sent: %d", total)
				return
			}

			// Can be a winning number, or -1 If no more winners.
			num,  errNum := c.conn.ReadInt()
			winners := make([]BetWinner, 0,5)
			i:=0
			// If server is not interrupted it should not happend that io.EOF is received here.
			for errNum != io.EOF && c.checkContinue(ctx, "Receive winner number", errNum) && num != WINNERS_EOF{
				dni, errDni := c.conn.ReadStr()
				if (!c.checkContinue(ctx, "Receive winner dni", errDni)){
					return
				}
				winners = append(winners, BetWinner {
					Dni: dni,
					number: num,
				})
				i+=1
			    log.Infof("action: winner_recv | result: success | Winner: %d | dni: %s",i, dni)

				num, errNum = c.conn.ReadInt()
			}
			

			log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d", len(winners))

		} else{
			log.Infof("action: exit | result: success | total_sent: %d", total)
		}

}


func (c *Client) StopClient() {
	c.lock.Lock()
    defer c.lock.Unlock()	

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			// Already closed?
		} else {
			log.Infof("action: close_client | result: success | client_id: %v", c.config.ID)
		}
	} else { // Should not really happen but just in case.
		log.Debugf("client %v: no connection to close", c.config.ID)
	}

	if c.betReader != nil {

		if err := c.betReader.Close(); err != nil {
			// Already closed?
		} else {
			log.Infof("action: close_reader | result: success | client_id: %v", c.config.ID)
		}		
		
	}
}
