package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

var wg sync.WaitGroup

type lockable struct {
	mu sync.Mutex
}

func receiver(c chan string, inputs []string) {
	defer wg.Done()
	for _, s := range inputs {
		fmt.Println("Processing File: ", s)
		infile, err := os.Open(s) // For read access.
		if err != nil {
			log.Fatal(err)
			fmt.Println("Error")
			return
		}
		defer infile.Close()
		scanner := bufio.NewScanner(infile)
		for scanner.Scan() {
			var hostname = scanner.Text()
			c <- hostname
		}

	}
	close(c)
}

func IsIpv4Net(host string) bool {
	return net.ParseIP(host) != nil
}

func (d *lockable) resolver(c chan string, of *os.File) {
	defer wg.Done()
	for hostname := range c {
		ips, err := net.LookupIP(hostname)
		if err != nil {
			d.mu.Lock()
			fmt.Fprintf(os.Stderr, "dnslookup error: %s\n", hostname)
			d.mu.Unlock()
		}
		if len(ips) > 0 {
			for _, num := range ips {
				if num.To4() == nil {
					d.mu.Lock()
					fmt.Fprintf(of, "%s %s\n", hostname, num.String())
					d.mu.Unlock()
					break
				}
			}
		} else {
			d.mu.Lock()
			fmt.Fprintf(of, "%s \n", hostname)
			d.mu.Unlock()
		}
	}
}

func main() {
	//Get the args without the executable
	argsWithoutProg := os.Args[1:]
	//get the number of args excluding the trailing output file
	end := len(argsWithoutProg) - 1

	//get the slices of the args for the inputs and outputs
	inputs := argsWithoutProg[:end]
	outfile := argsWithoutProg[end]
	d := lockable{}
	// For write access and to create the file if it doesn't exist
	of, err := os.Create(outfile)
	if err != nil {
		fmt.Println(err)
		return
	}

	fool := make(chan string)
	wg.Add(2)
	go receiver(fool, inputs)
	go d.resolver(fool, of)
	time.Sleep(time.Second)
	wg.Wait()

}
