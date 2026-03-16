package common

import (
	"encoding/binary"
	"fmt"
	"net"
)

const SEPARATOR = "|"

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

// SendBet serializa y envía una apuesta al servidor
// Formato: [2 bytes: largo][datos]
func SendBet(conn net.Conn, bet *Bet) error {
	data := serialize(bet)

	// Enviar header con el largo en 2 bytes big-endian
	header := make([]byte, 2)
	binary.BigEndian.PutUint16(header, uint16(len(data)))
	if err := sendAll(conn, header); err != nil {
		return err
	}

	// Enviar los datos
	return sendAll(conn, data)
}

// ReceiveConfirmation lee la confirmación del servidor
// Formato: [2 bytes: largo][texto]
func ReceiveConfirmation(conn net.Conn) (string, error) {
	// Leer header con el largo
	header, err := recvAll(conn, 2)
	if err != nil {
		return "", err
	}

	// Leer exactamente los bytes indicados en el header
	length := int(binary.BigEndian.Uint16(header))
	data, err := recvAll(conn, length)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
