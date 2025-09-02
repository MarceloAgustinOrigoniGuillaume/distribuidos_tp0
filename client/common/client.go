package common

import (
	"sync"
	"context"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
}





// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	protocol   *protocol.ClientProtocol
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


	prot, err := protocol.NewClientProtocol(c.config.ServerAddress)

	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}

	c.lock.Lock()
    defer c.lock.Unlock()	
	c.protocol = prot
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(ctx context.Context) {
			 
		c.createClientSocket()
		defer c.StopClient(); // Stop always since we dont really check/want to check wether it was already closed.
		
		select {
		case <-ctx.Done():
			log.Infof("action: loop_cancel | result: success | client_id: %v", c.config.ID)
			return
		default: // Continue
		}

		bet:= protocol.PersonBet {
			Name: "Some name",
			Surname: "Some surname",
			Dni: 324,
			Birth: "1999-03-17",
			Number: 213,
		}

		err:= c.protocol.SendStr(c.config.ID)
		if err == nil {
			err = c.protocol.SendBet(&bet)
		}
		
		if err != nil {
			select {
			case <-ctx.Done():
				log.Infof("action: bet_send_cancel | result: success | client_id: %v", c.config.ID)
			default:
				log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | %s | error: %s",
					c.config.ID,
					bet,					
					err,
				)				
			}

			return
		}

		log.Infof("action: apuesta_enviada | result: success | %s",
			bet.MainInfo(),
		)
}


func (c *Client) StopClient() {
	c.lock.Lock()
    defer c.lock.Unlock()	

	if c.protocol != nil {
		if err := c.protocol.Close(); err != nil {
			// Already closed?
		} else {
			log.Infof("client %v: connection closed", c.config.ID)
		}
	} else { // Should not really happen but just in case.
		log.Debugf("client %v: no connection to close", c.config.ID)
	}
}
