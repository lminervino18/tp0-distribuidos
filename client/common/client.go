package common

import (
	"encoding/csv"
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

// readBetsFromCSV lee todas las apuestas del archivo CSV de la agencia
func (c *Client) readBetsFromCSV(csvPath string) ([]*Bet, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	bets := make([]*Bet, 0, len(rows))
	for _, row := range rows {
		bet := NewBet(c.config.ID, row[0], row[1], row[2], row[3], row[4])
		bets = append(bets, bet)
	}
	return bets, nil
}

// sendBatches envía todas las apuestas en batches al servidor
// Retorna false si se recibió SIGTERM durante el envío
func (c *Client) sendBatches(bets []*Bet, stopChan chan struct{}) bool {
	for i := 0; i < len(bets); i += c.config.BatchMaxAmount {
		select {
		case <-stopChan:
			return false
		default:
		}

		end := i + c.config.BatchMaxAmount
		if end > len(bets) {
			end = len(bets)
		}
		batch := bets[i:end]

		c.createClientSocket()
		if err := SendBatch(c.conn, batch); err != nil {
			log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v",
				c.config.ID, err)
			c.conn.Close()
			return false
		}

		confirmation, err := ReceiveBatchConfirmation(c.conn)
		c.conn.Close()
		if err != nil {
			log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v",
				c.config.ID, err)
			return false
		}

		log.Infof("action: apuesta_enviada | result: %v | cantidad: %v",
			confirmation, len(batch))
	}
	return true
}

// notifyFin notifica al servidor que terminamos de enviar apuestas
func (c *Client) notifyFin() error {
	c.createClientSocket()
	if err := SendFin(c.conn, c.config.ID); err != nil {
		log.Errorf("action: notify_fin | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		c.conn.Close()
		return err
	}
	c.conn.Close()
	return nil
}

// queryWinners consulta los ganadores al servidor y loguea el resultado
func (c *Client) queryWinners() error {
	c.createClientSocket()
	if err := SendQuery(c.conn, c.config.ID); err != nil {
		log.Errorf("action: consulta_ganadores | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		c.conn.Close()
		return err
	}

	winners, err := ReceiveWinners(c.conn)
	c.conn.Close()
	if err != nil {
		log.Errorf("action: consulta_ganadores | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return err
	}

	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v",
		len(winners))
	return nil
}

// StartClientLoop lee el CSV, envía apuestas en batches, notifica fin y consulta ganadores
func (c *Client) StartClientLoop(csvPath string) {
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

	bets, err := c.readBetsFromCSV(csvPath)
	if err != nil {
		log.Errorf("action: read_csv | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return
	}

	if !c.sendBatches(bets, stopChan) {
		return
	}

	if err := c.notifyFin(); err != nil {
		return
	}

	c.queryWinners()
}
