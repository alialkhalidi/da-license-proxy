package main

import (
	"os"
	"os/signal"
	"syscall"
_	"fmt"

	"gmlserver"
  "log"
)

var (
	myLogger = log.New(log.Writer(),"gmlserver ",0)
  gmlseverconfig string
)

func main() {

	if len(os.Args) != 2 {
					myLogger.Fatal("config file not passed in")
	}

	gmlseverconfig = os.Args[1]

	var sigs = make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	gml, err := gmlserver.NewGmlServer(gmlseverconfig)
	if err != nil {
					myLogger.Fatalf("error configuring sim server: %v", err)
	}
	myLogger.Printf("******STARTING GML SERVER********")

	go func() {
					if err := gml.Start(); err != nil {
									myLogger.Printf("%s\n",err.Error())
									close(sigs)
					}
	}()
	s := <-sigs
	myLogger.Printf("Got shutdown signal: %v", s)
	if err := gml.Close(); err != nil {
					myLogger.Printf(err.Error())
	}

}
