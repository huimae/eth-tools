package demo

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/naiba/tokenhunter/internal/erc20"
	"github.com/naiba/tokenhunter/internal/ethutil"
)

func getToken(pk, tokenAddr, network string, pending *bool, msg *string) {
	defer func() {
		*pending = false
	}()
	*msg = "正在为您准备代币，请稍候"
	*msg = "正在连接节点"
	client, err := ethclient.Dial(network)
	for err != nil {
		log.Println(err)
		time.Sleep(time.Second * 1)
		client, err = ethclient.Dial(network)
	}
	*msg = "连接节点成功"
	TBCAdminPk, err := crypto.HexToECDSA(pk)
	if err != nil {
		*msg = fmt.Sprint("获取失败", "HexToECDSA", err)
		return
	}
	publicKey := TBCAdminPk.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		*msg = fmt.Sprint("获取失败", "cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		return
	}
	TBCAdminPub := crypto.PubkeyToAddress(*publicKeyECDSA)
	token, err := erc20.NewErc20(common.HexToAddress(tokenAddr), client)
	if err != nil {
		*msg = fmt.Sprint("获取失败", "NewErc20", err)
		return
	}
	*msg = "正在获取代币"
	tx, err := token.AddToken(ethutil.GenerateTransactOpts(client, TBCAdminPk, TBCAdminPub), TBCAdminPub, big.NewInt(10000*1000))
	if err != nil {
		*msg = fmt.Sprint("获取失败", "AddToken", err)
		return
	}
	_, _, err = ethutil.TrackTxResult(client, tx.Hash())
	if err != nil {
		*msg = fmt.Sprint("获取失败", "TrackTxResult", err)
		return
	}
	*msg = "获取代币成功"
	tmp, err := token.BalanceOf(&bind.CallOpts{}, TBCAdminPub)
	if err != nil {
		*msg = fmt.Sprint("获取失败", "BalanceOf", err)
		return
	}
	*msg = fmt.Sprintf("恭喜您，Token 已发送到您的账户：%s，您当前余额为：%d。", TBCAdminPub.String(), tmp.Uint64()/1000)
}
