package server

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/sajadblnyn/grpc-intro-go/data"
	"github.com/sajadblnyn/grpc-intro-go/protos/currency/protos/currency"
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
				err = k.Send(&currency.RateResponse{Rate: ra, Base: currency.Currencies(currency.Currencies_value[r.GetBase().String()]), Destination: currency.Currencies(currency.Currencies_value[r.GetDestination().String()])})
				if err != nil {
					fmt.Println(err)
				}

			}
		}

	}
}

func (c *CurrencyServer) GetRate(cx context.Context, rq *currency.RateRequest) (*currency.RateResponse, error) {
	r, err := c.rates.GetRate(rq.GetBase().String(), rq.GetDestination().String())
	if err != nil {
		return nil, err
	}
	return &currency.RateResponse{Rate: r,
		Base:        currency.Currencies(currency.Currencies_value[rq.GetBase().String()]),
		Destination: currency.Currencies(currency.Currencies_value[rq.GetDestination().String()])}, nil
}

func (c *CurrencyServer) SubscribeRates(cs currency.Currency_SubscribeRatesServer) error {

	for {
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
		rrl = append(rrl, rr)

		c.subscriptions[cs] = rrl

		fmt.Printf("client sent this message:%v\n", rr.Base)
	}

	return nil

}
