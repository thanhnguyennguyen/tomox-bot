package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/tomochain/tomochain/common"
	"github.com/tomochain/tomochain/crypto"
	"github.com/tomochain/tomochain/crypto/sha3"
	"github.com/tomochain/tomochain/rpc"
	"github.com/tomochain/tomochain/tomox/tomox_state"
	"math/big"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)
const (
	LendingStatusNew = "NEW"
	Borrowing = "BORROW"
	Investing = "INVEST"
	Limit = "LO"
	Market = "MO"
	LendingStatusCancelled = "CANCELLED"

)
type LendingOrderMsg struct {
	AccountNonce    uint64         `json:"nonce"    gencodec:"required"`
	Quantity        *big.Int       `json:"quantity,omitempty"`
	RelayerAddress  common.Address `json:"relayerAddress,omitempty"`
	UserAddress     common.Address `json:"userAddress,omitempty"`
	CollateralToken common.Address `json:"collateralToken,omitempty"`
	LendingToken    common.Address `json:"lendingToken,omitempty"`
	Interest        uint64         `json:"interest,omitempty"`
	Term            uint64         `json:"term,omitempty"`
	Status          string         `json:"status,omitempty"`
	Side            string         `json:"side,omitempty"`
	Type            string         `json:"type,omitempty"`
	LendingID       uint64         `json:"lendingID,omitempty"`
	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`

	// This is only used when marshaling to JSON.
	Hash common.Hash `json:"hash" rlp:"-"`
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
	//if len(os.Args) > 2 && os.Args[1] == "cancel" {
	//	orderId, _ := strconv.Atoi(os.Args[2])
	//	cancelOrder(rpcClient, big.NewInt(startNonce), uint64(orderId))
	//	return
	//}
	breakTime, _ := strconv.Atoi(os.Getenv("BREAK_TIME"))
	for {
		sendLending(rpcClient, uint64(startNonce))
		time.Sleep(time.Duration(int64(breakTime)) * time.Millisecond)
		startNonce++
	}
}


func EtherToWei(q *big.Int) *big.Int {
	return new(big.Int).Mul(q, new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
}

func sendLending(rpcClient *rpc.Client, nonce uint64) {
	rand.Seed(time.Now().UTC().UnixNano())
	item := &LendingOrderMsg{
		AccountNonce:    nonce,
		Quantity:        EtherToWei(big.NewInt(1000)),
		RelayerAddress: common.HexToAddress(os.Getenv("EXCHANGE_ADDRESS")), // "0x0D3ab14BBaD3D99F4203bd7a11aCB94882050E7e"
		UserAddress:     common.HexToAddress(os.Getenv("USER_ADDRESS")),
		CollateralToken: common.HexToAddress(os.Getenv("COLLATERAL_TOKEN")),
		LendingToken:    common.HexToAddress(os.Getenv("LENDING_TOKEN")),
		Interest:        uint64(100),
		Term:            uint64(30*86400),
		Status:          LendingStatusNew,
		Side:            Borrowing,
		Type:            Limit,
		V:               common.Big0,
		R:               common.Big0,
		S:               common.Big0,
		Hash:            common.Hash{},
	}
	hash := computeHash(item)
	if item.Status != LendingStatusCancelled {
		item.Hash = hash
	}
	privKey, _ := crypto.HexToECDSA(os.Getenv("PK"))
	message := crypto.Keccak256(
		[]byte("\x19Ethereum Signed Message:\n32"),
		hash.Bytes(),
	)
	signatureBytes, _ := crypto.Sign(message, privKey)
	sig := &tomox_state.Signature{
		R: common.BytesToHash(signatureBytes[0:32]),
		S: common.BytesToHash(signatureBytes[32:64]),
		V: signatureBytes[64] + 27,
	}
	item.R = sig.R.Big()
	item.S = sig.S.Big()
	item.V = new(big.Int).SetUint64(uint64(sig.V))

	var result interface{}

	err := rpcClient.Call(&result, "tomox_sendLending", item)
	fmt.Println("sendLendingitem", "nonce", item.AccountNonce)
	if err != nil {
		fmt.Println("rpcClient.Call tomox_sendLending failed", "err", err)
		os.Exit(1)
	}
}

func computeHash(l *LendingOrderMsg) common.Hash {
	sha := sha3.NewKeccak256()
	if l.Status == LendingStatusCancelled {
		sha := sha3.NewKeccak256()
		sha.Write(l.Hash.Bytes())
		sha.Write(common.BigToHash(big.NewInt(int64(l.AccountNonce))).Bytes())
		sha.Write(l.UserAddress.Bytes())
		sha.Write(common.BigToHash(big.NewInt(int64(l.LendingID))).Bytes())
		sha.Write([]byte(l.Status))
		sha.Write(l.RelayerAddress.Bytes())
	} else {
		sha.Write(l.RelayerAddress.Bytes())
		sha.Write(l.UserAddress.Bytes())
		sha.Write(l.CollateralToken.Bytes())
		sha.Write(l.LendingToken.Bytes())
		sha.Write(common.BigToHash(l.Quantity).Bytes())
		sha.Write(common.BigToHash(big.NewInt(int64(l.Term))).Bytes())
		if l.Type == Limit {
			sha.Write(common.BigToHash(big.NewInt(int64(l.Interest))).Bytes())
		}
		sha.Write([]byte(l.Side))
		sha.Write([]byte(l.Status))
		sha.Write([]byte(l.Type))
		sha.Write(common.BigToHash(big.NewInt(int64(l.AccountNonce))).Bytes())
	}
	return common.BytesToHash(sha.Sum(nil))

}

