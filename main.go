package main

import (
	"errors"
	"fmt"
	"httpfromtcp/internal/request"
	"io"
	"log"
	"net"

	//"os"
	"strings"
)

const inputFilePath = "messages.txt"

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("could not start listenter: %s", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Connection Accepted")

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
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
