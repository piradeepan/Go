// Piradeepan Nagarajan, Subhed Chavan
// CSCI640
// Assignment 4
// Submitted on 11/15/2020

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

var wg sync.WaitGroup

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
func client(c chan net.Conn) {
	defer wg.Done()
	for conn := range c {
		message := make([]byte, 4096)
		length, err := conn.Read(message)
		if err != nil {
			conn.Close()
			break
		}
		if length > 0 {
			fmt.Println("RECEIVED: " + string(message))

			var method string
			var version string

			// The fmt.Sscanf() function scans the specified string and stores the successive separated values into arguments as determined by the format.
			fmt.Sscanf(string(message), "%s %s %s", &method, &uri, &version)
			if method != "GET" {
				fmt.Println("Proxy does not implement the method")
				return
			}
			// Creating struct without specifying field names
			// To accept the return value in varaible, I call uriparts
			uriparts := UriParts{}
			uriparts.uri = uri
			uriparts.parseURI()

			getContent, err := http.Get(uriparts.uri)
			if err != nil {
				fmt.Println(err)
				break
			}

			defer getContent.Body.Close()
			// ioutil.ReadAll - reading all data from a io.Reader until EOF. It’s often used to read data such as HTTP response body, files and other data sources which implement io.Reader interface
			body, err := ioutil.ReadAll(getContent.Body)
			conn.Write(body)

			getIPaddr := conn.LocalAddr().(*net.TCPAddr)

			t := time.Now()
			// This mutex will synchronize access to state.
			// A Mutex is used to provide a locking mechanism to ensure that only one Goroutine is running the critical section of code at any point of time to prevent race condition from happening.
			mutex.Lock()

			of, _ := os.OpenFile("proxy.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
			defer of.Close()

			fmt.Fprintf(of, t.Format(time.RFC1123))
			fmt.Fprintf(of, ": %s %s %d\n", getIPaddr.IP, uri, len(body))
			mutex.Unlock()
			getContent.Body.Close()
		}
	}
	fmt.Println("Socket Closed")
}
func startThread(c chan net.Conn, conn net.Conn) {
	defer wg.Done()
	defer close(c)
	for {
		c <- conn
	}
}
func main() {
	argsWithoutProg := os.Args[1:]
	port := argsWithoutProg[0]

	// The net.Listen() call is used for telling a Go program to accept network connections and thus act as a server. The return value of net.Listen() is of the net.Conn type, which implements the io.Reader and io.Writer interfaces.

	// The first parameter of the net.Listen() function defines the type of network that will be used, while the second parameter defines the server address as well as the port number the server will listen to.
	listener, error := net.Listen("tcp", ":"+port)
	if error != nil {
		fmt.Println(error)
		return
	}
	channelValue := make(chan net.Conn)
	for {
		// The for loop allows the program to keep accepting new TCP clients using Accept() that will be handled by instances of the startThread() function, which are executed as goroutines.
		connection, error := listener.Accept()
		if error != nil {
			fmt.Println(error)
		} else {
			// Add adds (2), which may be negative, to the WaitGroup counter. If the counter becomes zero, all goroutines blocked on Wait are released. If the counter goes negative, Add panics.
			wg.Add(2)
			go startThread(channelValue, connection)
			go client(channelValue)
			// Wait blocks until the WaitGroup counter is zero.
			wg.Wait()
		}
	}
}

// GET http://www.bryancdixon.com:80/research/ HTTP/1.1
// GET http://www.aol.com:80 HTTP/1.1
// GET http://www.nfl.com:80/ HTTP/1.1

//Sun 27 Oct 2002 02:51:02 EST: 128.2.111.38 http://www.cs.cmu.edu/ 34314
