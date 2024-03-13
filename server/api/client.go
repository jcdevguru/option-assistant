package api

import (
	"fmt"

	"github.com/jcdevguru/option-assistant/lib"
)

// ExpiryPrice contains call and put prices for a specific expiry date
// @Description Contains the call and put prices for a specific number of days to expiry
type ExpiryPrice struct {
	DaysToExpiry float64 `json:"daysToExpiry"` // Days to expiry
	Price        float64 `json:"price"`        // Call option price
}

// StrikeExpiryPrice contains option prices for a specific strike price and expiry dates
// @Description Contains call and put prices for different expiry dates at a given strike price
type StrikeExpiryPrice struct {
	StrikePrice  float64       `json:"strikePrice"`  // Strike price
	ExpiryPrices []ExpiryPrice `json:"expiryPrices"` // Call and put prices per expiry date
}

// AssetStrikeExpiryPrice contains option prices for a specific asset price and strike price
// @Description Contains strike and expiry prices for a given asset price
type AssetStrikeExpiryPrice struct {
	AssetPrice         float64             `json:"assetPrice"`         // Asset price
	StrikeExpiryPrices []StrikeExpiryPrice `json:"strikeExpiryPrices"` // Prices per strike and expiry
}

type OptionChainRequest struct {
	OptionType       string  `json:"optionType"`
	AssetPriceLow    float64 `json:"assetPriceLow"`
	AssetPriceHigh   float64 `json:"assetPriceHigh"`
	AssetPriceStep   float64 `json:"assetPriceStep"`
	StrikePriceLow   float64 `json:"strikePriceLow"`
	StrikePriceHigh  float64 `json:"strikePriceHigh"`
	StrikePriceStep  float64 `json:"strikePriceStep"`
	DaysToExpiryLow  float64 `json:"daysToExpiryLow"`
	DaysToExpiryHigh float64 `json:"daysToExpiryHigh"`
	DaysToExpiryStep float64 `json:"daysToExpiryStep"`
	RiskFreeRate     float64 `json:"riskFreeRate"`
	Volatility       float64 `json:"volatility"`
}

type OptionChainValues [][][]float64

// OptionChainResponse represents the response structure for an option chain request
// @Description The response object for the CalculateOptionChain endpoint
type OptionChainResponse struct {
	AssetName    string             `json:"assetName"` // Name of the asset
	ChainRequest OptionChainRequest `json:"request"`   // Option chain request
	ChainValues  OptionChainValues  `json:"values"`    // Option prices per asset price and strike price
}

type OptionChainQuery struct {
	AssetName        string  `form:"assetName" binding:"required,min=2,alphanum"`
	OptionType       string  `form:"optionType" binding:"required,oneof=Call Put"`
	AssetPriceLow    float64 `form:"assetPriceLow" binding:"required,gt=0"`
	AssetPriceHigh   float64 `form:"assetPriceHigh" binding:"required,gtfield=AssetPriceLow"`
	AssetPriceStep   float64 `form:"assetPriceStep" binding:"required,gt=0"`
	StrikePriceLow   float64 `form:"strikePriceLow" binding:"required,gt=0"`
	StrikePriceHigh  float64 `form:"strikePriceHigh" binding:"required,gtfield=StrikePriceLow"`
	StrikePriceStep  float64 `form:"strikePriceStep" binding:"required,gt=0"`
	DaysToExpiryLow  float64 `form:"daysToExpiryLow" binding:"required,gt=0"`
	DaysToExpiryHigh float64 `form:"daysToExpiryHigh" binding:"required,gtfield=DaysToExpiryLow"`
	DaysToExpiryStep float64 `form:"daysToExpiryStep" binding:"required,gt=0"`
	RiskFreeRate     float64 `form:"riskFreeRate" binding:"required,gt=0"`
	Volatility       float64 `form:"volatility" binding:"required,gt=0"`
}

// OptionChain godoc
// @Summary Calculate option chain
// @Description Calculates option prices for a range of asset prices, strike prices, and days to expiry.
// @Tags options
// @Accept  json
// @Produce  json
// @Param assetName query string true "Name of asset"
// @Param optionType query string true "Type of option (Call, Put)"
// @Param assetPriceLow query float64 true "Low end of asset price range"
// @Param assetPriceHigh query float64 true "High end of asset price range"
// @Param assetPriceStep query float64 true "Step amount for asset price range"
// @Param strikePriceLow query float64 true "Low end of strike price range"
// @Param strikePriceHigh query float64 true "High end of strike price range"
// @Param strikePriceStep query float64 true "Step amount for strike price range"
// @Param daysToExpiryLow query float64 true "Low end of days to expiry range"
// @Param daysToExpiryHigh query float64 true "High end of days to expiry range"
// @Param daysToExpiryStep query float64 true "Step amount for days to expiry range"
// @Param riskFreeRate query float64 true "Risk-free interest rate"
// @Param volatility query float64 true "Volatility of the asset"
// @Success 200 {object} OptionChainResponse
// @Router /optionChain [get]
func OptionChain(
	assetName, optionType string,
	assetPriceLow, assetPriceHigh, assetPriceStep,
	strikePriceLow, strikePriceHigh, strikePriceStep,
	daysToExpiryLow, daysToExpiryHigh, daysToExpiryStep,
	riskFreeRate, volatility float64,
) (OptionChainResponse, error) {
	var optionTypeNum int
	var err error
	switch optionType {
	case "Call":
		optionTypeNum = lib.Call

	case "Put":
		optionTypeNum = lib.Put

	default:
		return OptionChainResponse{}, fmt.Errorf("unknown option type %s - use Call or Put", optionType)
	}

	assetPriceSpan := lib.ValueSpan{Low: assetPriceLow, High: assetPriceHigh, Step: assetPriceStep}
	strikePriceSpan := lib.ValueSpan{Low: strikePriceLow, High: strikePriceHigh, Step: strikePriceStep}
	daysToExpirySpan := lib.ValueSpan{Low: daysToExpiryLow, High: daysToExpiryHigh, Step: daysToExpiryStep}

	optionChain, err := lib.NewOptionChain(optionTypeNum, volatility, riskFreeRate, daysToExpiryHigh)
	if err != nil {
		return OptionChainResponse{}, err
	}

	chainValues, err := optionChain.ComputeOptionChain(&assetPriceSpan, &strikePriceSpan, &daysToExpirySpan)
	if err != nil {
		return OptionChainResponse{}, err
	}

	response := OptionChainResponse{
		AssetName: assetName,
		ChainRequest: OptionChainRequest{
			OptionType:       optionType,
			AssetPriceLow:    assetPriceLow,
			AssetPriceHigh:   assetPriceHigh,
			AssetPriceStep:   assetPriceStep,
			StrikePriceLow:   strikePriceLow,
			StrikePriceHigh:  strikePriceHigh,
			StrikePriceStep:  strikePriceStep,
			DaysToExpiryLow:  daysToExpiryLow,
			DaysToExpiryHigh: daysToExpiryHigh,
			DaysToExpiryStep: daysToExpiryStep,
			RiskFreeRate:     riskFreeRate,
			Volatility:       volatility,
		},
		ChainValues: chainValues,
	}

	return response, nil
}
