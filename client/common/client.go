package common

import (
	"encoding/csv"
	"fmt"
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
		log.Errorf(
			"action: connect | result: fail | client_id: %v | error: %v",
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

// readBetsFromCSV lee todas las apuestas del archivo CSV de la agencia
func (c *Client) readBetsFromCSV(csvPath string) ([]*Bet, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Leer todas las filas del CSV
	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Convertir cada fila a una Bet
	bets := make([]*Bet, 0, len(rows))
	for _, row := range rows {
		bet := NewBet(c.config.ID, row[0], row[1], row[2], row[3], row[4])
		bets = append(bets, bet)
	}
	return bets, nil
}

// StartClientLoop lee el CSV y envía las apuestas en batches al servidor
func (c *Client) StartClientLoop(csvPath string) {
	// Canal para recibir señales del sistema operativo
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)

	stopChan := make(chan struct{})

	go func() {
		<-sigChan
		log.Infof("action: receive_sigterm | result: success | client_id: %v", c.config.ID)
		if c.conn != nil {
			c.conn.Close()
			log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
		}
		close(stopChan)
	}()

	// Leer todas las apuestas del CSV
	bets, err := c.readBetsFromCSV(csvPath)
	if err != nil {
		log.Errorf("action: read_csv | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return
	}

	// Enviar apuestas en batches
	for i := 0; i < len(bets); i += c.config.BatchMaxAmount {
		// Verificar si llegó SIGTERM antes de cada batch
		select {
		case <-stopChan:
			return
		default:
		}

		// Calcular el fin del batch actual
		end := i + c.config.BatchMaxAmount
		if end > len(bets) {
			end = len(bets)
		}
		batch := bets[i:end]

		// Conectar y enviar el batch
		c.createClientSocket()
		if err := SendBatch(c.conn, batch); err != nil {
			log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v",
				c.config.ID, err)
			c.conn.Close()
			return
		}

		// Esperar confirmación del servidor
		confirmation, err := ReceiveBatchConfirmation(c.conn)
		c.conn.Close()
		if err != nil {
			log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v",
				c.config.ID, err)
			return
		}

		log.Infof("action: apuesta_enviada | result: %v | cantidad: %v",
			confirmation, len(batch))
	}

	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
