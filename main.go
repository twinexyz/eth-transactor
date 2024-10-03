package main

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	DefaultGasLimit = 2500000
)

var (
	sender     string
	receiver   string
	privateKey string
	rpcUrl     string
	value      string
)

func init() {
	flag.StringVar(&sender, "sender", "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266", "sender address")
	flag.StringVar(&receiver, "receiver", "0x1CBd3b2770909D4e10f157cABC84C7264073C9Ec", "receiver address")
	flag.StringVar(&privateKey, "private-key", "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80", "sender private key")
	flag.StringVar(&rpcUrl, "rpc", "http://127.0.0.1:8545", "eth rpc url")
	flag.StringVar(&value, "value", "100000000000000000", "value")
}

func main() {
	flag.Parse()
	println(sender, receiver, privateKey, rpcUrl, value)
	ethRPC, err := rpc.Dial(rpcUrl)
	if err != nil {
		return
	}
	client := ethclient.NewClient(ethRPC)
	senderAddress := common.HexToAddress(sender)
	receiverAddress := common.HexToAddress(receiver)

	balance, err := client.BalanceAt(context.Background(), senderAddress, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("balance:: ", balance)

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	secretKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		panic(err)
	}

	newTransactOpts := func() (*bind.TransactOpts, error) {
		txo, err := bind.NewKeyedTransactorWithChainID(secretKey, chainID)
		if err != nil {
			return nil, err
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		txo.GasPrice, _ = client.SuggestGasPrice(ctx)
		txo.GasLimit = uint64(DefaultGasLimit)
		valueBigInt := new(big.Int)
		valueBigInt, ok := valueBigInt.SetString(value, 10)
		if !ok {
			panic("not ok")
		}
		txo.Value = valueBigInt
		return txo, nil
	}

	nonce, err := client.PendingNonceAt(context.Background(), senderAddress)
	if err != nil {
		panic(err)
	}

	txOpts, err := newTransactOpts()
	if err != nil {
		panic(err)
	}

	transferTxn := types.NewTransaction(nonce, receiverAddress, txOpts.Value, DefaultGasLimit, txOpts.GasPrice, []byte{})

	signedTxn, err := types.SignTx(transferTxn, types.NewEIP155Signer(chainID), secretKey)
	if err != nil {
		panic(err)
	}

	err = client.SendTransaction(context.Background(), signedTxn)
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Second * 3)
	fmt.Println("tx sent")
}
