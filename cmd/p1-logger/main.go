package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path"

	"luit.eu/p1"
)

func main() {
	inputFile := flag.String("f", "", "input file")
	outputDir := flag.String("d", "/var/log/p1", "output directory")
	filePattern := flag.String("l", "2006-01-02.log", "output filename layout (time.Time.Format layout)")
	flag.Parse()

	input := os.Stdin
	if *inputFile != "" {
		f, err := os.Open(*inputFile)
		if err != nil {
			exit("error opening %s for reading: %v\n", *inputFile, err)
		}
		input = f
	}
	s := bufio.NewScanner(input)
	s.Split(p1.Split)
	for s.Scan() {
		t, err := p1.Parse(s.Bytes())
		if err != nil {
			logf("parse error: %v, payload: %q\n", err, s.Text())
		} else {
			d, ok := t.Data["0-0:1.0.0"]
			if !ok {
				logf("no date field in payload %q\n", s.Text())
				continue
			}
			tst, err := p1.DecodeTST(d)
			if err != nil {
				logf("invalid date: %v, payload: %q\n", err, s.Text())
				continue
			}
			fileName := tst.UTC().Format(*filePattern)
			fileName = path.Join(*outputDir, fileName)
			writeToFile(fileName, s.Bytes())
		}
	}
	if s.Err() != nil {
		exit("error scanning: %v\n", s.Err())
	}
}

var (
	currentFileName string
	currentFile     *os.File
)

func writeToFile(fileName string, data []byte) {
	if fileName != currentFileName || currentFile == nil {
		if currentFile != nil {
			if err := currentFile.Close(); err != nil {
				exit("error closing %s: %v\n", fileName, err)
			}
		}
		f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			exit("error opening %s for writing: %v\n", fileName, err)
		}
		currentFile = f
		currentFileName = fileName
	}
	n, err := currentFile.Write(data)
	if err != nil {
		exit("write error: %v\n", err)
	}
	if n != len(data) {
		exit("write was short (expected %d bytes, wrote %d)\n", len(data), n)
	}
}

func exit(format string, a ...interface{}) {
	logf(format, a...)
	os.Exit(1)
}

func logf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}
