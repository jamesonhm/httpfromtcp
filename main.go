package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const inputFilePath = "messages.txt"

func main() {
	f, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatalf("could not open %s: %s\n", inputFilePath, err)
	}

	linesCh := getLinesChannel(f)
	for line := range linesCh {
		fmt.Printf("read: %s\n", line)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)
	go func() {
		defer f.Close()
		defer close(ch)
		var currLine string
		for {
			b := make([]byte, 8)
			n, err := f.Read(b)
			if err != nil {
				if currLine != "" {
					ch <- currLine
					currLine = ""
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("error: %s\n", err.Error())
				break
			}
			str := string(b[:n])
			parts := strings.Split(str, "\n")
			for i := 0; i < len(parts)-1; i++ {
				ch <- fmt.Sprintf("%s%s", currLine, parts[i])
				currLine = ""
			}
			currLine += parts[len(parts)-1]
		}
	}()
	return ch
}
