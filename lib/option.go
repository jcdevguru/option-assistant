package lib

import (
	"fmt"
	"math"
	"time"
)

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

type Position struct {
	Cost float64
	Date time.Time
}

type AssetPosition struct {
	Asset
	Position
}

type OptionPosition struct {
	Option
	Position
}

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
type priceCalculatorFunc func(assetPrice, strikePrice, daysToExpiry float64) (float64, error)

// Standard normal cumulative distribution function
func NormCDF(x float64) float64 {
	return 0.5 * (1.0 + math.Erf(x/math.Sqrt(2.0)))
}

func validateSpan(name string, span *ValueSpan) (float64, float64, float64, error) {
	var err error
	if span.Low < 0 || span.High < 0 {
		err = fmt.Errorf("%s: range Low, High cannot be negative", name)
	} else if span.Low > span.High {
		err = fmt.Errorf("%s: Low > High", name)
	} else if span.Step <= 0 {
		err = fmt.Errorf("%s: Step == 0.0", name)
	} else if span.Step >= span.High-span.Low {
		err = fmt.Errorf("%s: Step >= High - Low", name)
	} else {
		for _, v := range []float64{span.Low, span.High, span.Step} {
			cVal := v * 4.0
			if math.Ceil(cVal) != math.Floor(cVal) {
				err = fmt.Errorf("%s: Low, High, Step values must be multiples of 0.25", name)
				break
			}
		}
	}

	if err != nil {
		return 0.0, 0.0, 0.0, err
	}

	return span.Low, span.High, span.Step, nil
}

type OptionChain struct {
	CalculatePrice          priceCalculatorFunc
	Volatility              float64
	RiskFreeRate            float64
	ExpiryInDays            float64
	d1d2CalculateFuncMap    map[float64]d1d2CalculateFunc
	d1d2CalculationValueMap map[priceKey]d1d2Calculation
}

func NewOptionChain(optionType int, volatility, riskFreeRate, expiryInDays float64) (*OptionChain, error) {
	chain := OptionChain{
		Volatility:   volatility,
		RiskFreeRate: riskFreeRate,
		ExpiryInDays: expiryInDays,
	}

	var priceCalculator priceCalculatorFunc
	switch optionType {
	case Call:
		priceCalculator = chain.BlackScholesCall
	case Put:
		priceCalculator = chain.BlackScholesPut
	default:
		return nil, fmt.Errorf("unrecognized optionType %d", optionType)
	}
	chain.CalculatePrice = priceCalculator
	chain.d1d2CalculateFuncMap = make(map[float64]d1d2CalculateFunc)
	chain.d1d2CalculationValueMap = make(map[priceKey]d1d2Calculation)

	return &chain, nil
}

// Black-Scholes formula for call option price

func (chain *OptionChain) d1d2calculator(daysToExpiry float64) (d1d2CalculateFunc, error) {
	yearsToExpiry := daysToExpiry / 365
	sqrtT := math.Sqrt(yearsToExpiry)
	maxReturn := (chain.RiskFreeRate + (chain.Volatility*chain.Volatility)/2.0) * yearsToExpiry
	volatilityAdjustment := chain.Volatility * sqrtT
	if volatilityAdjustment == 0.0 {
		return nil, fmt.Errorf(
			"volatilityAdjustment == 0.0, op = r/v/d/y = %v/%v/%v/%v",
			chain.RiskFreeRate, chain.Volatility, daysToExpiry, yearsToExpiry,
		)
	}

	return func(assetPrice float64, strikePrice float64) (*d1d2Calculation, error) {
		d1 := (math.Log(assetPrice/strikePrice) + maxReturn) / volatilityAdjustment
		d2 := d1 - volatilityAdjustment
		if math.IsNaN(d1) || math.IsNaN(d2) {
			return nil, fmt.Errorf(
				"d1 or d2 == NaN, op = d1/d2/a/s/va = %v/%v/%v/%v/%v",
				d1, d2, assetPrice, strikePrice, volatilityAdjustment,
			)
		}
		return &d1d2Calculation{d1, d2, yearsToExpiry}, nil
	}, nil
}

func (chain *OptionChain) calculateD1D2(assetPrice, strikePrice, daysToExpiry float64) (*d1d2Calculation, error) {
	var d1d2 d1d2Calculation
	valueMap := chain.d1d2CalculationValueMap
	key := priceKey{assetPrice, strikePrice, daysToExpiry}
	d1d2, ok := valueMap[key]
	if !ok {
		var err error
		funcMap := chain.d1d2CalculateFuncMap
		calculate := funcMap[daysToExpiry]
		if calculate == nil {
			calculate, err = chain.d1d2calculator(daysToExpiry)
			if err != nil {
				return nil, err
			}
			funcMap[daysToExpiry] = calculate
		}
		calcP, err := calculate(assetPrice, strikePrice)
		if err != nil {
			return nil, err
		}
		d1d2 = *calcP
		valueMap[key] = d1d2
	}
	return &d1d2, nil
}

func (chain *OptionChain) BlackScholesCall(assetPrice, strikePrice, daysToExpiry float64) (float64, error) {
	d1d2, err := chain.calculateD1D2(assetPrice, strikePrice, daysToExpiry)
	if err != nil {
		return 0.0, err
	}
	rc := assetPrice*NormCDF(d1d2.d1) - strikePrice*math.Exp(-chain.RiskFreeRate*d1d2.yearsToExpiry)*NormCDF(d1d2.d2)
	return rc, nil
}

// Black-Scholes formula for put option price
func (chain *OptionChain) BlackScholesPut(assetPrice, strikePrice, daysToExpiry float64) (float64, error) {
	d1d2, err := chain.calculateD1D2(assetPrice, strikePrice, daysToExpiry)
	if err != nil {
		return 0.0, err
	}
	rc := strikePrice*math.Exp(-chain.RiskFreeRate*d1d2.yearsToExpiry)*NormCDF(-d1d2.d2) - assetPrice*NormCDF(-d1d2.d1)
	return rc, nil
}

func (chain *OptionChain) ComputeOptionChain(assetPriceSpan, strikePriceSpan, daysToExpirySpan *ValueSpan) ([][][]float64, error) {
	apLow, apHigh, apStep, err := validateSpan("assetPriceRange", assetPriceSpan)
	if err != nil {
		return nil, err
	}

	spLow, spHigh, spStep, err := validateSpan("strikePriceRange", strikePriceSpan)
	if err != nil {
		return nil, err
	}

	dteLow, dteHigh, dteStep, err := validateSpan("strikePriceRange", strikePriceSpan)
	if err != nil {
		return nil, err
	}
	if dteLow == 0.0 {
		dteLow += dteStep
	}

	var strikeExpiryPricesPerAssetPrice [][][]float64
	for assetPrice := apLow; assetPrice <= apHigh; assetPrice += apStep {
		var expiryPricesPerStrike [][]float64
		for strikePrice := spLow; strikePrice <= spHigh; strikePrice += spStep {
			var pricesPerExpiry []float64
			for dte := dteHigh; dte >= dteLow; dte -= dteStep {
				price, err := chain.CalculatePrice(assetPrice, strikePrice, dte)
				if err == nil {
					pricesPerExpiry = append(pricesPerExpiry, price)
				} else {
					return nil, err
				}
			}
			expiryPricesPerStrike = append(expiryPricesPerStrike, pricesPerExpiry)
		}
		strikeExpiryPricesPerAssetPrice = append(strikeExpiryPricesPerAssetPrice, expiryPricesPerStrike)
	}
	return strikeExpiryPricesPerAssetPrice, nil
}
