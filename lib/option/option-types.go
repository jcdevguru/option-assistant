package option

// DTO objects

const (
	Call = iota
	Put
)

type Asset struct {
	Name       string
	Volatility float64
	RfReturn   float64
}

type Option struct {
	Asset  Asset
	Type   int
	Strike float64
	Expiry float64
}

type OptionPosition struct {
	Price        float64
	Strike       float64
	DaysToExpiry float64
}

// For Option Chain calculatins
type ValueSpan struct {
	Low  float64
	High float64
	Step float64
}

type d1d2Calculation struct {
	d1            float64
	d2            float64
	yearsToExpiry float64
}

type priceKey struct {
	assetPrice   float64
	strikePrice  float64
	daysToExpiry float64
}

type d1d2CalculateFunc func(assetPrice, strikePrice float64) (*d1d2Calculation, error)
type priceCalculatorFunc func(assetPrice, strikePrice, daysToExpiry float64, position *OptionPosition) error

type OptionChainCalculator struct {
	Volatility              float64
	RiskFreeRate            float64
	ExpiryInDays            float64
	calculatePrice          priceCalculatorFunc
	d1d2CalculateFuncMap    map[float64]d1d2CalculateFunc
	d1d2CalculationValueMap map[priceKey]d1d2Calculation
}

type OptionChain [][][]OptionPosition
