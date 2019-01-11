package main

import (
	"flag"
	"fmt"
	"log"

	. "github.com/serjk/go-pacemaker"
	"github.com/serjk/go-pacemaker/impl"
)

var f_verbose = flag.Bool("verbose", false, "print whole cib on each update")
var f_file = flag.String("file", "", "file to load as CIB")
var f_remote = flag.String("remote", "", "remote server to connect to (ip)")
var f_port = flag.Int("port", 3121, "remote port to connect to (3121)")
var f_user = flag.String("user", "hacluster", "remote user to connect as")
var f_password = flag.String("password", "", "remote password to connect with")
var f_encrypted = flag.Bool("encrypted", false, "set if remote connection is encrypted")

func listenToCib(c CibClient, restarter chan int) {
	_, err := c.Subscribe(func(event CibEvent, doc *CibDocument) {
		if event == UpdateEvent {
			fmt.Printf("\n")
			fmt.Printf("event: %s\n", event)
			if *f_verbose {
				fmt.Printf("cib: %s\n", string(doc.Xml()))
			}
		} else {
			log.Printf("lost connection: %s\n", event)
			restarter <- 1
		}
	})
	if err != nil {
		log.Printf("Failed to subscribe to CIB: %s", err)
	}
}

func connectToCib() (CibClient, error) {
	var c CibClient
	var err error
	if *f_file != "" {
		c, err = impl.NewCibClientImpl(impl.FromFile(*f_file))
	} else if *f_remote != "" {
		c, err = impl.NewCibClientImpl(impl.FromRemote(*f_remote, *f_user, *f_password, *f_port, *f_encrypted))
	} else {
		c, err = impl.NewCibClientImpl(impl.ForCommand)
	}
	if err != nil {
		log.Print("Failed to open CIB")
		return nil, err
	}

	err = c.Connect()
	if err != nil {
		log.Print("Failed connection to CIB")
		return nil, err
	}

	doc, err := c.Query()
	if err != nil {
		log.Print("Failed to query CIB")
		return nil, err
	}
	if *f_verbose {
		fmt.Printf("CIB: %s\n", string(doc.Xml()))
	}
	return c, nil
}

func main() {
	flag.Parse()

	cib, err := connectToCib()
	if err != nil {
		log.Fatal(err)
	}
	info, err := cib.Query()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		//return
	}
	fmt.Printf("%s\n", string(info.Xml()))
}
