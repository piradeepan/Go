package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
)

func main() {
	argsWithoutProg := os.Args[1:]
	end := len(argsWithoutProg) - 1
	inputs := argsWithoutProg[:end]
	outfile := argsWithoutProg[end]
	fmt.Println(reflect.TypeOf(outfile))
	of, err := os.Create(outfile)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(reflect.TypeOf(of))
	defer of.Close()
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
			ips, err := net.LookupIP(hostname)
			if err != nil {
				fmt.Fprintf(os.Stderr, "dnslookup error: %s\n", hostname)
			}
			if len(ips) > 0 {
				fmt.Fprintf(of, "%s %s\n", hostname, ips[0].String())
			} else {
				fmt.Fprintf(of, "%s \n", hostname)
			}
		}
	}
}
