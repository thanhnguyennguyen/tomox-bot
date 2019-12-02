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
	"github.com/tomochain/tomochain/tomox"
	"github.com/tomochain/tomochain/tomox/tomox_state"

	"github.com/joho/godotenv"
	"github.com/tomochain/tomochain/common"
	"github.com/tomochain/tomochain/rpc"
)

type OrderMsg struct {
	AccountNonce    uint64         `json:"nonce"    gencodec:"required"`
	Quantity        *big.Int       `json:"quantity,omitempty"`
	Price           *big.Int       `json:"price,omitempty"`
	ExchangeAddress common.Address `json:"exchangeAddress,omitempty"`
	UserAddress     common.Address `json:"userAddress,omitempty"`
	BaseToken       common.Address `json:"baseToken,omitempty"`
	QuoteToken      common.Address `json:"quoteToken,omitempty"`
	Status          string         `json:"status,omitempty"`
	Side            string         `json:"side,omitempty"`
	Type            string         `json:"type,omitempty"`
	PairName        string         `json:"pairName,omitempty"`
	OrderID         uint64         `json:"orderid,omitempty"`
	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`

	// This is only used when marshaling to JSON.
	Hash common.Hash `json:"hash"`
}

func buildOrder(nonce *big.Int, isCancel bool) *tomox_state.OrderItem {
	baseDecimal, err := strconv.Atoi(os.Getenv("BASE_DECIMAL"))
	if err != nil {
		panic(fmt.Errorf("fail to get BASE_DECIMAL . Err: %v", err))
	}
	quantityDecimal := new(big.Int).SetUint64(0).Exp(big.NewInt(10), big.NewInt(int64(baseDecimal)), nil)
	quoteDecimal, err := strconv.Atoi(os.Getenv("QUOTE_DECIMAL"))
	if err != nil {
		panic(fmt.Errorf("fail to get QUOTE_DECIMAL . Err: %v", err))
	}
	priceDecimal := new(big.Int).SetUint64(0).Exp(big.NewInt(10), big.NewInt(int64(quoteDecimal)), nil)
	rand.Seed(time.Now().UTC().UnixNano())
	lstBuySell := []string{"BUY", "SELL"}
	coingectkoPrice, _ := getPrice(os.Getenv("COINGECKO_PRICE_BASE_ID"), os.Getenv("COINGECKO_PRICE_QUOTE_ID"))
	price := coingectkoPrice
	if inverse := os.Getenv("PRICE_INVERSE"); strings.ToLower(inverse) == "yes" || strings.ToLower(inverse) == "true" {
		price = 1 / coingectkoPrice
	}
	randomPriceWithSixDecimal := int64((1 + float64(rand.Intn(10))/1000)*float64(price)*1000000) // 0 - 1% real price
	pricepoint := big.NewInt(0).Div(big.NewInt(0).Mul(big.NewInt(randomPriceWithSixDecimal), priceDecimal), big.NewInt(1000000))
	order := &tomox_state.OrderItem{
		Quantity:        big.NewInt(0).Div(big.NewInt(0).Mul(big.NewInt(int64(rand.Intn(10)+1)), quantityDecimal), big.NewInt(100)),
		Price:           pricepoint,
		ExchangeAddress: common.HexToAddress(os.Getenv("EXCHANGE_ADDRESS")), // "0x0D3ab14BBaD3D99F4203bd7a11aCB94882050E7e"
		UserAddress:     common.HexToAddress(os.Getenv("USER_ADDRESS")),
		BaseToken:       common.HexToAddress(os.Getenv("BASE_TOKEN")),  // 0x4d7eA2cE949216D6b120f3AA10164173615A2b6C
		QuoteToken:      common.HexToAddress(os.Getenv("QUOTE_TOKEN")), // common.TomoNativeAddress
		Side:            lstBuySell[rand.Int()%len(lstBuySell)],
		Type:            tomox_state.Limit,
		PairName:        os.Getenv("PAIR_NAME"),
		FilledAmount:    new(big.Int).SetUint64(0),
		Nonce:           nonce,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if isCancel {
		order.Status = tomox.OrderStatusCancelled
	} else {
		order.Status = tomox.OrderStatusNew
		fmt.Printf("Pair: %s . Price %0.6f . Quantity: %0.2f . Nonce: %d . Side: %s", order.PairName, float64(new(big.Int).Div(new(big.Int).Mul(order.Price, big.NewInt(1000000)), priceDecimal).Uint64()) / 1000000, float64(new(big.Int).Div(new(big.Int).Mul(order.Quantity, big.NewInt(100)), quantityDecimal).Uint64()) / 100, order.Nonce.Uint64(), order.Side)
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
	sig := &tomox_state.Signature{
		R: common.BytesToHash(signatureBytes[0:32]),
		S: common.BytesToHash(signatureBytes[32:64]),
		V: signatureBytes[64] + 27,
	}
	order.Signature = sig

	orderMsg := OrderMsg{
		AccountNonce:    order.Nonce.Uint64(),
		Quantity:        order.Quantity,
		Price:           order.Price,
		ExchangeAddress: order.ExchangeAddress,
		UserAddress:     order.UserAddress,
		BaseToken:       order.BaseToken,
		QuoteToken:      order.QuoteToken,
		Status:          order.Status,
		Side:            order.Side,
		Type:            order.Type,
		PairName:        order.PairName,
		V:               new(big.Int).SetUint64(uint64(signatureBytes[64] + 27)),
		R:               new(big.Int).SetBytes(signatureBytes[0:32]),
		S:               new(big.Int).SetBytes(signatureBytes[32:64]),
	}
	var result interface{}

	err := rpcClient.Call(&result, "tomox_sendOrder", orderMsg)
	if err != nil {
		fmt.Println("rpcClient.Call tomox_sendOrder failed", "err", err)
		os.Exit(1)
	}
}

func cancelOrder(rpcClient *rpc.Client, nonce *big.Int, orderId uint64, hash common.Hash) {
	order := buildOrder(nonce, true)
	order.Status = tomox.OrderStatusCancelled
	order.OrderID = orderId
	order.Hash = hash
	fmt.Printf("Cancel order: OrderId: %d . OrderHash: %s .", orderId, hash.Hex())
	newHash := ComputeHash(order)

	privKey, _ := crypto.HexToECDSA(os.Getenv("PK"))
	message := crypto.Keccak256(
		[]byte("\x19Ethereum Signed Message:\n32"),
		newHash.Bytes(),
	)
	signatureBytes, _ := crypto.Sign(message, privKey)
	sig := &tomox_state.Signature{
		R: common.BytesToHash(signatureBytes[0:32]),
		S: common.BytesToHash(signatureBytes[32:64]),
		V: signatureBytes[64] + 27,
	}
	order.Signature = sig

	orderMsg := OrderMsg{
		AccountNonce:    order.Nonce.Uint64(),
		Quantity:        order.Quantity,
		Price:           order.Price,
		ExchangeAddress: order.ExchangeAddress,
		UserAddress:     order.UserAddress,
		BaseToken:       order.BaseToken,
		QuoteToken:      order.QuoteToken,
		Status:          tomox.OrderStatusCancelled,
		Side:            order.Side,
		Type:            order.Type,
		PairName:        order.PairName,
		V:               new(big.Int).SetUint64(uint64(signatureBytes[64] + 27)),
		R:               new(big.Int).SetBytes(signatureBytes[0:32]),
		S:               new(big.Int).SetBytes(signatureBytes[32:64]),
	}
	var result interface{}

	err := rpcClient.Call(&result, "tomox_sendOrder", orderMsg)
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
	// param 3: hash
	if len(os.Args) == 4 && os.Args[1] == "cancel" {
		orderId, _ := strconv.Atoi(os.Args[2])
		cancelOrder(rpcClient, big.NewInt(startNonce), uint64(orderId), common.HexToHash(os.Args[3]))
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

func ComputeHash(item *tomox_state.OrderItem) common.Hash {
	sha := sha3.NewKeccak256()
	if item.Status == tomox.OrderStatusCancelled {
		sha.Write(item.Hash.Bytes())
		sha.Write(common.BigToHash(big.NewInt(int64(item.Nonce.Uint64()))).Bytes())
		sha.Write(item.UserAddress.Bytes())
		sha.Write(common.BigToHash(big.NewInt(int64(item.OrderID))).Bytes())
		sha.Write([]byte(item.Status))
		sha.Write(item.ExchangeAddress.Bytes())
	} else {
		sha.Write(item.ExchangeAddress.Bytes())
		sha.Write(item.UserAddress.Bytes())
		sha.Write(item.BaseToken.Bytes())
		sha.Write(item.QuoteToken.Bytes())
		sha.Write(common.BigToHash(item.Quantity).Bytes())
		if item.Price != nil {
			sha.Write(common.BigToHash(item.Price).Bytes())
		}
		if item.Side == tomox_state.Bid {
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
