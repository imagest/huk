package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

type HukServer struct {
	fileName string
	// TODO extend this to serve more than one file
	// fileList map[string][string]

	// TODO add a list of currently-ongoing-goroutines that're serving?
}

func Run(port, fileName string) {
	log.Println("Start server on Port ", port, "...")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Println("error listening to port "+port, err)
		return
	}
	connections := makeChannels(listener, fileName)
	for {
		go serveInChunk(<-connections, fileName)
		//go serveInBlock(<-connections, fileName)
	}
}

func makeChannels(listener net.Listener, fileName string) chan net.Conn {
	channel := make(chan net.Conn)
	// perpetually run this concurrently
	go func() {
		for {
			conn, err := serveHandshake(listener, fileName)
			if conn == nil {
				return
			}
			if err != nil {
				log.Println("error accepting connection", err)
				return
			}
			channel <- conn
		}
	}()
	return channel
}

func serveInChunk(conn net.Conn, fileName string) {
	log.Println("serve in chunk")
	if conn == nil {
		conn.Close()
		return
	}
	file, err := os.Open(fileName)
	file, err = encryptFile(file, "")
	if err != nil {
		log.Println("error opening "+fileName, err)
		return
	}
	defer file.Close()

	numSent, err := io.Copy(conn, file)
	if err != nil {
		log.Println(err)
	}
	log.Println(numSent, " sent to ", conn.LocalAddr().String())

	// finish with this client
	conn.Close()
}

func encryptFile(file *os.File, key string) (*os.File, error) {
	return file, nil
}

//func encryptBlock(reader *bufio.Reader, key string) (*bufio.Reader, error) {
//	return reader, nil
//}

func serveInBlock(conn net.Conn, fileName string) {
	file, err := os.Open(fileName)
	file, err = encryptFile(file, "")
	if err != nil {
		log.Println("error opening "+fileName, err)
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	outBuffer := make([]byte, 2048)
	for {
		// read a chunk
		numRead, err := reader.Read(outBuffer)
		if err != nil {
			log.Println("problem with reader")
			log.Println(numRead, err)
			break
		}
		// write that chunk to outgoing request
		numSent, err := conn.Write(outBuffer[0:numRead])
		log.Println(numRead, "bytes read", numSent, "bytes sent")
	}

	conn.Close()
}

func serveHandshake(listener net.Listener, fileName string) (net.Conn, error) {
	var message string
	conn, err := listener.Accept()

	if err != nil {
		return nil, err
	}

	userName, _ := bufio.NewReader(conn).ReadString('\n')
	if trustConnection(userName) {
		message = fileName + "\n"
		conn.Write([]byte(message))

		return conn, err
	}
	conn.Close()
	return nil, nil

}

func trustConnection(username string) bool {
	var trust rune

	fmt.Printf("%v wants to get your file, do you trust them [Y/n]?:", username)

	for trust != 'y' && trust != 'Y' && trust != 'n' && trust != 'N' {
		fmt.Scanf("%c", &trust)
	}

	if trust == 'y' || trust == 'Y' {
		return true
	}
	fmt.Printf("Whew, that was a close one, goodbye!\n")
	return false
}
