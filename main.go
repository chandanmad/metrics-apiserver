package main

import (
	"fmt"
	"github.com/chandanmad/metrics-apiserver/server"
)

func main() {
	metricserver, err := server.NewMetricServer()
	if err != nil {
		panic(err)
	}
	fmt.Println("Starting metric server")
	if err := metricserver.Start(); err != nil {
		panic(err)
	}
}
