package p1

import "time"

// DecodeTST decodes a COSEM value of type TST into time.Time.
func DecodeTST(data string) (time.Time, error) {
	switch data[len(data)-2] {
	case 'S':
		data = data[1:len(data)-2] + " +0200"
	case 'W':
		data = data[1:len(data)-2] + " +0100"
	default:
		return time.Time{}, parseError("expected S or W, got " + string([]byte{data[len(data)-1]}))
	}
	return time.Parse("060102150405 -0700", data)
}
