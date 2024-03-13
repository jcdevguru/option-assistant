package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jcdevguru/option-assistant/server/api"
	docs "github.com/jcdevguru/option-assistant/server/docs" // import generated docs
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func getOptionChain(c *gin.Context) {
	var query api.OptionChainQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call CalculateOptionChain with the extracted parameters
	optionChain, err := api.OptionChain(
		query.AssetName, query.OptionType,
		query.AssetPriceLow, query.AssetPriceHigh, query.AssetPriceStep,
		query.StrikePriceLow, query.StrikePriceHigh, query.StrikePriceStep,
		query.DaysToExpiryLow, query.DaysToExpiryHigh, query.DaysToExpiryStep,
		query.RiskFreeRate, query.Volatility,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the result as JSON
	c.JSON(http.StatusOK, optionChain)
}

func main() {
	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/"
	router.SetTrustedProxies(nil)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/optionChain", getOptionChain)

	router.Run("localhost:8080")
}
