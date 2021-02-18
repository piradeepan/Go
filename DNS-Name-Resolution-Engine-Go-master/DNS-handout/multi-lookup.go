// Piradeepan Nagarajan
// CSCI640
// Assignment 3
// Submitted on 10/28/2020

package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

var MAX_INPUT_FILES int = 10
var MAX_RESOLVER_THREADS int = 10
var MIN_RESOLVER_THREADS int = 2
var MAX_NAME_LENGTH int = 1025
var MAX_IP_LENGTH int = 10

var mutex = &sync.Mutex{}
var wg sync.WaitGroup

func getHostname(c chan string, inputs []string) {
	defer wg.Done()
	defer close(c)
	for _, s := range inputs {
		infile, err := os.Open(s) // For read access.
		if err != nil {
			log.Fatal(err)
			return
		}
		defer infile.Close()
		// bufio - buffering IO is a technique used to temporarily accumulate the results for an IO operation before transmitting it forward
		scanner := bufio.NewScanner(infile)
		for scanner.Scan() {
			var hostname = scanner.Text()
			c <- hostname
		}
	}
}

func displayIPaddress(c chan string, of *os.File) {
	defer wg.Done()
	for hostResult := range c {
		ips, err := net.LookupIP(hostResult)
		if err != nil {
			mutex.Lock()
			fmt.Fprintf(os.Stderr, "dnslookup error: %s\n", hostResult)
			mutex.Unlock()
		}
		if len(ips) > 0 {
			mutex.Lock()
			fmt.Fprintf(of, "%s, %s\n", hostResult, ips[0].String())
			mutex.Unlock()
		} else {
			mutex.Lock()
			fmt.Fprintf(of, "%s \n", hostResult)
			mutex.Unlock()
		}
	}
}

func main() {
	argsWithoutProg := os.Args[1:]
	end := len(argsWithoutProg) - 1
	inputs := argsWithoutProg[:end]
	outfile := argsWithoutProg[end]

	if len(inputs) > MAX_INPUT_FILES {
		fmt.Println("Input files not allowed more than 10")
		return
	}

	of, err := os.Create(outfile)
	if err != nil {
		fmt.Println(err)
		return
	}

	channelValue := make(chan string, 10)

	wg.Add(3)
	go getHostname(channelValue, inputs)
	go displayIPaddress(channelValue, of)
	go displayIPaddress(channelValue, of)
	wg.Wait()
	of.Close()
}
