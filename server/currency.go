package server

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/sajadblnyn/grpc-intro-go/data"
	"github.com/sajadblnyn/grpc-intro-go/protos/currency/protos/currency"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CurrencyServer struct {
	rates *data.ExchangeRates

	subscriptions map[currency.Currency_SubscribeRatesServer][]*currency.RateRequest
}

func NewServer(rates *data.ExchangeRates) *CurrencyServer {
	cs := &CurrencyServer{rates: rates, subscriptions: make(map[currency.Currency_SubscribeRatesServer][]*currency.RateRequest)}
	go cs.handleUpdateCurrencies()

	return cs

}

func (c *CurrencyServer) handleUpdateCurrencies() {

	ru := c.rates.MonitorRates(3 * time.Second)

	for range ru {
		fmt.Println("Got Updated rates")
		for k, v := range c.subscriptions {
			for _, r := range v {
				ra, err := c.rates.GetRate(r.GetBase().String(), r.GetDestination().String())
				if err != nil {
					fmt.Println(err)
				}
				err = k.Send(&currency.StreamingRateResponse{
					Message: &currency.StreamingRateResponse_RateResponse{
						RateResponse: &currency.RateResponse{
							Rate:        ra,
							Base:        currency.Currencies(currency.Currencies_value[r.GetBase().String()]),
							Destination: currency.Currencies(currency.Currencies_value[r.GetDestination().String()])}}})

				if err != nil {
					fmt.Println(err)
				}

			}
		}

	}
}

func (c *CurrencyServer) GetRate(cx context.Context, rq *currency.RateRequest) (*currency.RateResponse, error) {
	if rq.Base.String() == rq.Destination.String() {
		err := status.Newf(codes.InvalidArgument, "base and destination currency are the same base:%s dest:%s", rq.Base.String(), rq.Destination.String())
		err, wd := err.WithDetails(rq)
		if wd != nil {
			return nil, wd
		}
		return nil, err.Err()
	}

	r, err := c.rates.GetRate(rq.GetBase().String(), rq.GetDestination().String())
	if err != nil {
		return nil, err
	}
	return &currency.RateResponse{Rate: r,
		Base:        currency.Currencies(currency.Currencies_value[rq.GetBase().String()]),
		Destination: currency.Currencies(currency.Currencies_value[rq.GetDestination().String()])}, nil
}

func (c *CurrencyServer) SubscribeRates(cs currency.Currency_SubscribeRatesServer) error {
	var requestAlreadyExistsForClient bool = false
	for {
		requestAlreadyExistsForClient = false
		rr, err := cs.Recv()
		if err == io.EOF {
			fmt.Println("connection has been closed by client", err)
			return err
		}
		if err != nil {
			fmt.Println(err)
			return err
		}
		rrl, ok := c.subscriptions[cs]

		if !ok {

			rrl = []*currency.RateRequest{}
		}
		for _, v := range rrl {

			if v.Base == rr.Base && v.Destination == rr.Destination {
				requestAlreadyExistsForClient = true
				grpcErr := status.New(codes.AlreadyExists, "this currency rate request is already exists for you")
				grpcErr, err = grpcErr.WithDetails(rr)
				if err != nil {
					fmt.Println("error in set details for grpc error")
					break
				}
				cs.Send(&currency.StreamingRateResponse{Message: &currency.StreamingRateResponse_Error{Error: grpcErr.Proto()}})
				break

			}

		}
		if !requestAlreadyExistsForClient {
			rrl = append(rrl, rr)

			c.subscriptions[cs] = rrl
			fmt.Printf("client sent this request base:%s destination:%s", rr.Base, rr.Destination)

		}

	}

	return nil

}
