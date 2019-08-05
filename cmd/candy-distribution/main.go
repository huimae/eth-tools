package main

import (
	"bufio"
	"crypto/ecdsa"
	"fmt"
	"math"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/andlabs/ui"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/naiba/eth-tools/internal/erc20"
	"github.com/naiba/eth-tools/internal/ethutil"
	"github.com/naiba/eth-tools/internal/uiutil"
)

var targetWalletsFile string
var targetWallets []common.Address
var logEntry *ui.MultilineEntry

func setupUI() {
	mainwin := ui.NewWindow("糖果分发器", 300, 418, true)
	mainwin.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
		mainwin.Destroy()
		return true
	})
	mainwin.SetMargined(true)
	mainBox := ui.NewVerticalBox()
	mainBox.SetPadded(true)

	pkEntry, pkBox := uiutil.GetEntry("钱包私钥")
	mainBox.Append(pkBox, false)

	tkEntry, tkBox := uiutil.GetEntry("代币地址")
	mainBox.Append(tkBox, false)

	amountEntry, amountBox := uiutil.GetEntry("分发数量")
	mainBox.Append(amountBox, false)

	dbBox := ui.NewHorizontalBox()
	dbLb := ui.NewLabel("请点击左侧按钮导入")
	dbBtn := ui.NewButton("导入用户钱包")
	dbBtn.OnClicked(func(b *ui.Button) {
		dbBtn.Disable()
		targetWalletsFile = ui.OpenFile(mainwin)
		dbLb.SetText(targetWalletsFile)
		go parseWallets(b)
	})
	dbBox.Append(dbBtn, false)
	dbBox.Append(dbLb, true)
	mainBox.Append(dbBox, false)

	doBtn := ui.NewButton("分发糖果")
	doBtn.OnClicked(func(b *ui.Button) {
		b.Disable()
		amount, _ := strconv.ParseInt(amountEntry.Text(), 10, 64)
		if amount <= 0 {
			b.Enable()
			appendLog("分发数量有误：" + amountEntry.Text())
			return
		}
		go distribution(b, pkEntry.Text(), tkEntry.Text(), amount)
	})
	mainBox.Append(doBtn, false)

	logEntry = ui.NewMultilineEntry()
	logEntry.SetReadOnly(true)
	mainBox.Append(logEntry, true)

	mainwin.SetChild(mainBox)
	mainwin.Show()
}

func appendLog(msg string) {
	ui.QueueMain(func() {
		logEntry.SetText(time.Now().Format("15:04:05") + "：" + msg + "\n" + logEntry.Text())
	})
}

func parseWallets(btn *ui.Button) {
	defer ui.QueueMain(func() {
		btn.Enable()
		appendLog(fmt.Sprintf("导入目标钱包完成，共导入 %d 个钱包地址", len(targetWallets)))
	})
	targetWallets = make([]common.Address, 0)
	file, err := os.OpenFile(targetWalletsFile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		appendLog(fmt.Sprintf("打开文件失败：%s,%s", targetWalletsFile, err))
		return
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		targetWallets = append(targetWallets, common.HexToAddress(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		appendLog(fmt.Sprintf("读取文件失败：%s,%s", targetWalletsFile, err))
		return
	}
}

func distribution(btn *ui.Button, pk, token string, amount int64) {
	defer btn.Enable()
	network := "wss://mainnet.infura.io/ws/v3/c520f3240b964adc94750241a96bd328"
	client, err := ethclient.Dial(network)
	for err != nil {
		appendLog(fmt.Sprintf("网络错误，正在重连：%s", err))
		time.Sleep(time.Second * 1)
		client, err = ethclient.Dial(network)
	}
	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		appendLog(fmt.Sprintf("解析私钥错误：%s", err))
		return
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		appendLog("解析公钥错误")
		return
	}
	wallet := crypto.PubkeyToAddress(*publicKeyECDSA)
	tokenAddr := common.HexToAddress(token)

	tokenContract, err := erc20.NewErc20(tokenAddr, client)
	if err != nil {
		appendLog(fmt.Sprintf("代币错误：%s", err))
		return
	}
	decimal, err := tokenContract.Decimals(&bind.CallOpts{
		From: wallet,
	})
	if err != nil {
		appendLog(fmt.Sprintf("获取代币小数位数错误：%s", err))
		return
	}
	bnAmount := big.NewInt(amount)
	bnAmount = bnAmount.Mul(bnAmount, big.NewInt(int64(math.Pow10(int(decimal)))))
	for i := 0; i < len(targetWallets); i++ {
		tx, err := tokenContract.Transfer(ethutil.GenerateTransactOpts(client, privateKey, wallet), targetWallets[i], bnAmount)
		if err != nil {
			appendLog(fmt.Sprintf("糖果分发错误：钱包-%s,Transaction-%s,Error-%s", targetWallets[i], tx.Hash().String(), err))
		} else {
			appendLog(fmt.Sprintf("糖果分发成功：钱包-%s,Transaction-%s,数量-%d", targetWallets[i], tx.Hash().String(), amount))
		}
	}
}

func main() {
	ui.Main(setupUI)
}
