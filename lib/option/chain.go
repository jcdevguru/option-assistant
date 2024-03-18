package option

import (
	"fmt"
	"math"
)

// Standard normal cumulative distribution function
func normalizedCDF(x float64) float64 {
	return 0.5 * (1.0 + math.Erf(x/math.Sqrt(2.0)))
}

func validateSpan(name string, span *ValueSpan) (float64, float64, float64, error) {
	var err error
	if span.Low < 0 || span.High < 0 || span.Step < 0 {
		err = fmt.Errorf("%s: span Low, High, Step cannot be negative", name)
	} else if span.Low > span.High {
		err = fmt.Errorf("%s: span Low > High", name)
	} else if span.High > span.Low && span.Step == 0 {
		err = fmt.Errorf("%s: span Step must be > 0 for Span.High > Span.Low", name)
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

func NewOptionChain(optionType int, volatility, riskFreeRate, expiryInDays float64) (*OptionChainCalculator, error) {
	chain := OptionChainCalculator{
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
	chain.calculatePrice = priceCalculator
	chain.d1d2CalculateFuncMap = make(map[float64]d1d2CalculateFunc)
	chain.d1d2CalculationValueMap = make(map[priceKey]d1d2Calculation)

	return &chain, nil
}

// Black-Scholes formula for call option price

func (chain *OptionChainCalculator) d1d2calculator(daysToExpiry float64) (d1d2CalculateFunc, error) {
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

func (chain *OptionChainCalculator) calculateD1D2(assetPrice, strikePrice, daysToExpiry float64) (*d1d2Calculation, error) {
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

func (chain *OptionChainCalculator) BlackScholesCall(assetPrice, strikePrice, daysToExpiry float64, position *OptionPosition) error {
	d1d2, err := chain.calculateD1D2(assetPrice, strikePrice, daysToExpiry)
	if err != nil {
		return err
	}
	price := assetPrice*normalizedCDF(d1d2.d1) - strikePrice*math.Exp(-chain.RiskFreeRate*d1d2.yearsToExpiry)*normalizedCDF(d1d2.d2)
	position.Price = price
	position.Strike = strikePrice
	position.DaysToExpiry = daysToExpiry
	return nil
}

// Black-Scholes formula for put option price
func (chain *OptionChainCalculator) BlackScholesPut(assetPrice, strikePrice, daysToExpiry float64, position *OptionPosition) error {
	d1d2, err := chain.calculateD1D2(assetPrice, strikePrice, daysToExpiry)
	if err != nil {
		return err
	}
	price := strikePrice*math.Exp(-chain.RiskFreeRate*d1d2.yearsToExpiry)*normalizedCDF(-d1d2.d2) - assetPrice*normalizedCDF(-d1d2.d1)
	position.Price = price
	position.Strike = strikePrice
	position.DaysToExpiry = daysToExpiry
	return nil
}

func (chain *OptionChainCalculator) ComputeOptionChain(assetPriceSpan, strikePriceSpan, daysToExpirySpan *ValueSpan) (OptionChain, error) {
	apLow, apHigh, apStep, err := validateSpan("assetPriceRange", assetPriceSpan)
	if err != nil {
		return nil, err
	}

	spLow, spHigh, spStep, err := validateSpan("strikePriceRange", strikePriceSpan)
	if err != nil {
		return nil, err
	}

	dteLow, dteHigh, dteStep, err := validateSpan("daysToExpirySpan", daysToExpirySpan)
	if err != nil {
		return nil, err
	}
	if dteLow == 0.0 {
		dteLow += dteStep
	}

	var result OptionChain
	for assetPrice := apLow; assetPrice <= apHigh; assetPrice += apStep {
		var strikePositionsPerAssetPrice [][]OptionPosition
		for strikePrice := spLow; strikePrice <= spHigh; strikePrice += spStep {
			var positionsPerStrike []OptionPosition
			for dte := dteHigh; dte >= dteLow; dte -= dteStep {
				var optionPosition OptionPosition
				err := chain.calculatePrice(assetPrice, strikePrice, dte, &optionPosition)
				if err == nil {
					positionsPerStrike = append(positionsPerStrike, optionPosition)
				} else {
					return nil, err
				}
			}
			strikePositionsPerAssetPrice = append(strikePositionsPerAssetPrice, positionsPerStrike)
		}
		result = append(result, strikePositionsPerAssetPrice)
	}
	return result, nil
}
