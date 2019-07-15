package ethutil

import (
	"context"
	"crypto/ecdsa"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var lastNonce uint64

// TrackTxResult ...
func TrackTxResult(clent *ethclient.Client, tx common.Hash) (bool, uint64, error) {
	for {
		rp, err := clent.TransactionReceipt(context.Background(), tx)
		if err != nil && err != ethereum.NotFound {
			return false, 0, err
		}
		if rp != nil {
			return rp.Status == types.ReceiptStatusSuccessful, rp.GasUsed, err
		}
		time.Sleep(time.Second * 3)
	}
}

// GenerateTransactOpts ...
func GenerateTransactOpts(client *ethclient.Client, pk *ecdsa.PrivateKey, addr common.Address) *bind.TransactOpts {
	nonce, err := client.PendingNonceAt(context.Background(), addr)
	if err != nil {
		log.Fatal("PendingNonceAt", err)
	}
	if nonce != 0 && nonce == lastNonce {
		nonce++
	}
	lastNonce = nonce
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal("SuggestGasPrice", err)
	}
	auth := bind.NewKeyedTransactor(pk)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)      // in wei
	auth.GasLimit = uint64(3000000) // in units
	auth.GasPrice = gasPrice
	return auth
}
