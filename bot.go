package main

// using go get
import (
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tomochain/tomochain/crypto"
	"github.com/tomochain/tomochain/crypto/sha3"
	"github.com/tomochain/tomochain/tomox/tradingstate"

	"github.com/joho/godotenv"
	"github.com/tomochain/tomochain/common"
	"github.com/tomochain/tomochain/rpc"
)

const (
	NEW_ORDER       = "NEW"
	CANCELLED_ORDER = "CANCELLED"
)

type OrderMsg struct {
	AccountNonce    string `json:"nonce"    gencodec:"required"`
	Quantity        string    `json:"quantity,omitempty"`
	Price           string    `json:"price,omitempty"`
	ExchangeAddress common.Address `json:"exchangeAddress,omitempty"`
	UserAddress     common.Address `json:"userAddress,omitempty"`
	BaseToken       common.Address `json:"baseToken,omitempty"`
	QuoteToken      common.Address `json:"quoteToken,omitempty"`
	Status          string         `json:"status,omitempty"`
	Side            string         `json:"side,omitempty"`
	Type            string         `json:"type,omitempty"`
	PairName        string         `json:"pairName,omitempty"`
	OrderID         string `json:"orderid,omitempty"`
	// Signature values
	V string `json:"v" gencodec:"required"`
	R string `json:"r" gencodec:"required"`
	S string `json:"s" gencodec:"required"`

	// This is only used when marshaling to JSON.
	Hash common.Hash `json:"hash"`
}

func buildOrder(nonce *big.Int, isCancel bool) *tradingstate.OrderItem {
	b, err := strconv.Atoi(os.Getenv("BASE_DECIMAL"))
	if err != nil {
		panic(fmt.Errorf("fail to get BASE_DECIMAL . Err: %v", err))
	}
	baseDecimal := new(big.Int).SetUint64(0).Exp(big.NewInt(10), big.NewInt(int64(b)), nil)
	q, err := strconv.Atoi(os.Getenv("QUOTE_DECIMAL"))
	if err != nil {
		panic(fmt.Errorf("fail to get QUOTE_DECIMAL . Err: %v", err))
	}
	quoteDecimal := new(big.Int).SetUint64(0).Exp(big.NewInt(10), big.NewInt(int64(q)), nil)
	rand.Seed(time.Now().UTC().UnixNano())
	lstBuySell := []string{"BUY", "SELL"}
	coingectkoPrice, _ := getPrice(os.Getenv("COINGECKO_PRICE_BASE_ID"), os.Getenv("COINGECKO_PRICE_QUOTE_ID"))
	price := coingectkoPrice
	if inverse := os.Getenv("PRICE_INVERSE"); strings.ToLower(inverse) == "yes" || strings.ToLower(inverse) == "true" {
		price = 1 / coingectkoPrice
	}
	quan, _ := strconv.Atoi(os.Getenv("QUANTITY_DECIMAL"))
	quantityDecimal := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(quan)), big.NewInt(0))
	p, _ := strconv.Atoi(os.Getenv("PRICE_DECIMAL"))
	priceDecimal := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(p)), big.NewInt(0))
	randomPrice := int64((1 + float64(rand.Intn(10))/1000) * float64(price) * float64(priceDecimal.Int64())) // 0 - 1% real price
	pricepoint := big.NewInt(0).Div(big.NewInt(0).Mul(big.NewInt(randomPrice), quoteDecimal), priceDecimal)
	order := &tradingstate.OrderItem{
		Quantity:        big.NewInt(0).Div(big.NewInt(0).Mul(big.NewInt(int64(rand.Intn(20)+2)), baseDecimal), quantityDecimal),
		Price:           pricepoint,
		ExchangeAddress: common.HexToAddress(os.Getenv("EXCHANGE_ADDRESS")), // "0x0D3ab14BBaD3D99F4203bd7a11aCB94882050E7e"
		UserAddress:     common.HexToAddress(os.Getenv("USER_ADDRESS")),
		BaseToken:       common.HexToAddress(os.Getenv("BASE_TOKEN")),  // 0x4d7eA2cE949216D6b120f3AA10164173615A2b6C
		QuoteToken:      common.HexToAddress(os.Getenv("QUOTE_TOKEN")), // common.TomoNativeAddress
		Side:            lstBuySell[rand.Int()%len(lstBuySell)],
		Type:            tradingstate.Limit,
		FilledAmount:    new(big.Int).SetUint64(0),
		Nonce:           nonce,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if isCancel {
		order.Status = CANCELLED_ORDER
	} else {
		order.Status = NEW_ORDER
		fmt.Printf("Price %0.8f . Quantity: %0.8f . Nonce: %d . Side: %s", float64(new(big.Int).Div(new(big.Int).Mul(order.Price, priceDecimal), quoteDecimal).Uint64())/float64(priceDecimal.Uint64()), float64(new(big.Int).Div(new(big.Int).Mul(order.Quantity, quantityDecimal), baseDecimal).Uint64())/float64(quantityDecimal.Uint64()), order.Nonce.Uint64(), order.Side)
	}
	fmt.Println()
	return order
}

func sendOrder(rpcClient *rpc.Client, nonce *big.Int) {
	order := buildOrder(nonce, false)
	order.Hash = ComputeHash(order)

	privKey, _ := crypto.HexToECDSA(os.Getenv("PK"))
	message := crypto.Keccak256(
		[]byte("\x19Ethereum Signed Message:\n32"),
		order.Hash.Bytes(),
	)
	signatureBytes, _ := crypto.Sign(message, privKey)

	orderMsg := OrderMsg{
		AccountNonce:    "0x" + order.Nonce.Text(16),
		Quantity:        "0x" + order.Quantity.Text(16),
		Price:           "0x" + order.Price.Text(16),
		ExchangeAddress: order.ExchangeAddress,
		UserAddress:     order.UserAddress,
		BaseToken:       order.BaseToken,
		QuoteToken:      order.QuoteToken,
		Status:          order.Status,
		Hash:            order.Hash,
		Side:            order.Side,
		Type:            order.Type,
		V:               "0x" + new(big.Int).SetUint64(uint64(signatureBytes[64] + 27)).Text(16),
		R:               "0x" + new(big.Int).SetBytes(signatureBytes[0:32]).Text(16),
		S:               "0x" + new(big.Int).SetBytes(signatureBytes[32:64]).Text(16),
	}
	var result interface{}

	err := rpcClient.Call(&result, "tomox_sendOrder", orderMsg)
	if err != nil {
		fmt.Println("rpcClient.Call tomox_sendOrder failed", "err", err)
		os.Exit(1)
	}
}

func cancelOrder(rpcClient *rpc.Client, nonce *big.Int, orderId uint64) {
	order := buildOrder(nonce, true)
	order.Status = CANCELLED_ORDER
	order.OrderID = orderId
	baseToken := os.Getenv("BASE_TOKEN")
	quoteToken := os.Getenv("QUOTE_TOKEN")
	// getOrderById
	var res interface{}
	err := rpcClient.Call(&res, "tomox_getOrderById", baseToken, quoteToken, orderId)
	if err != nil {
		fmt.Println("cancel order: tomox_getOrderById failed", "err", err)
		os.Exit(1)
	}
	originOrder := res.(map[string]interface{})
	hash := common.HexToHash(originOrder["hash"].(string))
	order.Hash = hash
	fmt.Printf("Cancel order: OrderId: %d . OrderHash: %s .", orderId, hash.Hex())
	fmt.Println()
	newHash := ComputeHash(order)

	privKey, _ := crypto.HexToECDSA(os.Getenv("PK"))
	message := crypto.Keccak256(
		[]byte("\x19Ethereum Signed Message:\n32"),
		newHash.Bytes(),
	)
	signatureBytes, _ := crypto.Sign(message, privKey)

	orderMsg := OrderMsg{
		AccountNonce:    "0x" + order.Nonce.Text(16),
		ExchangeAddress: order.ExchangeAddress,
		UserAddress:     order.UserAddress,
		BaseToken:       order.BaseToken,
		QuoteToken:      order.QuoteToken,
		Status:          CANCELLED_ORDER,
		Hash:            hash,
		Side:            order.Side,
		V:               "0x" + new(big.Int).SetUint64(uint64(signatureBytes[64] + 27)).Text(16),
		R:               "0x" + new(big.Int).SetBytes(signatureBytes[0:32]).Text(16),
		S:               "0x" + new(big.Int).SetBytes(signatureBytes[32:64]).Text(16),
		OrderID:         "0x" + new(big.Int).SetUint64(orderId).Text(16),
	}
	var result interface{}

	err = rpcClient.Call(&result, "tomox_sendOrder", orderMsg)
	if err != nil {
		fmt.Println("cancel tomox_sendOrder failed", "err", err)
		os.Exit(1)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	rpcClient, err := rpc.DialHTTP(os.Getenv("RPC_ENDPOINT"))
	defer rpcClient.Close()
	if err != nil {
		fmt.Println("rpc.DialHTTP failed", "err", err)
		os.Exit(1)
	}
	var result interface{}
	err = rpcClient.Call(&result, "tomox_getOrderCount", os.Getenv("USER_ADDRESS"))
	if err != nil {
		fmt.Println("rpcClient.Call tomox_getOrderCount failed", "err", err)
		os.Exit(1)
	}
	startNonce, _ := strconv.ParseInt(strings.TrimLeft(result.(string), "0x"), 16, 64)
	fmt.Println("Your current orderNonce: ", startNonce)

	// cancel order
	// param 1: string "cancel"
	// param 2: uint64 orderId
	if len(os.Args) > 2 && os.Args[1] == "cancel" {
		orderId, _ := strconv.Atoi(os.Args[2])
		cancelOrder(rpcClient, big.NewInt(startNonce), uint64(orderId))
		return
	}
	breakTime, _ := strconv.Atoi(os.Getenv("BREAK_TIME"))
	for {
		sendOrder(rpcClient, new(big.Int).SetUint64(uint64(startNonce)))
		time.Sleep(time.Duration(int64(breakTime)) * time.Millisecond)
		startNonce++
	}
}

func getPrice(base, quote string) (float32, error) {
	resp, err := http.Get("https://api.coingecko.com/api/v3/simple/price?ids=" + base + "&vs_currencies=" + quote)
	if err != nil {
		return float32(0), fmt.Errorf(err.Error())
	}
	var data map[string]map[string]float32
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return float32(0), fmt.Errorf(err.Error())
	}
	return data[base][quote], nil
}

func ComputeHash(item *tradingstate.OrderItem) common.Hash {
	sha := sha3.NewKeccak256()
	if item.Status == CANCELLED_ORDER {
		sha.Write(item.Hash.Bytes())
		sha.Write(common.BigToHash(big.NewInt(int64(item.Nonce.Uint64()))).Bytes())
		sha.Write(item.UserAddress.Bytes())
		sha.Write(common.BigToHash(big.NewInt(int64(item.OrderID))).Bytes())
		sha.Write([]byte(item.Status))
		sha.Write(item.ExchangeAddress.Bytes())
		sha.Write(item.BaseToken.Bytes())
		sha.Write(item.QuoteToken.Bytes())
	} else {
		sha.Write(item.ExchangeAddress.Bytes())
		sha.Write(item.UserAddress.Bytes())
		sha.Write(item.BaseToken.Bytes())
		sha.Write(item.QuoteToken.Bytes())
		sha.Write(common.BigToHash(item.Quantity).Bytes())
		if item.Price != nil {
			sha.Write(common.BigToHash(item.Price).Bytes())
		}
		if item.Side == tradingstate.Bid {
			sha.Write(common.BigToHash(big.NewInt(0)).Bytes())
		} else {
			sha.Write(common.BigToHash(big.NewInt(1)).Bytes())
		}
		sha.Write([]byte(item.Status))
		sha.Write([]byte(item.Type))
		sha.Write(common.BigToHash(item.Nonce).Bytes())
	}
	return common.BytesToHash(sha.Sum(nil))
}
