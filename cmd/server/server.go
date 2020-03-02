package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const serverDirName = "serverdata"

func init() {
	err := os.Mkdir(serverDirName, 0666)
	if err != nil {
		if !os.IsExist(err) {
			log.Fatalf("can't create directory: %s", err)
		}
	}
}

func main() {
	host := "0.0.0.0"
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "9998"
	}
	err := start(fmt.Sprintf("%v:%v", host, port)) // 0.0.0.0:9998
	if err != nil {
		log.Fatal(err)
	}
}

func start(addr string) (err error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("can't listen %s: %w", addr, err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		log.Print("accept connection")
		if err != nil {
			log.Printf("can't accept: %v", err)
			continue
		}
		log.Print("handle connection")
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	log.Print("read request to buffer")

	const maxHeaderSize = 4096
	reader := bufio.NewReaderSize(conn, maxHeaderSize)
	writer := bufio.NewWriter(conn)
	counter := 0
	buf := [maxHeaderSize]byte{}
	// naive header limit
	for {
		if counter == maxHeaderSize {
			log.Printf("too long request header")
			writer.WriteString("HTTP/1.1 413 Payload Too Large\r\n")
			writer.WriteString("Content-Length: 0\r\n")
			writer.WriteString("Connection: close\r\n")
			writer.WriteString("\r\n")
			writer.Flush()
			return
		}

		read, err := reader.ReadByte()
		if err != nil {
			log.Printf("can't read request line: %v", err)
			writer.WriteString("HTTP/1.1 400 Bad Request\r\n")
			writer.WriteString("Content-Length: 0\r\n")
			writer.WriteString("Connection: close\r\n")
			writer.WriteString("\r\n")
			writer.Flush()
			return
		}
		buf[counter] = read
		counter++

		if counter < 4 {
			continue
		}

		if string(buf[counter-4:counter]) == "\r\n\r\n" {
			break
		}
	}

	log.Print("headers found")
	headersStr := string(buf[:counter - 4])

		headers := make(map[string]string) // TODO: в оригинале map[string][]string
	headers["METHOD"] = "GET"



		requestHeaderParts := strings.Split(headersStr, "\r\n")
	if len(requestHeaderParts) < 2{

		return
	}
	log.Print("parse request line")
	requestLine := requestHeaderParts[0]
	log.Printf("request line: %s", requestLine)

	requestLineParts := strings.Split(requestLine, " ")
	query := requestLineParts[1]

	log.Print("send response")

	if query == "/" {
		filenames := list(conn)
		var length string
		if len(filenames) > 0 {
			length = strconv.Itoa(len(filenames))
		} else {
			filenames = "no files on server"
			length = "18"
		}
		writer.WriteString("HTTP/1.1 200 OK\r\n")
		writer.WriteString("Content-Length: ")
		writer.WriteString(length)
		writer.WriteString("\r\n")
		writer.WriteString("Connection: close\r\n")
		writer.WriteString("\r\n")
		writer.WriteString(filenames)
		writer.Flush()
	} else {
		query := query[1:]
		file, err := ioutil.ReadFile(serverDirName + "/" + query)
		if err != nil {
			log.Printf("can't read file %s: %s", query, err)
			writer.WriteString("HTTP/1.1 200 OK\r\n")
			writer.WriteString("Content-Length: 0\r\n")
			writer.WriteString("Connection: close\r\n")
			writer.WriteString("\r\n")
			writer.Flush()
		}
		length := strconv.Itoa(len(file))
		fileType := strings.ToLower(filepath.Ext(query))
		//fileType := strings.ToLower(filetype)
		var contentType string
		switch fileType {
		case ".txt":
			contentType = "text/plain"
		case ".pdf":
			contentType = "application/pdf"
		case ".png":
			contentType = "image/png"
		case ".jpg":
			contentType = "image/jpeg"
		case ".html":
			contentType = "text/html"
		}

		writer.WriteString("HTTP/1.1 200 OK\r\n")
		writer.WriteString("Content-Length: ")
		writer.WriteString(length)
		writer.WriteString("\r\n")
		writer.WriteString("Content-Type: ")
		writer.WriteString(contentType)
		writer.WriteString("\r\n")
		writer.WriteString("Connection: close\r\n")
		writer.WriteString("\r\n")
		writer.Write(file)
		writer.Flush()
	}



	log.Print("done")
	return
}

func list(conn net.Conn) string {
	dirents, err := ioutil.ReadDir(serverDirName)
	if err != nil {
		log.Printf("can't get dirents for path %s/: %s", serverDirName, err)
		return ""
	}

	filenames := make([]string, 0)
	for _, entry := range dirents {
		if !entry.IsDir() {
			filenames = append(filenames, entry.Name())
		}
	}

	filenamesStr := strings.Join(filenames, " ")

	return filenamesStr
}
