package bitfinex

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

type Bitfinex struct {
	API_KEY        string
	API_SECRET_KEY string
}

func (bfx *Bitfinex) Sign(path string, nonce int64, data []byte) string {
	signature := hmac.New(sha512.New384, []byte(bfx.API_SECRET_KEY))

	payload := fmt.Sprintf("/api/%v%d%v", path, nonce, string(data))
	fmt.Println(payload)
	signature.Write([]byte(payload))
	return hex.EncodeToString(signature.Sum(nil))
}

func (bfx *Bitfinex) Call(method string, path string, data map[string]interface{}) (gjson.Result, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return gjson.Result{}, err
	}

	config, err := http.NewRequest(strings.ToUpper(method), "https://api.bitfinex.com/"+path, bytes.NewBuffer(body))
	if err != nil {
		return gjson.Result{}, err
	}

	nonce := time.Now().Unix() * 1000

	config.Header.Add("Content-Type", "application/json")
	config.Header.Add("bfx-nonce", fmt.Sprintf("%d", nonce))
	config.Header.Add("bfx-apikey", bfx.API_KEY)
	config.Header.Add("bfx-signature", bfx.Sign(path, nonce, body))

	request := &http.Client{}
	response, err := request.Do(config)
	if err != nil {
		return gjson.Result{}, err
	}
	defer response.Body.Close()

	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.ParseBytes(body), nil
}

func (bfx *Bitfinex) DepositAddress(method string) (gjson.Result, error) {
	if method == "" {
		method = "bitcoin"
	}

	data := map[string]interface{}{"method": method, "wallet": "exchange"}
	return bfx.Call("POST", "v2/auth/w/deposit/address", data)
}

func (bfx *Bitfinex) CreateInvoice(amount float64) (gjson.Result, error) {
	data := map[string]interface{}{"currency": "LNX", "wallet": "exchange", "amount": amount}
	return bfx.Call("POST", "v2/auth/w/deposit/invoice", data)
}

func (bfx *Bitfinex) GetWallets() (gjson.Result, error) {
	data := map[string]interface{}{}
	return bfx.Call("POST", "v2/auth/r/wallets", data)
}

func (bfx *Bitfinex) OrderSubmit(symbol string, amount float64) (gjson.Result, error) {
	data := map[string]interface{}{"symbol": symbol, "amount": amount, "type": "EXCHANGE"}
	return bfx.Call("POST", "v2/auth/w/order/submit", data)
}
