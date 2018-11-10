package cb

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/text/encoding/charmap"
)

type Valute struct {
	CharCode string
	Value    string
}

type xmlRes struct {
	Valute []Valute
}

type singleton struct {
	mx   sync.RWMutex
	curs map[string]map[string]decimal.Decimal
}

//http://marcio.io/2015/07/singleton-pattern-in-go/
var instance *singleton
var once sync.Once

func GetInstance() *singleton {
	once.Do(func() {
		instance = &singleton{
			curs: make(map[string]map[string]decimal.Decimal),
		}
	})
	return instance
}

// Get currency exchange rate (to RUB)
// cur -- currency code ("USD", "EUR", ...)
// t -- time (time.Now(), ...)
func GetRate(cur string, t time.Time) (decimal.Decimal, bool) {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Println("can not load loc")
		panic(err)
	}
	key := t.In(loc).Format("02/01/2006")

	GetInstance().mx.RLock()
	res, ok := GetInstance().curs[key][cur]
	GetInstance().mx.RUnlock()

	if !ok {
		ReloadOnDate(t)

		GetInstance().mx.RLock()
		res, ok = GetInstance().curs[key][cur]
		GetInstance().mx.RUnlock()
	}
	return res, ok
}

// Reload currency rate from Сentral Bank api in specified time
func ReloadOnDate(t time.Time) {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Println("can not load loc")
		panic(err)
	}

	rates, err := updateRates("http://www.cbr.ru/scripts/XML_daily.asp?date_req=" + t.In(loc).Format("02/01/2006"))
	if err != nil {
		log.Println(err)
		return
	}

	key := t.In(loc).Format("02/01/2006")
	log.Println(key)

	GetInstance().mx.Lock()
	defer GetInstance().mx.Unlock()

	GetInstance().curs[key] = rates
	GetInstance().curs[key]["RUB"], _ = decimal.NewFromString("1")

	log.Println("new currency rates:")
	log.Println(GetInstance().curs)
}

// Reload last currency rate from Сentral Bank api
func ReloadLast() {
	rates, err := updateRates("http://www.cbr.ru/scripts/XML_daily.asp")
	if err != nil {
		log.Println(err)
		return
	}

	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Println("can not load loc")
		panic(err)
	}

	key := time.Now().In(loc).Format("02/01/2006")
	log.Println(key)

	GetInstance().mx.Lock()
	defer GetInstance().mx.Unlock()

	GetInstance().curs[key] = rates
	GetInstance().curs[key]["RUB"], _ = decimal.NewFromString("1")

	log.Println("new currency rates:")
	log.Println(GetInstance().curs)
}

// Using Сentral Bank api for updating rates
func updateRates(query string) (map[string]decimal.Decimal, error) {
	log.Println("Reloading currency rates")
	resp, err := http.Get(query)
	if err != nil {
		log.Println("Can not read currency")
		log.Println(err)
		return map[string]decimal.Decimal{}, err
	}
	defer resp.Body.Close()

	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		switch charset {
		case "windows-1251":
			return charmap.Windows1251.NewDecoder().Reader(input), nil
		default:
			return nil, fmt.Errorf("unknown charset: %s", charset)
		}
	}
	var res xmlRes
	err = decoder.Decode(&res)

	if err != nil {
		log.Println("Can not read xml from cbr resp")
		log.Println(err)
		return map[string]decimal.Decimal{}, err
	}

	rate := map[string]decimal.Decimal{}
	rate["RUB"], _ = decimal.NewFromString("1")

	for _, val := range res.Valute {
		val.Value = strings.Replace(val.Value, ",", ".", 1)
		rate[val.CharCode], err = decimal.NewFromString(val.Value)
		if err != nil {
			log.Println("Error while parsing cb float")
			return map[string]decimal.Decimal{}, err
		}
	}
	return rate, nil
}
