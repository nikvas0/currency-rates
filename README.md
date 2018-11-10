# currency-rates
Singleton for working with currency exchange rates.
Using api by [The Central Bank of the Russain Federation](http://www.cbr.ru/development/sxml/).

Example
```golang
usd2rub := cb.GetRate("USD", time.Now())
```
