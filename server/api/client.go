package api

import (
	"fmt"

	"github.com/jcdevguru/option-assistant/lib/option"
	"github.com/jcdevguru/option-assistant/lib/util"
)

// Position contains call and put prices for a specific expiry date
// @Description Contains the call and put prices for a specific number of days to expiry
type Position struct {
	Price        float64 `json:"price"`        // Option price
	DaysToExpiry float64 `json:"daysToExpiry"` // Days to expiry
}

// Strike_Positions contains option prices for a specific strike price and expiry dates
// @Description Contains option prices for different expiry dates at a given strike price
type Strike_Positions struct {
	StrikePrice float64    `json:"strikePrice"` // Strike price
	Positions   []Position `json:"positions"`   // Option prices per expiry date
}

// AssetPrice_Strike_Positions contains option prices for a specific asset price and strike price
// @Description Contains strike and expiry prices for a given asset price
type AssetPrice_Strike_Positions struct {
	AssetPrice      float64            `json:"assetPrice"`      // Asset price
	StrikePositions []Strike_Positions `json:"strikePositions"` // Option price/dtes per strike
}

// OptionChainResponse represents the response structure for an option chain request
// @Description The response object for the CalculateOptionChain endpoint
type OptionChainResponse struct {
	AssetName   string                        `json:"assetName"`   // Name of the asset
	OptionChain []AssetPrice_Strike_Positions `json:"optionChain"` // Option price/dtes per asset price / strike
}

type OptionChainQuery struct {
	AssetName        string  `form:"assetName" binding:"required,min=2,alphanum"`
	OptionType       string  `form:"optionType" binding:"required,oneof=Call Put"`
	AssetPriceLow    float64 `form:"assetPriceLow" binding:"required,gt=0"`
	AssetPriceHigh   float64 `form:"assetPriceHigh" binding:"required,gtefield=AssetPriceLow"`
	AssetPriceStep   float64 `form:"assetPriceStep,default=1.0" binding:"required,gt=0.0"`
	StrikePriceLow   float64 `form:"strikePriceLow" binding:"required,gt=0"`
	StrikePriceHigh  float64 `form:"strikePriceHigh" binding:"required,gtefield=StrikePriceLow"`
	StrikePriceStep  float64 `form:"strikePriceStep,default=1.0" binding:"required,gt=0.0"`
	DaysToExpiryLow  float64 `form:"daysToExpiryLow" binding:"required,gt=0"`
	DaysToExpiryHigh float64 `form:"daysToExpiryHigh" binding:"required,gtefield=DaysToExpiryLow"`
	DaysToExpiryStep float64 `form:"daysToExpiryStep,default=1.0" binding:"required,gt=0.0"`
	RiskFreeRate     float64 `form:"riskFreeRate" binding:"required,gt=0"`
	Volatility       float64 `form:"volatility" binding:"required,gt=0"`
}

func encodeResponse(assetPriceSpan option.ValueSpan, chain option.OptionChain) []AssetPrice_Strike_Positions {
	var result []AssetPrice_Strike_Positions
	assetIndex := 0
	for assetPrice := assetPriceSpan.Low; assetPrice <= assetPriceSpan.High; assetPrice += assetPriceSpan.Step {
		strikesPerAssetPrice := chain[assetIndex]
		var strikePositions []Strike_Positions
		for _, positionsPerStrike := range strikesPerAssetPrice {
			positionsForStrike := Strike_Positions{StrikePrice: positionsPerStrike[0].Strike}
			for _, position := range positionsPerStrike {
				positionsForStrike.Positions = append(positionsForStrike.Positions, Position{Price: util.Round(position.Price, 2), DaysToExpiry: position.DaysToExpiry})
			}
			strikePositions = append(strikePositions, positionsForStrike)
		}
		result = append(result, AssetPrice_Strike_Positions{AssetPrice: assetPrice, StrikePositions: strikePositions})
	}
	return result
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
// @Param assetPriceStep query float64 false "Step amount for asset price range (default = 1.0)"
// @Param strikePriceLow query float64 true "Low end of strike price range"
// @Param strikePriceHigh query float64 true "High end of strike price range"
// @Param strikePriceStep query float64 false "Step amount for strike price range (default = 1.0)"
// @Param daysToExpiryLow query float64 true "Low end of days to expiry range"
// @Param daysToExpiryHigh query float64 true "High end of days to expiry range"
// @Param daysToExpiryStep query float64 false "Step amount for days to expiry range (default = 1.0)"
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
		optionTypeNum = option.Call

	case "Put":
		optionTypeNum = option.Put

	default:
		return OptionChainResponse{}, fmt.Errorf("unknown option type %s - use Call or Put", optionType)
	}

	assetPriceSpan := option.ValueSpan{Low: assetPriceLow, High: assetPriceHigh, Step: assetPriceStep}
	strikePriceSpan := option.ValueSpan{Low: strikePriceLow, High: strikePriceHigh, Step: strikePriceStep}
	daysToExpirySpan := option.ValueSpan{Low: daysToExpiryLow, High: daysToExpiryHigh, Step: daysToExpiryStep}

	optionChain, err := option.NewOptionChain(optionTypeNum, volatility, riskFreeRate, daysToExpiryHigh)
	if err != nil {
		return OptionChainResponse{}, err
	}

	chainValues, err := optionChain.ComputeOptionChain(&assetPriceSpan, &strikePriceSpan, &daysToExpirySpan)
	if err != nil {
		return OptionChainResponse{}, err
	}

	response := OptionChainResponse{
		AssetName:   assetName,
		OptionChain: encodeResponse(assetPriceSpan, chainValues),
	}

	return response, nil
}
