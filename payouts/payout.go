package payouts

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vertiond/verthash-one-click-miner/util"
)

type Payout interface {
	GetID() int
	GetName() string
	GetTicker() string
	GetPassword() string
	GetCoingeckoExchange() string
}

func GetPayouts(testnet bool) []Payout {
	if testnet {
		return []Payout{
			NewVTCPayout(),
		}
	}
	return []Payout{
		NewDOGEPayout(),
		//NewVTCPayout(),
		//NewBTCPayout(),
		//NewLTCPayout(),
		//NewBCHPayout(),
		//NewDASHPayout(),
	}
}

func GetPayout(payout int, testnet bool) Payout {
	payouts := GetPayouts(testnet)
	for _, p := range payouts {
		if p.GetID() == payout {
			return p
		}
	}
	return payouts[0]
}

// func GetBitcoinPerUnitCoin(coinName string, coinTicker string, coingeckoExchange string) float64 {
// 	jsonPayload := map[string]interface{}{}
// 	err := util.GetJson(fmt.Sprintf(
// 		"https://api.coingecko.com/api/v3/exchanges/%s/tickers?coin_ids=%s",
// 		coingeckoExchange, strings.ReplaceAll(strings.ToLower(coinName), " ", "-")),
// 		&jsonPayload)
// 	if err != nil {
// 		return 0.0
// 	}
// 	jsonTickersArr, ok := jsonPayload["tickers"].([]interface{})
// 	if !ok {
// 		return 0.0
// 	}

// 	result := 0.0
// 	for _, jsonTickerInfo := range jsonTickersArr {
// 		jsonTickerInfoMap := jsonTickerInfo.(map[string]interface{})
// 		jsonTickerBase, ok1 := jsonTickerInfoMap["base"]
// 		jsonTickerTarget, ok2 := jsonTickerInfoMap["target"]
// 		if !ok1 || !ok2 {
// 			continue
// 		}
// 		if jsonTickerBase == coinTicker && jsonTickerTarget == "BTC" {
// 			jsonTickerConvertedLast, ok := jsonTickerInfoMap["converted_last"].(map[string]interface{})
// 			if ok {
// 				jsonTickerConvertedLastBTC, ok := jsonTickerConvertedLast["btc"].(float64)
// 				if ok {
// 					result = jsonTickerConvertedLastBTC
// 				}
// 			}
// 			break
// 		}
// 	}
// 	return result
// }

func GetBitcoinPerUnitCoin(coinName string, coinTicker string, coingeckoExchange string) float64 {
	jsonPayload := map[string]interface{}{}
	err := util.GetJson(fmt.Sprintf(
		"https://api.cryptonator.com/api/ticker/%s-btc",
		strings.ToLower(coinTicker)),
		&jsonPayload)
	if err != nil {
		return 0.0
	}
	jsonTicker, ok := jsonPayload["ticker"].(map[string]interface{})
	if !ok {
		return 0.0
	}
	jsonTickerPrice, ok := jsonTicker["price"].(string)
	if !ok {
		return 0.0
	}
	result, _ := strconv.ParseFloat(jsonTickerPrice, 64)
	return result
}
