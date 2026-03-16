package common

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

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
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	return &Client{config: config}
}

// createClientSocket Initializes client socket retrying until success
func (c *Client) createClientSocket() error {
	for {
		conn, err := net.Dial("tcp", c.config.ServerAddress)
		if err == nil {
			c.conn = conn
			return nil
		}
		log.Errorf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		time.Sleep(1 * time.Second)
	}
}

// StartClientLoop sends a bet to the server and waits for confirmation
func (c *Client) StartClientLoop(bet *Bet) {
	// Canal para recibir señales del sistema operativo
	sigChan := make(chan os.Signal, 1)
	// Registrar SIGTERM en el canal para graceful shutdown
	signal.Notify(sigChan, syscall.SIGTERM)

	// Canal para notificar al loop principal que debe terminar
	stopChan := make(chan struct{})

	// Goroutine que espera SIGTERM en paralelo al loop principal
	go func() {
		<-sigChan
		log.Infof("action: receive_sigterm | result: success | client_id: %v", c.config.ID)
		if c.conn != nil {
			c.conn.Close()
			log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
		}
		close(stopChan)
	}()

	// Conectar al servidor y enviar la apuesta
	c.createClientSocket()

	if err := SendBet(c.conn, bet); err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		c.conn.Close()
		return
	}

	// Esperar confirmación del servidor
	confirmation, err := ReceiveConfirmation(c.conn)
	c.conn.Close()

	if err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return
	}

	log.Infof("action: apuesta_enviada | result: %v | dni: %v | numero: %v",
		confirmation, bet.Document, bet.Number)
}
