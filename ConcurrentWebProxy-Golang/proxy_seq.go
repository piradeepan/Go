package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type UriParts struct {
	uri      string
	hostname string
	pathname string
	port     int
}

// Client is struct to hold socket and data channel
type Client struct {
	socket net.Conn
	data   chan []byte
}

var mutex = &sync.Mutex{}

func filterNewLines(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case 0x000A, 0x000B, 0x000C, 0x000D, 0x0085, 0x2028, 0x2029:
			return -1
		default:
			return r
		}
	}, s)
}

func (uri *UriParts) parseURI() bool {
	if !strings.HasPrefix(uri.uri, "http://") && !strings.HasPrefix(uri.uri, "https://") {
		uri.hostname = ""
		return false
	}
	var testsplit = strings.Split(uri.uri, "/")
	if strings.Contains(testsplit[2], ":") {
		var hostsplit = strings.Split(testsplit[2], ":")
		testsplit[2] = hostsplit[0]
		hostsplit[1] = string(bytes.Trim([]byte(filterNewLines(hostsplit[1])), "\x00"))
		port, err := strconv.Atoi(hostsplit[1])
		if err != nil {
			fmt.Println(err)
			return false
		}
		uri.port = port
	} else {
		if strings.HasPrefix(uri.uri, "https://") {
			uri.port = 443
		} else {
			uri.port = 80
		}
	}
	uri.hostname = testsplit[2]
	var paths = strings.SplitAfterN(uri.uri, "/", 4)
	if len(paths) >= 4 {
		uri.pathname = string(bytes.Trim([]byte(filterNewLines(paths[3])), "\x00"))
	} else {
		uri.pathname = ""
	}
	return true
}
func startServer(port string, of *os.File) {
	listener, error := net.Listen("tcp", ":"+port)
	if error != nil {
		fmt.Println(error)
		return
	}
	for {
		connection, error := listener.Accept()
		if error != nil {
			fmt.Println(error)
		} else {
			client := &Client{socket: connection, data: make(chan []byte)}
			go client.receive(of)
		}
	}
}
func (client *Client) receive(of *os.File) {
	for {
		message := make([]byte, 4096)
		length, err := client.socket.Read(message)
		if err != nil {
			client.socket.Close()
			break
		}
		if length > 0 {
			fmt.Println("RECEIVED: " + string(message))

			var method string
			var uri string
			var version string

			fmt.Sscanf(string(message), "%s %s %s", &method, &uri, &version)
			if method != "GET" {
				fmt.Println("Proxy does not implement the method")

			}
			uriparts := UriParts{}
			uriparts.uri = uri
			uriparts.parseURI()

			resp, err := http.Get(uriparts.uri)

			if err != nil {
				fmt.Println(err)
				break
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			client.socket.Write(body)

			getIPaddr := client.socket.LocalAddr().(*net.TCPAddr)

			t := time.Now()
			mutex.Lock()
			fmt.Fprintf(of, t.Format(time.RFC1123))
			fmt.Fprintf(of, ": %s %s %d\n", getIPaddr.IP, uri, len(body))
			mutex.Unlock()
			resp.Body.Close()
		}
	}
	fmt.Println("Socket Closed")
}

// GET http://www.bryancdixon.com:80/research/ HTTP/1.1
// GET http://www.aol.com:80 HTTP/1.1
// GET http://www.nfl.com:80/ HTTP/1.1

//Sun 27 Oct 2002 02:51:02 EST: 128.2.111.38 http://www.cs.cmu.edu/ 34314

func main() {
	argsWithoutProg := os.Args[1:]
	of, err := os.Create("proxy.log")
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(argsWithoutProg) >= 1 {
		startServer(argsWithoutProg[0], of)
	} else {
		startServer("1234", of)
	}
	of.Close()
}
