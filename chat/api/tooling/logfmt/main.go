package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
)

// this program take the structured log output and makes it readable

var service string

func init() {
	fmt.Println("running init func")
	// 	flag.StringVar(&service, "service", "", "filter which service to see")

	//		shutdownChan := make(chan os.Signal, 1)
	//		signal.Notify(shutdownChan, syscall.SIGINT)
	//		// os.Kill.Signal()
	//		// <-shutdownChan
}
func main() {
	fmt.Println("running main func")

	flag.Parse()
	// service := strings.ToLower(service)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Println(string(scanner.Bytes()))
		if err := scanner.Err(); err != nil {
			log.Println(err)
		}
	}
}
