package util

import (
	"github.com/shopspring/decimal"
)

func Round(value float64, places int32) float64 {
	val := decimal.NewFromFloat(value)

	return val.Round(places).InexactFloat64()
}
