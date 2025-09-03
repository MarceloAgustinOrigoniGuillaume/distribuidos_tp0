package common

import (
	"sync"
	"context"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/serial"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

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


	conn, err := serial.NewClientConnection(c.config.ServerAddress)

	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}

	c.lock.Lock()
    defer c.lock.Unlock()	
	c.conn = conn
	return nil
}

func (c *Client) createClientReader() error {

	betReader, err := serial.NewBetReader(c.config.BatchSize, c.config.BetsCSV)

	if err != nil {
		log.Errorf("action: open bets file | result: fail | client_id: %v | error: %s",
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
			log.Infof("action: loop_cance at read first batch | result: success | client_id: %v", c.config.ID)
			return
		default: // Continue
		}	
		total:=int32(0)
		count, err := c.betReader.YieldBatch(packetBuilder)

		for count > 0 && c.checkContinue(ctx, "Read of batch", err) {
			log.Infof("action: loaded %d bets | result: success", count)
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
				log.Infof("action: failed finish send bets | total sent: %d", total)
				return
			}
		}

		log.Infof("action: finished sending bets | total sent: %d", total)

}


func (c *Client) StopClient() {
	c.lock.Lock()
    defer c.lock.Unlock()	

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			// Already closed?
		} else {
			log.Infof("client %v: connection closed", c.config.ID)
		}
	} else { // Should not really happen but just in case.
		log.Debugf("client %v: no connection to close", c.config.ID)
	}

	if c.betReader != nil {

		if err := c.betReader.Close(); err != nil {
			// Already closed?
		} else {
			log.Infof("client %v: reader closed", c.config.ID)
		}		
		
	}
}
