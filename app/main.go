package main

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"
	"strconv"

	"github.com/naiba/tokenhunter/internal/erc20"
	"github.com/naiba/tokenhunter/internal/ethutil"
	"github.com/naiba/tokenhunter/internal/uiutil"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	
	"github.com/andlabs/ui"
	_ "github.com/andlabs/ui/winmanifest"
)

var networks = []string{
	"Ropsten #wss://ropsten.infura.io/ws",
	"NBTestNet #ws://tokenbank.tk:7545/ws",
	"DBLTestNet #ws://120.55.15.98:9527/ws",
}

func setupUI() {
	mainwin := ui.NewWindow("Token 领取器", 400, 50, true)
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

	networkBox := ui.NewHorizontalBox()
	networkBox.SetPadded(true)
	networkLabel := ui.NewLabel("选择节点")
	networkCombo := ui.NewCombobox()
	for i := 0; i < len(networks); i++ {
		networkCombo.Append(networks[i])
	}
	networkBox.Append(networkLabel, false)
	networkBox.Append(networkCombo, true)

	tokenEntry, tokenBox := uiutil.GetEntry("代币地址")
	pkEntry, pkBox := uiutil.GetEntry("用户私钥")
	numEntry, numBox := uiutil.GetEntry("领取数量")

	getBox := ui.NewHorizontalBox()
	getBox.SetPadded(true)
	getBtn := ui.NewButton("获取代币")
	getBtn.OnClicked(func(b *ui.Button) {
		b.Disable()
		num,_:= strconv.ParseInt(numEntry.Text(),10,64)
		network := strings.Split(networks[networkCombo.Selected()], "#")
		go getToken(pkEntry.Text(), tokenEntry.Text(), network[1],num, b, mainwin)
	})

	getBox.Append(getBtn, true)


	mainBox.Append(networkBox, false)
	mainBox.Append(tokenBox, false)
	mainBox.Append(pkBox, false)
	mainBox.Append(numBox, false)
	mainBox.Append(getBox, false)
	mainwin.SetChild(mainBox)
	mainwin.Show()
}

func getToken(pk, tokenAddr, network string,num int64, btn *ui.Button, win *ui.Window) {
	setTitle := func(t string) {
		go ui.QueueMain(func() {
			win.SetTitle("Token 获取器：" + t)
		})
	}
	defer func() {
		go ui.QueueMain(func() {
			win.SetTitle("Token 获取器")
			btn.Enable()
		})
	}()
	setTitle("正在连接节点")
	client, err := ethclient.Dial(network)
	for err != nil {
		log.Println(err)
		time.Sleep(time.Second * 1)
		client, err = ethclient.Dial(network)
	}
	setTitle("连接节点成功，解析密钥")
	TBCAdminPk, err := crypto.HexToECDSA(pk)
	if err != nil {
		setTitle(fmt.Sprint("获取失败", "HexToECDSA", err))
		return
	}
	publicKey := TBCAdminPk.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		setTitle(fmt.Sprint("获取失败", "cannot assert type: publicKey is not of type *ecdsa.PublicKey"))
		return
	}
	TBCAdminPub := crypto.PubkeyToAddress(*publicKeyECDSA)
	token, err := erc20.NewErc20(common.HexToAddress(tokenAddr), client)
	if err != nil {
		setTitle(fmt.Sprint("获取失败", "NewErc20", err))
		return
	}
	setTitle("解析成功，正在获取代币")
	tx, err := token.AddToken(ethutil.GenerateTransactOpts(client, TBCAdminPk, TBCAdminPub), TBCAdminPub, big.NewInt(num*1000))
	if err != nil {
		setTitle(fmt.Sprint("获取失败", "AddToken", err))
		return
	}
	_, _, err = ethutil.TrackTxResult(client, tx.Hash())
	if err != nil {
		setTitle(fmt.Sprint("获取失败", "TrackTxResult", err))
		return
	}
	setTitle("获取代币成功")
	tmp, err := token.BalanceOf(&bind.CallOpts{}, TBCAdminPub)
	if err != nil {
		setTitle(fmt.Sprint("获取失败", "BalanceOf", err))
		return
	}
	setTitle(fmt.Sprintf("恭喜您，Token 已发送到您的账户：%s，您当前余额为：%d。", TBCAdminPub.String(), tmp.Uint64()/1000))
	time.Sleep(time.Second * 4)
}

func main() {
	ui.Main(setupUI)
}
