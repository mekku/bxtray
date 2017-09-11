package main

import (
	"fmt"
	"time"
	"encoding/json"
	"encoding/hex"
	"crypto/sha256"
	"gopkg.in/resty.v0"
	"strconv"

	"github.com/getlantern/systray"
)

type Pairing struct {
	PairingID         int     `json:"pairing_id"`
	PrimaryCurrency   string  `json:"primary_currency"`
	SecondaryCurrency string  `json:"secondary_currency"`
	Change            float64 `json:"change"`
	LastPrice         float64 `json:"last_price"`
	Volume24Hours     float64 `json:"volume_24hours"`
	Orderbook         struct {
		Bids struct {
			Total   int     `json:"total"`
			Volume  float64 `json:"volume"`
			Highbid float64     `json:"highbid"`
		} `json:"bids"`
		Asks struct {
			Total   int     `json:"total"`
			Volume  float64 `json:"volume"`
			Highbid float64 `json:"highbid"`
		} `json:"asks"`
	} `json:"orderbook"`
}

type BalanceStruct struct {
	Total       json.Number `json:"total"`
	Available   json.Number `json:"available"`
	Orders      json.Number    `json:"orders"`
	Withdrawals int    `json:"withdrawals"`
	Deposits    int    `json:"deposits"`
	Options     int    `json:"options"`
}

type BalanceResponse struct {
	Success bool `json:"success"`
	Balance map[string]*BalanceStruct `json:"balance"`
}


func main() {

	onExit := func() {
	}
	// Should be called at the very beginning of main().
	systray.Run(onReady, onExit)

   	
}

func onReady() {

	thb := 0.0
   	omg := 0.0
   	btc := 0.0
   	currentValue := 0.0

	mQuit := systray.AddMenuItem("Quit", "Quit")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()

   	
    lastPriceOMG := 0.0
    lastPriceBTC := 0.0
    apikey := ""
    apisecret := ""
    nonce := ""

    var balanceResp BalanceResponse

	for {
		
		resp, err := resty.R().SetHeader("Content-Type", "application/json").Get("https://bx.in.th/api/")
		// Get price
		if err != nil {
			fmt.Printf("\nError: %v", err)
		} else {
		
			m := map[string]Pairing{}
			jerr := json.Unmarshal([]byte(resp.Body()), &m)
			if jerr != nil {
			    panic(err)
			}

			lastPriceOMG = m["26"].LastPrice
			lastPriceBTC = m["1"].LastPrice

		}

		if apikey != "" && apisecret != "" {
			nonce = strconv.FormatInt(time.Now().UnixNano() , 10)
			hash := sha256.New()
			hash.Write([]byte(apikey + nonce +  apisecret))
			md := hash.Sum(nil)
			sum := hex.EncodeToString(md)

			uresp, err := resty.R().SetFormData(map[string]string{"key":apikey, "nonce":nonce, "signature":sum}).Post("https://bx.in.th/api/balance/")

			// Get balance
			if err != nil {
				fmt.Printf("\nError: %v", err)
			} else {
				jerr := json.Unmarshal([]byte(uresp.Body()), &balanceResp)
				if jerr != nil {
				    panic(jerr)
				}

				btc, _ = balanceResp.Balance["BTC"].Available.Float64()
				omg, _ = balanceResp.Balance["OMG"].Available.Float64()
				thb, _ = balanceResp.Balance["THB"].Available.Float64()
			}

			currentValue = ((omg * lastPriceOMG ) + (btc * lastPriceBTC)) +thb
			systray.SetTitle(fmt.Sprintf("OMG: %v/%v   |   BTC: %v/%v   |   VALUES(THB): %.2f", omg, lastPriceOMG, btc, lastPriceBTC, currentValue))

			time.Sleep(30 * time.Second)
		} else {
			systray.SetTitle(fmt.Sprintf("OMG: %v   |   BTC: %v", lastPriceOMG, lastPriceBTC))
		}

	}

}