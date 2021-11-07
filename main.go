package main

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/machinebox/graphql"
)

type AssetDetails struct {
	Token struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		Symbol         string `json:"symbol"`
		VolumeUSD      string `json:"volumeUSD"`
		WhitelistPools []struct {
			Token0 struct {
				Name   string `json:"name"`
				Symbol string `json:"symbol"`
			} `json:"token0"`
			Token1 struct {
				Name   string `json:"name"`
				Symbol string `json:"symbol"`
			} `json:"token1"`
		} `json:"whitelistPools"`
	} `json:"token"`
}

type BlockDetails struct {
	Transactions []struct {
		Swaps []struct {
			Amount0   string `json:"amount0"`
			Amount1   string `json:"amount1"`
			Timestamp string `json:"timestamp"`
			Token0    struct {
				Symbol string `json:"symbol"`
			} `json:"token0"`
			Token1 struct {
				Symbol string `json:"symbol"`
			} `json:"token1"`
		} `json:"swaps"`
	} `json:"transactions"`
}

type AssetResponse struct {
	Name            string   `json:"name"`
	Symbol          string   `json:"symbol"`
	ContractAddress string   `json:"contractAddress"`
	VolumeUSD       string   `json:"volumeUSD"`
	Pools           []string `json:"pools"`
}

type BlockResponse struct {
	BlockHeight   string   `json:"blockHeight"`
	Swaps         []string `json:"swaps"`
	AssetsSwapped []string `json:"assetsSwapped"`
}

var graphAPI = "https://api.thegraph.com/subgraphs/name/ianlapham/uniswap-v3-alt"

func Dedupe(input []string) []string {

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

func GetAssetByID(ctx *gin.Context) {

	// prep gql query
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

	// send it
	client := graphql.NewClient(graphAPI)
	request := graphql.NewRequest(query)

	var result AssetDetails

	if err := client.Run(
		context.Background(),
		request,
		&result); err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": "Contract address not found"})
		return
	}

	// parse results
	tokenName := result.Token.Name
	tokenSymbol := result.Token.Symbol
	tokenContractAddress := result.Token.ID
	tokenVolumeUSD := result.Token.VolumeUSD

	var pairs []string

	// collect list of active pairs in all pools
	for _, pool := range result.Token.WhitelistPools {
		in := pool.Token0.Symbol
		out := pool.Token1.Symbol
		currentPair := fmt.Sprintf("%s:%s", in, out)
		pairs = append(pairs, currentPair)
	}

	// sort + dedupe list
	sort.Strings(pairs)
	p := Dedupe(pairs)

	// assemble response
	response := &AssetResponse{
		Name:            tokenName,
		Symbol:          tokenSymbol,
		ContractAddress: tokenContractAddress,
		VolumeUSD:       tokenVolumeUSD,
		Pools:           p,
	}

	// return response
	ctx.IndentedJSON(http.StatusOK, response)
}

func GetBlockByNumber(ctx *gin.Context) {

	// prep gql query
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

	// full send
	client := graphql.NewClient(graphAPI)
	request := graphql.NewRequest(query)

	var result BlockDetails

	if err := client.Run(
		context.Background(),
		request,
		&result); err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": "Block number not found"})
		return
	}

	// empty list to collect assets that were traded
	var assets []string
	// empty list to collect swaps that were made
	var swaps []string

	// parse transactions
	for _, txs := range result.Transactions {
		if len(txs.Swaps) > 0 {
			for _, swap := range txs.Swaps {
				in := swap.Token0.Symbol
				out := swap.Token1.Symbol
				inAmount := swap.Amount0
				outAmount := swap.Amount1
				currentSwap := fmt.Sprintf("%s %s : %s %s", in, inAmount, out, outAmount)
				swaps = append(swaps, currentSwap)
				assets = append(assets, in, out)
			}
		}
	}

	// sort + dedupe assets
	sort.Strings(assets)
	a := Dedupe(assets)

	// assemble response
	response := &BlockResponse{
		BlockHeight:   blockNumber,
		Swaps:         swaps,
		AssetsSwapped: a,
	}

	// return response
	ctx.IndentedJSON(http.StatusOK, response)
}

func initWebserver() {
	gin.SetMode("release")
	app := gin.Default()
	app.GET("/asset/:id", GetAssetByID)
	app.GET("/block/:blocknumber", GetBlockByNumber)
	app.Run("localhost:8000")
}

func main() {
	initWebserver()
}
