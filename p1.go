// Package p1 implements DSMR P1 message reading, validation, and a bit of
// parsing.
package p1 // import "luit.eu/p1"
import (
	"bytes"
	"errors"
	"fmt"
)

// Telegram represents a parsed DSMR P1 telegram.
type Telegram struct {
	Identifier []byte
}

// Split is a split function for a bufio.Scanner to scan for DSMR P1 messages.
// It splits by finding '!' and blindly ingesting another 6 characters after
// that (which should be the CRC and trailing "\r\n"). Split does not care
// about the start of the message either.
func Split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	i := bytes.IndexByte(data, '!')
	if i != -1 && i+7 <= len(data) {
		token = data[:i+7]
		advance = i + 7
	} else if atEOF && len(data) > 0 {
		advance = len(data)
		token = data
	}
	return
}

type parseError string

func (p parseError) Error() string {
	return "p1 parse error: " + string(p)
}

var (
	errCRC = errors.New("p1 CRC error")
)

// Parse validates the P1 message in data by checking its CRC, and parses it
// into a Telegram value.
func Parse(data []byte) (t *Telegram, err error) {
	if err := checkFormat(data); err != nil {
		return nil, err
	}
	crcData := data[len(data)-6 : len(data)-2]
	var crc uint16
	for _, b := range crcData {
		switch {
		case '0' <= b && b <= '9':
			crc = crc<<4 + uint16(b-'0')
		case 'A' <= b && b <= 'F':
			crc = crc<<4 + 10 + uint16(b-'A')
		default:
			return nil, parseError("expected CRC")
		}
	}
	if crc != crc16(data[:len(data)-6]) {
		fmt.Printf("crc mismatch: parsed %04X, calculated %04X\n", crc, crc16(data[:len(data)-6]))
		return nil, errCRC
	}
	return nil, nil
}

func checkFormat(data []byte) error {
	// minimal message: "/XXX5\r\n\r\n!ABCD\r\n" where ABCD = the CRC
	if len(data) < 16 {
		return parseError("message too short")
	}
	if data[0] != '/' {
		return parseError("expected '/'")
	}
	if data[4] != '5' {
		return parseError("expected '5'")
	}
	if data[len(data)-7] != '!' {
		return parseError("expected '!'")
	}
	if data[len(data)-2] != '\r' {
		return parseError("expected '\\r'")
	}
	if data[len(data)-1] != '\n' {
		return parseError("expected '\\n'")
	}
	return nil
}