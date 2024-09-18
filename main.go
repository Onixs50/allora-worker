package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/rand"
)

type Kline struct {
	OpenTime  time.Time
	CloseTime time.Time
	Interval  string
	Symbol    string
	Open      string
	High      string
	Low       string
	Close     string
	Volume    string
	Closed    bool
}

type envConfig struct {
	APIKey              string
	RPC                 string
	CoinGeckoAPIKey     string
	CryptoCompareAPIKey string
}

func main() {
	cfg := &envConfig{
		APIKey:              os.Getenv("UPSHOT_APIKEY"),
		RPC:                 os.Getenv("RPC"),
		CoinGeckoAPIKey:     os.Getenv("COINGECKO_APIKEY"),
		CryptoCompareAPIKey: os.Getenv("CRYPTOCOMPARE_APIKEY"),
	}

	fmt.Println("UPSHOT_APIKEY: ", cfg.APIKey)
	fmt.Println("RPC: ", cfg.RPC)
	fmt.Println("COINGECKO_APIKEY: ", cfg.CoinGeckoAPIKey)
	fmt.Println("CRYPTOCOMPARE_APIKEY: ", cfg.CryptoCompareAPIKey)

	router := gin.Default()

	router.GET("/inference/:token", func(c *gin.Context) {
		token := c.Param("token")
		switch token {
		case "MEME":
			handleMemeRequest(c, cfg)
		case "DeFi":
			handleDeFiRequest(c, cfg)
		case "NFT":
			handleNFTRequest(c, cfg)
		default:
			handleCryptoRequest(c, cfg, token)
		}
	})

	router.Run(":8000")
}

func handleCryptoRequest(c *gin.Context, cfg *envConfig, token string) {
	symbol := fmt.Sprintf("%sUSDT", token)

	k, err := getLastKlines(symbol, "15m")
	if err != nil {
		fmt.Println(err)
		c.String(500, "Error fetching klines")
		return
	}

	rate, err := calculatePriceChangeRate(*k)
	if err != nil {
		fmt.Println(err)
		c.String(500, "Error calculating price change rate")
		return
	}
	rate = multiplyChangeRate(rate)
	close, _ := strconv.ParseFloat(k.Close, 64)
	price := close + (close * rate)

	// Get additional data from CoinGecko
	cgPrice, err := getCoinGeckoPrice(token, cfg.CoinGeckoAPIKey)
	if err != nil {
		fmt.Println("CoinGecko API error:", err)
	}

	// Get additional data from CryptoCompare
	ccPrice, err := getCryptoComparePrice(token, cfg.CryptoCompareAPIKey)
	if err != nil {
		fmt.Println("CryptoCompare API error:", err)
	}

	// Calculate weighted average
	weightedPrice := (price + cgPrice + ccPrice) / 3

	c.JSON(200, gin.H{
		"price":               weightedPrice,
		"binance_price":       price,
		"coingecko_price":     cgPrice,
		"cryptocompare_price": ccPrice,
	})
}

func handleMemeRequest(c *gin.Context, cfg *envConfig) {
	if cfg.APIKey == "" {
		c.String(400, "need api key")
		return
	}

	if cfg.RPC == "" {
		c.String(500, "Invalid RPC configuration")
		return
	}

	lb, err := getLatestBlock(cfg.RPC)
	if err != nil {
		fmt.Println(err)
		c.String(500, "Error fetching latest block")
		return
	}

	meme, err := getMemeOracleData(lb, cfg.APIKey)
	if err != nil {
		fmt.Println(err)
		c.String(500, "Error fetching meme oracle data")
		return
	}

	mp, err := getMemePrice(meme.Data.Platform, meme.Data.Address)
	if err != nil {
		fmt.Println(err)
		c.String(500, "Error fetching meme price")
		return
	}

	fmt.Printf("\nBlockHeight: \"%s\", Meme: \"%s\", Platform: \"%s\", Price: \"%s\"\n\n",
		lb, meme.Data.TokenSymbol, meme.Data.Platform, mp)

	mpf, _ := strconv.ParseFloat(mp, 64)

	c.String(http.StatusOK, strconv.FormatFloat(random(mpf), 'g', -1, 64))
}

func handleDeFiRequest(c *gin.Context, cfg *envConfig) {
	// Example implementation - you should replace this with actual DeFi data fetching and analysis
	totalValueLocked, err := getTotalValueLocked()
	if err != nil {
		c.String(500, "Error fetching DeFi data")
		return
	}

	yieldFarmingRate := calculateYieldFarmingRate()

	c.JSON(200, gin.H{
		"total_value_locked": totalValueLocked,
		"yield_farming_rate": yieldFarmingRate,
		"defi_score":         random(100), // Example random DeFi score
	})
}

func handleNFTRequest(c *gin.Context, cfg *envConfig) {
	// Example implementation - you should replace this with actual NFT data fetching and analysis
	floorPrice, err := getNFTFloorPrice("example_collection")
	if err != nil {
		c.String(500, "Error fetching NFT data")
		return
	}

	tradingVolume := getNFTTradingVolume()

	c.JSON(200, gin.H{
		"floor_price":    floorPrice,
		"trading_volume": tradingVolume,
		"nft_score":      random(100), // Example random NFT score
	})
}

func getLastKlines(symbol, interval string) (*Kline, error) {
	ur, _ := url.Parse("https://api.binance.com/api/v1/klines")
	queryParams := url.Values{}
	queryParams.Add("endTime", strconv.Itoa(int(time.Now().UnixMilli())))
	queryParams.Add("limit", "1")
	queryParams.Add("symbol", symbol)
	queryParams.Add("interval", interval)
	ur.RawQuery = queryParams.Encode()
	resp, err := http.DefaultClient.Get(ur.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	var ks [][]interface{}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &ks)
	if err != nil {
		return nil, err
	}

	if len(ks) == 0 {
		return nil, fmt.Errorf("no klines data")
	}

	kline := ks[0]
	return &Kline{
		OpenTime: time.UnixMilli(int64(kline[0].(float64))),
		Interval: interval,
		Symbol:   symbol,
		Open:     kline[1].(string),
		High:     kline[4].(string),
		Low:      kline[2].(string),
		Close:    kline[3].(string),
		Volume:   kline[5].(string),
	}, nil
}

func calculatePriceChangeRate(kline Kline) (float64, error) {
	open, err := strconv.ParseFloat(kline.Open, 64)
	if err != nil {
		return 0, err
	}
	close, err := strconv.ParseFloat(kline.Close, 64)
	if err != nil {
		return 0, err
	}

	if open == 0 {
		return 0, fmt.Errorf("open price cannot be zero")
	}

	priceChangeRate := (close - open) / open
	return priceChangeRate, nil
}

func multiplyChangeRate(changeRate float64) float64 {
	r := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))

	multiplier := r.Float64()*0.8 + 0.1
	newChangeRate := changeRate * multiplier
	return newChangeRate + changeRate
}

func getCoinGeckoPrice(token string, apiKey string) (float64, error) {
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd&x_cg_demo_api_key=%s", token, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result[token]["usd"], nil
}

func getCryptoComparePrice(token string, apiKey string) (float64, error) {
	url := fmt.Sprintf("https://min-api.cryptocompare.com/data/price?fsym=%s&tsyms=USD&api_key=%s", token, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result["USD"], nil
}

func getMemePrice(network, memeAddress string) (string, error) {
	url := fmt.Sprintf("https://api.geckoterminal.com/api/v2/simple/networks/%s/token_price/%s", network, memeAddress)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create new request: %w", err)
	}
	req.Header.Set("accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	res := &tokenPriceResponse{}
	err = json.Unmarshal(body, res)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return res.Data.Attributes.TokenPrices[memeAddress], nil
}

type tokenPriceResponse struct {
	Data struct {
		Attributes struct {
			TokenPrices map[string]string `json:"token_prices"`
		} `json:"attributes"`
	} `json:"data"`
}

type latestBlockResponse struct {
	Result struct {
		SyncInfo struct {
			LatestBlockHeight string `json:"latest_block_height"`
		} `json:"sync_info"`
	} `json:"result"`
}

func getLatestBlock(rpc string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/status", rpc), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create new request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var response latestBlockResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.Result.SyncInfo.LatestBlockHeight, nil
}

type memeOracleResponse struct {
	RequestID string `json:"request_id"`
	Status    bool   `json:"status"`
	Data      struct {
		TokenID     string `json:"token_id"`
		TokenSymbol string `json:"token_symbol"`
		Platform    string `json:"platform"`
		Address     string `json:"address"`
	} `json:"data"`
}

func getMemeOracleData(blockHeight string, apiKey string) (*memeOracleResponse, error) {
	url := fmt.Sprintf("https://api.upshot.xyz/v2/allora/tokens-oracle/token/%s", blockHeight)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("x-api-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	res := &memeOracleResponse{}
	err = json.Unmarshal(body, res)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return res, nil
}

func random(price float64) float64 {
	randomPercent := rand.Float64()*6 - 3

	priceChange := price * (randomPercent / 100)

	return price + priceChange
}

// Example functions for DeFi and NFT - replace these with actual implementations
func getTotalValueLocked() (float64, error) {
	// Implement actual TVL fetching logic here
	return 1000000000, nil // Example $1 billion TVL
}

func calculateYieldFarmingRate() float64 {
	// Implement actual yield farming rate calculation here
	return 0.05 // Example 5% APY
}

func getNFTFloorPrice(collection string) (float64, error) {
	// Implement actual NFT floor price fetching logic here
	return 1.5, nil // Example 1.5 ETH floor price
}
