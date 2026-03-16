package common

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

const SEPARATOR = "|"

// Tipos de mensaje del protocolo
const (
	MSG_BATCH = uint8(0x01) // batch de apuestas
	MSG_FIN   = uint8(0x02) // notificación de fin de envío
	MSG_QUERY = uint8(0x03) // consulta de ganadores
)

// serialize convierte una Bet a bytes con el formato del protocolo:
// los campos separados por '|'
func serialize(bet *Bet) []byte {
	msg := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s",
		bet.Agency, SEPARATOR,
		bet.FirstName, SEPARATOR,
		bet.LastName, SEPARATOR,
		bet.Document, SEPARATOR,
		bet.Birthdate, SEPARATOR,
		bet.Number,
	)
	return []byte(msg)
}

// sendAll envía exactamente todos los bytes evitando short-write
func sendAll(conn net.Conn, data []byte) error {
	totalSent := 0
	for totalSent < len(data) {
		n, err := conn.Write(data[totalSent:])
		if err != nil {
			return err
		}
		totalSent += n
	}
	return nil
}

// recvAll lee exactamente n bytes evitando short-read
func recvAll(conn net.Conn, n int) ([]byte, error) {
	buf := make([]byte, n)
	totalRead := 0
	for totalRead < n {
		read, err := conn.Read(buf[totalRead:])
		if err != nil {
			return nil, err
		}
		totalRead += read
	}
	return buf, nil
}

// sendMessageType envía el byte de tipo de mensaje
func sendMessageType(conn net.Conn, msgType uint8) error {
	return sendAll(conn, []byte{msgType})
}

// SendBet serializa y envía una apuesta al servidor
// Formato: [2 bytes: largo][datos separados por '|']
func SendBet(conn net.Conn, bet *Bet) error {
	data := serialize(bet)

	// Enviar header con el largo en 2 bytes big-endian
	header := make([]byte, 2)
	binary.BigEndian.PutUint16(header, uint16(len(data)))
	if err := sendAll(conn, header); err != nil {
		return err
	}

	return sendAll(conn, data)
}

// SendBatch envía un batch de apuestas al servidor
// Formato: [1 byte: MSG_BATCH][2 bytes: cantidad][apuesta1]...[apuestaN]
func SendBatch(conn net.Conn, bets []*Bet) error {
	if err := sendMessageType(conn, MSG_BATCH); err != nil {
		return err
	}

	// Enviar cantidad de apuestas en 2 bytes big-endian
	header := make([]byte, 2)
	binary.BigEndian.PutUint16(header, uint16(len(bets)))
	if err := sendAll(conn, header); err != nil {
		return err
	}

	// Enviar cada apuesta individualmente reutilizando SendBet
	for _, bet := range bets {
		if err := SendBet(conn, bet); err != nil {
			return err
		}
	}
	return nil
}

// ReceiveBatchConfirmation lee la confirmación del servidor para un batch
// Formato: [2 bytes: largo][texto]
func ReceiveBatchConfirmation(conn net.Conn) (string, error) {
	header, err := recvAll(conn, 2)
	if err != nil {
		return "", err
	}
	length := int(binary.BigEndian.Uint16(header))
	data, err := recvAll(conn, length)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SendFin notifica al servidor que la agencia terminó de enviar apuestas
// Formato: [1 byte: MSG_FIN][2 bytes: agency_id]
func SendFin(conn net.Conn, agencyID string) error {
	if err := sendMessageType(conn, MSG_FIN); err != nil {
		return err
	}

	// Enviar agency_id en 2 bytes big-endian
	id := make([]byte, 2)
	agencyIDInt := uint16(0)
	fmt.Sscanf(agencyID, "%d", &agencyIDInt)
	binary.BigEndian.PutUint16(id, agencyIDInt)
	return sendAll(conn, id)
}

// SendQuery consulta al servidor los ganadores de la agencia
// Formato: [1 byte: MSG_QUERY][2 bytes: agency_id]
func SendQuery(conn net.Conn, agencyID string) error {
	if err := sendMessageType(conn, MSG_QUERY); err != nil {
		return err
	}

	// Enviar agency_id en 2 bytes big-endian
	id := make([]byte, 2)
	agencyIDInt := uint16(0)
	fmt.Sscanf(agencyID, "%d", &agencyIDInt)
	binary.BigEndian.PutUint16(id, agencyIDInt)
	return sendAll(conn, id)
}

// ReceiveWinners lee la lista de DNIs ganadores del servidor
// Formato: [2 bytes: largo total][DNI1|DNI2|...]
func ReceiveWinners(conn net.Conn) ([]string, error) {
	// Leer largo total del payload
	header, err := recvAll(conn, 2)
	if err != nil {
		return nil, err
	}
	length := int(binary.BigEndian.Uint16(header))

	// Si no hay ganadores retornar lista vacía
	if length == 0 {
		return []string{}, nil
	}

	// Leer exactamente los bytes indicados y splitear por '|'
	data, err := recvAll(conn, length)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(data), SEPARATOR), nil
}
