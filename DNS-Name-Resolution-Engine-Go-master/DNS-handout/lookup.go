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
	///Get the args without the executable
	argsWithoutProg := os.Args[1:]
	// fmt.Println(argsWithoutProg)
	// Output: input/names1.txt input/names2.txt input/names3.txt input/names4.txt input/names5.txt results.txt]

	//get the number of args excluding the trailing output file
	end := len(argsWithoutProg) - 1
	// fmt.Println(end)
	//Output: 5

	//get the slices of the args for the inputs and outputs
	inputs := argsWithoutProg[:end]

	// fmt.Println(inputs)
	// Output: [input/names1.txt input/names2.txt input/names3.txt input/names4.txt input/names5.txt]

	outfile := argsWithoutProg[end]
	fmt.Println(reflect.TypeOf(outfile))
	// fmt.Println(outfile)
	// Output: results.txt

	// For write access and to create the file if it doesn't exist
	of, err := os.Create(outfile)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(reflect.TypeOf(of))
	defer of.Close()
	//"inputs" has all the 5 txt files that we need to parse.
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
