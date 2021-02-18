// Authors: Subhed Chavan, Piradeepan Nagarajan
// Class: CSCI-640 Operating Systems
// Assignment 3

package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"sync"
)

var MAX_INPUT_FILES int = 10
var MAX_RESOLVER_THREADS int = runtime.NumCPU()
var MIN_RESOLVER_THREADS int = 2
var MAX_NAME_LENGTH int = 1025
var MAX_IP_LENGTH int = 10

var mutex = &sync.Mutex{}
var wg sync.WaitGroup

func getHost(c chan string, inputs []string) {

	defer wg.Done()
	defer close(c)

	for _, s := range inputs {

		infile, err := os.Open(s) // For read access.
		if err != nil {
			log.Fatal(err)
			return
		}

		defer infile.Close()
		scanner := bufio.NewScanner(infile)
		for scanner.Scan() {
			var hostname = scanner.Text()
			c <- hostname
		}

	}
}

func getIP(c chan string, of *os.File) {

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
			fmt.Fprintf(of, "%s", hostResult)
			for _, ip := range ips {
				fmt.Fprintf(of, ", %s", ip)
			}
			fmt.Fprintf(of, "\n")
			mutex.Unlock()
		} else {
			mutex.Lock()
			fmt.Fprintf(of, "%s \n", hostResult)
			mutex.Unlock()
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

	// Debugging Input
	fmt.Println("Arguments:", argsWithoutProg, "\n")
	fmt.Println("End Index:", end, "\n")
	fmt.Println("Inputs:", inputs, "\n")
	fmt.Println("Outputs:", outfile, "\n")

	if len(inputs) > MAX_INPUT_FILES {
		fmt.Println("Input files not allowed more than 10")
		return
	}

	channel := make(chan string, 10)

	of, err := os.Create(outfile)
	if err != nil {
		fmt.Println(err)
		return
	}

	wg.Add(MAX_RESOLVER_THREADS + 1)
	go getHost(channel, inputs)
	for i := 0; i < MAX_RESOLVER_THREADS && i < 10; i++ {
		go getIP(channel, of)
	}

	wg.Wait()
	of.Close()

}
