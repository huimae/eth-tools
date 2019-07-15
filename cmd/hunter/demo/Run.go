package demo

import (
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/inkyblackness/imgui-go"
)

// Platform covers mouse/keyboard/gamepad inputs, cursor shape, timing, windowing.
type Platform interface {
	// ShouldStop is regularly called as the abort condition for the program loop.
	ShouldStop() bool
	// ProcessEvents is called once per render loop to dispatch any pending events.
	ProcessEvents()
	// DisplaySize returns the dimension of the display.
	DisplaySize() [2]float32
	// FramebufferSize returns the dimension of the framebuffer.
	FramebufferSize() [2]float32
	// NewFrame marks the begin of a render pass. It must update the imgui IO state according to user input (mouse, keyboard, ...)
	NewFrame()
	// PostRender marks the completion of one render pass. Typically this causes the display buffer to be swapped.
	PostRender()
}

// Renderer covers rendering imgui draw data.
type Renderer interface {
	// PreRender causes the display buffer to be prepared for new output.
	PreRender(clearColor [4]float32)
	// Render draws the provided imgui draw data.
	Render(displaySize [2]float32, framebufferSize [2]float32, drawData imgui.DrawData)
}

// Run implements the main program loop of the demo. It returns when the platform signals to stop.
// This demo application shows some basic features of ImGui, as well as exposing the standard demo window.
func Run(p Platform, r Renderer) {
	backgroundColor := [4]float32{255.0, 255.0, 255.0, 1.0}
	networks := []string{
		"Ropsten#wss://ropsten.infura.io/ws)",
		"NBTestNet#ws://tokenbank.tk:7545/ws",
		"DBLTestNet#ws://120.55.15.98:9527/ws",
	}
	var network = networks[0]
	var privateKey, tokenAddr, pendingMsg string
	var demo, pending, showPending bool

	f, err := os.OpenFile("./data/config.txt", os.O_CREATE, 0655)
	if err == nil {
		b, err := ioutil.ReadAll(f)
		if err == nil {
			lines := strings.Split(string(b), "\n")
			if len(lines) > 1 {
				privateKey = lines[0]
				tokenAddr = lines[1]
			}
		}
	}

	for !p.ShouldStop() {
		p.ProcessEvents()
		p.NewFrame()
		imgui.NewFrame()

		if demo {
			imgui.ShowDemoWindow(&demo)
		}

		if showPending {
			imgui.OpenPopup("pendingPop")
		}

		if imgui.BeginPopupModalV("pendingPop", &showPending, imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoMove) {
			imgui.Text(pendingMsg)
			if !pending {
				if imgui.Button("关闭") {
					imgui.CloseCurrentPopup()
				}
			}
			imgui.EndPopup()
		}

		{
			imgui.Begin("TokenHunter")
			imgui.Checkbox("「展示更多组件」", &demo)
			imgui.Text("请选择测试节点，并将钱包私钥填入输入框")
			if imgui.BeginCombo("测试节点", network) {
				for i := 0; i < len(networks); i++ {
					isSelected := network == networks[i]
					if imgui.Selectable(networks[i]) {
						network = networks[i]
					}
					if isSelected {
						imgui.SetItemDefaultFocus()
					}
				}
				imgui.EndCombo()
			}
			imgui.InputText("代币地址", &tokenAddr)
			imgui.InputText("钱包私钥", &privateKey)
			if imgui.Button("领取 Token") {
				showPending = true
				pending = true
				networkSplit := strings.Split(network, "#")
				go getToken(privateKey, tokenAddr, networkSplit[1], &pending, &pendingMsg)
			}
			imgui.End()
		}

		imgui.Render()
		r.PreRender(backgroundColor)
		r.Render(p.DisplaySize(), p.FramebufferSize(), imgui.RenderedDrawData())
		p.PostRender()
		<-time.After(time.Millisecond * 25)
	}
}
