package common

import (
	"encoding/csv"
	"fmt"
	"io"
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
	ID             string
	ServerAddress  string
	LoopAmount     int
	LoopPeriod     time.Duration
	BatchMaxAmount int
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

// createClientSocket Initializes client socket, retrying until success or shutdown
func (c *Client) createClientSocket(stopChan <-chan struct{}) error {
	for {
		conn, err := net.Dial("tcp", c.config.ServerAddress)
		if err == nil {
			c.conn = conn
			return nil
		}
		log.Warningf(
			"action: connect | result: in_progress | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		select {
		case <-time.After(1 * time.Second):
		case <-stopChan:
			return fmt.Errorf("shutdown requested")
		}
	}
}

// sendBatches abre una única conexión, lee el CSV de a chunks y envía cada batch
// En ningún momento hay más de BatchMaxAmount apuestas en memoria
func (c *Client) sendBatches(csvPath string, stopChan <-chan struct{}) bool {
	// Abrir conexión persistente para toda la transmisión de batches
	if err := c.createClientSocket(stopChan); err != nil {
		log.Errorf("action: connect | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return false
	}
	defer c.conn.Close()

	file, err := os.Open(csvPath)
	if err != nil {
		log.Errorf("action: read_csv | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return false
	}
	defer file.Close()

	reader := csv.NewReader(file)
	batch := make([]*Bet, 0, c.config.BatchMaxAmount)

	for {
		// Verificar SIGTERM antes de cada fila
		select {
		case <-stopChan:
			return false
		default:
		}

		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("action: read_csv | result: fail | client_id: %v | error: %v",
				c.config.ID, err)
			return false
		}

		batch = append(batch, NewBet(c.config.ID, row[0], row[1], row[2], row[3], row[4]))

		// Cuando el batch está lleno, enviarlo
		if len(batch) == c.config.BatchMaxAmount {
			if !c.sendOneBatch(batch) {
				return false
			}
			batch = batch[:0]
		}
	}

	// Enviar el último batch parcial si quedaron apuestas
	if len(batch) > 0 {
		return c.sendOneBatch(batch)
	}
	return true
}

// sendOneBatch envía un batch por la conexión ya establecida y espera confirmación
func (c *Client) sendOneBatch(batch []*Bet) bool {
	if err := SendBatch(c.conn, batch); err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return false
	}

	confirmation, err := ReceiveBatchConfirmation(c.conn)
	if err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return false
	}

	log.Infof("action: apuesta_enviada | result: %v | cantidad: %v",
		confirmation, len(batch))
	return true
}

// StartClientLoop lee el CSV y envía las apuestas en batches al servidor
func (c *Client) StartClientLoop(csvPath string) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)

	stopChan := make(chan struct{})
	done := make(chan struct{})

	go func() {
		select {
		case <-sigChan:
			log.Infof("action: receive_sigterm | result: success | client_id: %v", c.config.ID)
			if c.conn != nil {
				c.conn.Close()
				log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
			}
			close(stopChan)
		case <-done:
			// loop terminó normalmente, salir
		}
	}()

	defer close(done)

	if !c.sendBatches(csvPath, stopChan) {
		return
	}

	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
