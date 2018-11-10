package main

import (
	"log"
	"time"

	"github.com/nikvas0/currency-rates"
)

func main() {
	usd2rub, ok := cb.GetRate("USD", time.Now())
	log.Println(usd2rub, ok)
	eur2rub, ok := cb.GetRate("EUR", time.Now())
	log.Println(eur2rub, ok)

	fake, ok := cb.GetRate("FAKE", time.Now())
	log.Println(fake, ok)
}
