package main

// using go get
import (
	"encoding/json"
	"fmt"
	"github.com/tomochain/tomochain/crypto"
	"github.com/tomochain/tomochain/crypto/sha3"
	"github.com/tomochain/tomochain/tomox"
	"github.com/tomochain/tomochain/tomox/tomox_state"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/tomochain/tomochain/common"
	"github.com/tomochain/tomochain/rpc"
)
const RPC_ENDPOINT = "http://127.0.0.1:1545"

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

func buildOrder(userAddr string, nonce *big.Int) *tomox_state.OrderItem {
	var ether = big.NewInt(1000000000000000000)
	var bitUnit = big.NewInt(1000000)
	rand.Seed(time.Now().UTC().UnixNano())
	lstBuySell := []string{"BUY", "SELL"}
	tomoPrice, _ := getPrice("tomochain", "btc")
	btcPrice := int(1 / tomoPrice)
	order := &tomox_state.OrderItem{
		Quantity:        big.NewInt(0).Mul(big.NewInt(int64(rand.Intn(10)+1)), bitUnit),
		Price:           big.NewInt(0).Mul(big.NewInt(int64(rand.Intn(100)+btcPrice)), ether),
		ExchangeAddress: common.HexToAddress("0x0D3ab14BBaD3D99F4203bd7a11aCB94882050E7e"),
		UserAddress:     common.HexToAddress(userAddr),
		BaseToken:       common.HexToAddress("0x4d7eA2cE949216D6b120f3AA10164173615A2b6C"),
		QuoteToken:      common.HexToAddress(common.TomoNativeAddress),
		Status:          tomox.OrderStatusNew,
		Side:            lstBuySell[rand.Int()%len(lstBuySell)],
		Type:            tomox_state.Limit,
		PairName:        "BTC/TOMO",
		FilledAmount:    new(big.Int).SetUint64(0),
		Nonce:           nonce,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	fmt.Printf("price %v  . Side: %s . Pair: %s . ToExchange: %s . ", order.Price, order.Side, order.PairName, order.ExchangeAddress.Hex())
	return order
}

func testCreateOrder(userAddr, pk string, nonce *big.Int) {
	order := buildOrder(userAddr, nonce)
	order.Hash = ComputeHash(order)

	privKey, _ := crypto.HexToECDSA(pk)
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

	fmt.Println("nonce: ", nonce.Uint64())


	//create topic
	rpcClient, err := rpc.DialHTTP(RPC_ENDPOINT)
	defer rpcClient.Close()
	if err != nil {
		fmt.Println("rpc.DialHTTP failed", "err", err)
		os.Exit(1)
	}

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


	err = rpcClient.Call(&result, "tomox_sendOrder", orderMsg)
	if err != nil {
		fmt.Println("rpcClient.Call tomox_sendOrder failed", "err", err)
		os.Exit(1)
	}
}

func main() {
	userAddr := os.Args[1]
	pk := os.Args[2]
	startNonce, _ := strconv.Atoi(os.Args[3])
	breakTime, _ := strconv.Atoi(os.Args[4])
	for {
		testCreateOrder(userAddr, pk, new(big.Int).SetUint64(uint64(startNonce)))
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
	return common.BytesToHash(sha.Sum(nil))
}
