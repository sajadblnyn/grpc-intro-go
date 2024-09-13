package data

import (
	"fmt"
	"testing"
)

func TestNewRates(t *testing.T) {
	er, err := NewRates()

	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("rates data : %v\n", er.rates)
}
