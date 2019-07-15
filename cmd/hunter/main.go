// +build sdl

package main

import (
	"fmt"
	"os"

	"github.com/inkyblackness/imgui-go"

	"github.com/naiba/tokenhunter/cmd/hunter/demo"
	"github.com/naiba/tokenhunter/internal/platforms"
	"github.com/naiba/tokenhunter/internal/renderers"
)

func main() {
	context := imgui.CreateContext(nil)
	defer context.Destroy()
	io := imgui.CurrentIO()

	platform, err := platforms.NewSDL(io, platforms.SDLClientAPIOpenGL3)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	defer platform.Dispose()

	renderer, err := renderers.NewOpenGL3(io)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	defer renderer.Dispose()

	demo.Run(platform, renderer)
}
