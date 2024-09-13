package data

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"math/rand"
)

type ExchangeRates struct {
	rates map[string]float64
}

func NewRates() (*ExchangeRates, error) {
	er := &ExchangeRates{rates: map[string]float64{}}
	err := er.getRates()
	return er, err

}

func (e *ExchangeRates) MonitorRates(interval time.Duration) chan struct{} {
	ret := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ticker.C:
				// just add a random difference to the rate and return it
				// this simulates the fluctuations in currency rates

				for k, v := range e.rates {
					// change can be 10% of original value
					change := (rand.Float64() / 10)
					// is this a postive or negative change
					direction := rand.Intn(1)

					if direction == 0 {
						// new value with be min 90% of old
						change = 1 - change
					} else {
						// new value will be 110% of old
						change = 1 + change
					}

					// modify the rate
					e.rates[k] = v * change
				}
				ret <- struct{}{}

			}
		}
	}()
	return ret

}

func (e *ExchangeRates) GetRate(base, dest string) (float64, error) {
	br, ok := e.rates[base]
	if !ok {
		return 0, fmt.Errorf("there is no rate for base currency:%s", base)
	}
	dr, ok := e.rates[dest]
	if !ok {
		return 0, fmt.Errorf("there is no rate for destination currency:%s", dest)
	}

	return dr / br, nil
}

func (e *ExchangeRates) getRates() error {
	res, err := http.DefaultClient.Get("https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml")

	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed http status code:%d", res.StatusCode)
	}
	defer res.Body.Close()
	cubes := &Cubes{}
	err = xml.NewDecoder(res.Body).Decode(cubes)
	if err != nil {
		return err
	}

	for _, v := range cubes.CubeData {
		e.rates[v.Currency], err = strconv.ParseFloat(v.Rate, 64)
		if err != nil {
			return err
		}
	}

	return nil
}

type Cubes struct {
	CubeData []Cube `xml:"Cube>Cube>Cube"`
}

type Cube struct {
	Currency string `xml:"currency,attr"`
	Rate     string `xml:"rate,attr"`
}
