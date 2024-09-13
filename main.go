package main

import (
	"log"
	"net"

	"github.com/sajadblnyn/grpc-intro-go/data"
	"github.com/sajadblnyn/grpc-intro-go/protos/currency/protos/currency"
	"github.com/sajadblnyn/grpc-intro-go/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	rates, err := data.NewRates()

	if err != nil {
		log.Fatal(err)
	}

	gs := grpc.NewServer()
	cs := server.NewServer(rates)

	currency.RegisterCurrencyServer(gs, cs)

	reflection.Register(gs)
	l, err := net.Listen("tcp", ":9092")

	if err != nil {
		log.Fatal(err)
	}

	gs.Serve(l)
}
