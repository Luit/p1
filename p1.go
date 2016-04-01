// Package p1 implements DSMR P1 message reading, validation, and a bit of
// parsing.
package p1 // import "luit.eu/p1"
import (
	"bytes"
	"errors"
)

// Telegram represents a parsed DSMR P1 telegram.
type Telegram struct {
	Identifier string
	Data       map[string]string
}

// Split is a split function for a bufio.Scanner to scan for DSMR P1 messages.
// Any data before a '/' is emitted as a token. If the first byte is '/', the
// '!' + 6 bytes (expecting the CRC and '\r' + '\n')
func Split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	i := bytes.IndexByte(data, '/')
	switch i {
	case -1:
		// No '/' found, need more (or emit last chunk in buffer)
		if atEOF {
			advance = len(data)
			token = data
		}
	case 0:
		// First byte is '/', as expected. Look for '!' + 6 more
		// bytes.
		i = bytes.IndexByte(data, '!')
		if i != -1 && i+7 <= len(data) {
			token = data[:i+7]
			advance = i + 7
		} else if atEOF {
			advance = len(data)
			token = data
		}
	default:
		// A '/' was found beyond the first byte, emit everything up
		// to this byte.
		advance = i
		token = data[:i]
	}
	// We don't need empty tokens at clean EOF
	if len(token) == 0 {
		token = nil
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
	err = checkFormat(data)
	if err != nil {
		return
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
			err = parseError("expected CRC")
			return
		}
	}
	if crc != crc16(data[:len(data)-6]) {
		err = errCRC
		return
	}
	var n int
	t = &Telegram{}
	n, t.Identifier, err = parseIdentifier(data)
	if err != nil {
		return
	}
	t.Data, err = parseData(data[n : len(data)-7])
	return
}

// expects "/XXX5" prefix (ignores contents of first 5 bytes)
func parseIdentifier(data []byte) (int, string, error) {
	n := bytes.Index(data, []byte{'\r', '\n', '\r', '\n'})
	if n == -1 {
		return 0, "", parseError(`expected "\r\n\r\n"`)
	}
	return n + 4, string(data[5:n]), nil
}

func parseData(data []byte) (map[string]string, error) {
	var m map[string]string
	pos := 0
	for pos < len(data) {
		n := bytes.Index(data[pos:], []byte{'\r', '\n'})
		if n == -1 {
			return m, parseError(`expected "\r\n"`)
		}
		k, v, err := parseLine(data[pos : pos+n])
		if err != nil {
			return m, err
		}
		if m == nil {
			m = make(map[string]string)
		}
		m[k] = v
		pos += n + 2
	}
	return m, nil
}

func parseLine(data []byte) (k, v string, err error) {
	i := bytes.IndexByte(data, '(')
	if i == -1 {
		return "", "", parseError("expected '('")
	}
	return string(data[:i]), string(data[i:]), nil
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
		return parseError(`expected '\r'`)
	}
	if data[len(data)-1] != '\n' {
		return parseError(`expected '\n'`)
	}
	return nil
}
