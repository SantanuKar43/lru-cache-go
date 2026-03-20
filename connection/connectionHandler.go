package connection

import (
	"net"
	"bufio"
	"strings"
	"log"
	"github.com/SantanuKar43/lru-cache-go/cache"
)

func HandleConnection(conn net.Conn, cache *cache.Cache) {
	defer closeConnection(conn)
	reader := bufio.NewReader(conn)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Print("Error while reading message, closing connection", err)
			return
		}
		log.Printf("received message: %s", input)
		command := strings.Fields(input)
		switch command[0] {
		case "PING":
			// for testing connectivity
			conn.Write([]byte("PONG\n"))
		case "GET":
			// GET <key>
			if len(command) < 2 {
				conn.Write([]byte("INVALID_GET_COMMAND\n"))
				continue
			}
			key := command[1]
			val,err := cache.Get(key)
			if err != nil {
				conn.Write([]byte("GET_ERROR:" + err.Error() + "\n"))
				return
			}
			conn.Write([]byte(val + "\n"))
		case "PUT":
			// PUT <key> <value>
			if len(command) < 3 {
				conn.Write([]byte("INVALID_PUT_COMMAND\n"))
				continue
			}
			key := command[1]
			value := command[2]
			err := cache.Put(key, value)
			if err != nil {
				conn.Write([]byte("PUT_ERROR:" + err.Error() + "\n"))
				return
			}
			conn.Write([]byte("SUCCESS\n"))
		case "EXIT":
			conn.Write([]byte("BYE"))
			return
		default:
			conn.Write([]byte("UNKNOWN_COMMAND\n"))
		}
	}
}

func closeConnection(conn net.Conn) {
	log.Printf("closing connection: %s", conn.RemoteAddr())
	conn.Close()
}