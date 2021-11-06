package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDedupe(t *testing.T) {

	t.Run("Test deduplication", func(t *testing.T) {
		sample := []string{"fnord", "fnord", "foo", "foo", "foo"}

		got := Dedupe(sample)
		want := []string{"fnord", "foo"}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %q want %q", got, want)
		}
	})

}

func TestGetAssetById(t *testing.T) {

	t.Run("Test Getting Asset Details by Contract Address", func(t *testing.T) {
		gin.SetMode(gin.ReleaseMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Params = []gin.Param{
			gin.Param{Key: "id", Value: "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"},
		}

		GetAssetByID(c)

		var asset AssetResponse

		data, _ := ioutil.ReadAll(w.Body)

		json.Unmarshal(data, &asset)

		if w.Code != 200 {
			t.Fatalf("Expected HTTP 200, got %d", w.Code)
		}

		if asset.Symbol != "USDC" {
			t.Fatalf("Expected asset symbol to be USDC, got %s", asset.Symbol)
		}

		if asset.Name != "USD Coin" {
			t.Fatalf("Expected asset name to be USD Coin, got %s", asset.Name)
		}

		if asset.ContractAddress != "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48" {
			t.Fatalf("Expected asset contract address to be 0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48, got %s", asset.ContractAddress)
		}

		if len(asset.Pools) == 0 {
			t.Fatalf("Expected non-zero length of whitelisted pools, got %d", len(asset.Pools))
		}
	})

}

func TestGetBlockByNumber(t *testing.T) {

	t.Run("Test Getting Block Details by Block Number", func(t *testing.T) {
		gin.SetMode(gin.ReleaseMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Params = []gin.Param{
			gin.Param{Key: "blocknumber", Value: "13315378"},
		}

		GetBlockByNumber(c)

		var block BlockResponse

		data, _ := ioutil.ReadAll(w.Body)

		json.Unmarshal(data, &block)

		if w.Code != 200 {
			t.Fatalf("Expected HTTP 200, got %d", w.Code)
		}

		if block.BlockHeight != "13315378" {
			t.Fatalf("Expected block height to be 13315378, got %s", block.BlockHeight)
		}

		if len(block.Swaps) != 116 {
			t.Fatalf("Expected number of swaps to be 116, got %d", len(block.Swaps))
		}

		if len(block.AssetsSwapped) != 54 {
			t.Fatalf("Expected number of assets swapped to be 54, got %d", len(block.AssetsSwapped))
		}
	})

}
