package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/naiba/eth-debugger/internal/erc20"
	"github.com/naiba/eth-debugger/internal/ethutil"
	"github.com/naiba/eth-debugger/internal/uiutil"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/andlabs/ui"
	_ "github.com/andlabs/ui/winmanifest"
)

var networks = []string{
	"Kovan #wss://kovan.infura.io/ws",
	"Ropsten #wss://ropsten.infura.io/ws",
	"NBTestNet #ws://tokenbank.tk:7545/ws",
	"DBLTestNet #ws://120.55.15.98:9527/ws",
}

func setupUI() {
	mainwin := ui.NewWindow("ETH 调试工具", 300, 418, true)
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

	// ============= Choose Network =============
	networkBox := ui.NewHorizontalBox()
	networkBox.SetPadded(true)
	networkLabel := ui.NewLabel("选择节点")
	networkCombo := ui.NewCombobox()
	for i := 0; i < len(networks); i++ {
		networkCombo.Append(networks[i])
	}
	networkBox.Append(networkLabel, false)
	networkBox.Append(networkCombo, true)
	// ============= Set Token Info =============
	walletEntry, walletBox := uiutil.GetEntry("钱包地址")
	numEntry, numBox := uiutil.GetEntry("领取数量")

	// ============= Get Token =============
	getTokenBox := ui.NewVerticalBox()
	getTokenBox.SetPadded(true)
	tokenEntry, tokenBox := uiutil.GetEntry("代币地址")
	getBtn := ui.NewButton("获取代币")
	getBtn.OnClicked(func(b *ui.Button) {
		if !b.Enabled() {
			return
		}
		b.Disable()
		num, _ := strconv.ParseInt(numEntry.Text(), 10, 64)
		network := strings.Split(networks[networkCombo.Selected()], "#")
		go getToken(false, "01491231C2C71D99A16C7FB2120E185EAAE0861548B5FBF971859099DAB5DCA2", tokenEntry.Text(), walletEntry.Text(), network[1], num, b, mainwin)
	})
	getTokenBox.Append(tokenBox, true)
	getTokenBox.Append(getBtn, true)

	// ============= Get ETH =============
	getETHBox := ui.NewVerticalBox()
	getETHBox.SetPadded(true)

	qiongbiBtn := ui.NewButton("穷逼领钱")
	qiongbiBtn.OnClicked(func(b *ui.Button) {
		if !b.Enabled() {
			return
		}
		num, _ := strconv.ParseInt(numEntry.Text(), 10, 64)
		if num > 10 || num < 1 {
			ui.MsgBox(mainwin, "数量错误", "数量在 1 到 100 之间")
			return
		}
		b.Disable()
		network := strings.Split(networks[networkCombo.Selected()], "#")
		go getToken(true, "01491231C2C71D99A16C7FB2120E185EAAE0861548B5FBF971859099DAB5DCA2", tokenEntry.Text(), walletEntry.Text(), network[1], num, b, mainwin)
	})
	getETHBox.Append(qiongbiBtn, true)
	chargeBtn := ui.NewButton("充值ETH")
	chargeBtn.OnClicked(func(b *ui.Button) {
		if !b.Enabled() {
			return
		}
		b.Disable()
		go charge(b, mainwin)
	})
	getETHBox.Append(chargeBtn, true)

	// ============= A Tab =============
	mainTab := ui.NewTab()

	// ============= Tips Tab =============
	tipsLb := ui.NewMultilineEntry()
	tipsLb.SetReadOnly(true)
	tipsLb.SetText("前置操作：\n" +
		"1.填写你的钱包地址\n" +
		"2.填写需要领取的数量（领取代币无限制，ETH 限额为 1-100）\n" +
		"领取代币：\n" +
		"1.填写代币地址\n" +
		"领取ETH：\n" +
		"1.点击「穷逼领钱」按钮\n" +
		"2.如果领取不成功点击「充值ETH」充值库存",
	)

	mainBox.Append(networkBox, false) //选择网络
	mainBox.Append(walletBox, false)  // 钱包地址
	mainBox.Append(numBox, false)     // 设置数量
	mainTab.Append("获取代币", getTokenBox)
	mainTab.Append("获取ETH", getETHBox)
	mainBox.Append(mainTab, false) // 领取 ETH 或 代币
	mainBox.Append(tipsLb, true)   // 使用说明
	mainwin.SetChild(mainBox)
	mainwin.Show()
}

func getToken(isETH bool, pk, tokenAddr, walletAddr, network string, num int64, btn *ui.Button, win *ui.Window) {
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

	if isETH {
		if !strings.Contains(network, "ropsten") {
			setTitle("此网络无法进行充值")
			return
		}
		opt := ethutil.GenerateTransactOpts(client, TBCAdminPk, TBCAdminPub)
		opt.Value = big.NewInt(num * 10000000000000) // in wei (1 eth)
		var data []byte
		tx := types.NewTransaction(opt.Nonce.Uint64(), common.HexToAddress(walletAddr), opt.Value, opt.GasLimit, opt.GasPrice, data)

		chainID, err := client.NetworkID(context.Background())
		if err != nil {
			setTitle(fmt.Sprintf("获取失败 %s", err))
			return
		}

		setTitle("Transcation 签名")
		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), TBCAdminPk)
		if err != nil {
			setTitle(fmt.Sprintf("获取失败 %s", err))
			return
		}

		setTitle("广播 Transcation")
		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			setTitle(fmt.Sprintf("获取失败 %s", err))
			return
		}
		setTitle("发送成功")
	} else {
		token, err := erc20.NewErc20(common.HexToAddress(tokenAddr), client)
		if err != nil {
			setTitle(fmt.Sprint("获取失败", "NewErc20", err))
			return
		}
		setTitle("解析成功，正在获取代币")
		tx, err := token.AddToken(ethutil.GenerateTransactOpts(client, TBCAdminPk, TBCAdminPub), common.HexToAddress(walletAddr), big.NewInt(num*10000000000))
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
		setTitle(fmt.Sprintf("恭喜您，您当前余额为：%d。", tmp.Uint64()))
	}

	time.Sleep(time.Second * 4)
}

func charge(btn *ui.Button, win *ui.Window) {
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
	setTitle("正在充值")
	resp, err := http.Get("https://faucet.ropsten.be/donate/0x1f944B7F5aF34740541D438c75a93cD5200ef1c2")
	if err != nil || resp.StatusCode != 200 {
		setTitle("充值失败")
		return
	}
	setTitle("充值成功")
	time.Sleep(time.Second * 3)
}

func main() {
	ui.Main(setupUI)
}
