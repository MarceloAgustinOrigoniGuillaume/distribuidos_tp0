package common

import (
	"bufio"
	"fmt"
	"net"
	"time"
	"sync"
	"context"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int	
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
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
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.lock.Lock()
    defer c.lock.Unlock()	
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(ctx context.Context) {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		c.createClientSocket()
		select {
		case <-ctx.Done():
			c.StopClient(); // Stop again since we dont know for sure IF the close was to this socket.
			log.Infof("action: loop_cancel | result: success | client_id: %v", c.config.ID)
			return
		default: // Continue
		}
		// TODO: Modify the send to avoid short-write
		fmt.Fprintf(
			c.conn,
			"[CLIENT %v] Message N°%v\n",
			c.config.ID,
			msgID,
		)
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		c.conn.Close()

		if err != nil {
			select {
			case <-ctx.Done():
				log.Infof("action: loop_cancel | result: success | client_id: %v", c.config.ID)
			default:
				log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
					c.config.ID,
					err,
				)				
			}

			return
		}

		log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
			c.config.ID,
			msg,
		)

		// Wait a time between sending one message and the next one
		select {
		case <-ctx.Done():
			log.Infof("action: loop_cancel | result: success | client_id: %v", c.config.ID)
			return
		case <-time.After(c.config.LoopPeriod):
		}

	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
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
}
