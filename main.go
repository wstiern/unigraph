package main

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/machinebox/graphql"
)

type AssetResponse struct {
	Name            string   `json:"Name"`
	Symbol          string   `json:"Symbol"`
	ContractAddress string   `json:"ContractAddress"`
	VolumeUSD       string   `json:"VolumeUSD"`
	Pools           []string `json:"Pools"`
}

type BlockResponse struct {
	BlockHeight   string   `json:"BlockHeight"`
	Swaps         []string `json:"Swaps"`
	AssetsSwapped []string `json:"AssetsSwapped"`
}

var graphAPI = "https://api.thegraph.com/subgraphs/name/ianlapham/uniswap-v3-alt"

func dedupe(input []string) []string {

	duped := map[string]bool{}
	result := []string{}

	for v := range input {
		if !duped[input[v]] {
			duped[input[v]] = true
			result = append(result, input[v])
		}
	}

	return result
}

func getAssetByID(ctx *gin.Context) {

	assetId := ctx.Param("id")
	query := fmt.Sprintf(`
	{
		token(id:"%s") {
		  id
		  name
		  symbol
		  volumeUSD
		  whitelistPools {
			token0 {
			  name
			  symbol
			}
			token1 {
			  name
			  symbol
			}
		  }
		}
	  }`, assetId)

	client := graphql.NewClient(graphAPI)
	request := graphql.NewRequest(query)

	var result map[string]interface{}

	if err := client.Run(
		context.Background(),
		request,
		&result); err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": "Contract address not found"})
		return
	}

	tokenData := result["token"].(map[string]interface{})

	tokenName := tokenData["name"]
	tokenSymbol := tokenData["symbol"]
	tokenContractAddress := tokenData["id"]
	tokenVolumeUSD := tokenData["volumeUSD"]
	tokenWhitelistPools := tokenData["whitelistPools"].([]interface{})

	var pairs []string

	// unpack whitelistPools[]
	for _, pool := range tokenWhitelistPools {

		var in string
		var out string

		// unpack pool{}
		if pair, ok := pool.(map[string]interface{}); ok {
			for index, token := range pair {

				if coin, ok := token.(map[string]interface{}); ok {

					if index == "token0" {
						in = coin["symbol"].(string)
					} else {
						out = coin["symbol"].(string)
					}
				}
			}
		}
		currentPair := fmt.Sprintf("%s:%s", in, out)
		pairs = append(pairs, currentPair)
	}

	sort.Strings(pairs)
	p := dedupe(pairs)

	response := &AssetResponse{
		Name:            tokenName.(string),
		Symbol:          tokenSymbol.(string),
		ContractAddress: tokenContractAddress.(string),
		VolumeUSD:       tokenVolumeUSD.(string),
		Pools:           p,
	}

	ctx.IndentedJSON(http.StatusOK, response)
}

func getBlockByNumber(ctx *gin.Context) {

	blockNumber := ctx.Param("blocknumber")
	query := fmt.Sprintf(`
	{
		transactions( block: { number:%s } )
		{
			swaps {
				timestamp
				amount0
				amount1
				token0 {
					symbol
				}
				token1 {
					symbol
				}
			}
		}
	}`, blockNumber)

	client := graphql.NewClient(graphAPI)
	request := graphql.NewRequest(query)

	var result map[string]interface{}

	if err := client.Run(
		context.Background(),
		request,
		&result); err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": "Block number not found"})
		return
	}

	transactionData := result["transactions"].([]interface{})

	var assets []string
	var swaps []string

	// unpack transactions[]
	for _, txs := range transactionData {

		var in string
		var out string
		var inAmount string
		var outAmount string

		if swap, ok := txs.(map[string]interface{}); ok {

			// unpack swaps[]
			for _, v := range swap {
				if data, ok := v.([]interface{}); ok {

					// unpack swap{} in swaps[]
					for _, v := range data {
						if s, ok := v.(map[string]interface{}); ok {

							inAmount = s["amount0"].(string)
							outAmount = s["amount1"].(string)

							// unpack token0{} in swap{}
							if symbol, ok := s["token0"].(map[string]interface{}); ok {
								in = symbol["symbol"].(string)
							}

							// unpack token0{} in swap{}
							if symbol, ok := s["token1"].(map[string]interface{}); ok {
								out = symbol["symbol"].(string)
							}

							// TODO: left hand number is always negative and right hand number is always positive
							currentSwap := fmt.Sprintf("%s %s : %s %s", in, inAmount, out, outAmount)
							swaps = append(swaps, currentSwap)
							assets = append(assets, in, out)
						}
					}

				}
			}
		}
	}

	sort.Strings(assets)
	a := dedupe(assets)

	response := &BlockResponse{
		BlockHeight:   blockNumber,
		Swaps:         swaps,
		AssetsSwapped: a,
	}

	ctx.IndentedJSON(http.StatusOK, response)
}

func setup() {
	gin.SetMode("release")
	app := gin.Default()
	app.GET("/asset/:id", getAssetByID)
	app.GET("/block/:blocknumber", getBlockByNumber)
	app.Run("localhost:8000")
}

func main() {
	setup()
}
