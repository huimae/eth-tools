// +build sdl

package platforms

import (
	"fmt"
	"runtime"

	"github.com/inkyblackness/imgui-go"
	"github.com/veandco/go-sdl2/sdl"
)

// SDLClientAPI identifies the render system that shall be initialized.
type SDLClientAPI string

// SDLClientAPI constants
const (
	SDLClientAPIOpenGL2 SDLClientAPI = "OpenGL2"
	SDLClientAPIOpenGL3 SDLClientAPI = "OpenGL3"
)

// SDL implements a platform based on github.com/veandco/go-sdl2 (v2).
type SDL struct {
	imguiIO imgui.IO

	window     *sdl.Window
	shouldStop bool

	time        uint64
	buttonsDown [3]bool
}

// NewSDL attempts to initialize an SDL context.
func NewSDL(io imgui.IO, clientAPI SDLClientAPI) (*SDL, error) {
	runtime.LockOSThread()

	err := sdl.Init(sdl.INIT_VIDEO)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SDL2: %v", err)
	}

	window, err := sdl.CreateWindow("ImGui-Go SDL2+"+string(clientAPI)+" example", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, 1280, 720, sdl.WINDOW_OPENGL)
	if err != nil {
		sdl.Quit()
		return nil, fmt.Errorf("failed to create window: %v", err)
	}

	platform := &SDL{
		imguiIO: io,
		window:  window,
	}
	platform.setKeyMapping()

	switch clientAPI {
	case SDLClientAPIOpenGL2:
		_ = sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 2)
		_ = sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 1)
	case SDLClientAPIOpenGL3:
		_ = sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 3)
		_ = sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 2)
		_ = sdl.GLSetAttribute(sdl.GL_CONTEXT_FLAGS, sdl.GL_CONTEXT_FORWARD_COMPATIBLE_FLAG)
		_ = sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)
	default:
		platform.Dispose()
		return nil, fmt.Errorf("unsupported ClientAPI: <%s>", clientAPI)
	}
	_ = sdl.GLSetAttribute(sdl.GL_DOUBLEBUFFER, 1)
	_ = sdl.GLSetAttribute(sdl.GL_DEPTH_SIZE, 24)
	_ = sdl.GLSetAttribute(sdl.GL_STENCIL_SIZE, 8)

	glContext, err := window.GLCreateContext()
	if err != nil {
		platform.Dispose()
		return nil, fmt.Errorf("failed to create OpenGL context: %v", err)
	}
	err = window.GLMakeCurrent(glContext)
	if err != nil {
		platform.Dispose()
		return nil, fmt.Errorf("failed to set current OpenGL context: %v", err)
	}

	_ = sdl.GLSetSwapInterval(1)

	return platform, nil
}

// Dispose cleans up the resources.
func (platform *SDL) Dispose() {
	if platform.window != nil {
		_ = platform.window.Destroy()
		platform.window = nil
	}
	sdl.Quit()
}

// ShouldStop returns true if the window is to be closed.
func (platform *SDL) ShouldStop() bool {
	return platform.shouldStop
}

// ProcessEvents handles all pending window events.
func (platform *SDL) ProcessEvents() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		platform.processEvent(event)
	}
}

// DisplaySize returns the dimension of the display.
func (platform *SDL) DisplaySize() [2]float32 {
	w, h := platform.window.GetSize()
	return [2]float32{float32(w), float32(h)}
}

// FramebufferSize returns the dimension of the framebuffer.
func (platform *SDL) FramebufferSize() [2]float32 {
	w, h := platform.window.GLGetDrawableSize()
	return [2]float32{float32(w), float32(h)}
}

// NewFrame marks the begin of a render pass. It forwards all current state to imgui.CurrentIO().
func (platform *SDL) NewFrame() {
	// Setup display size (every frame to accommodate for window resizing)
	displaySize := platform.DisplaySize()
	platform.imguiIO.SetDisplaySize(imgui.Vec2{X: displaySize[0], Y: displaySize[1]})

	// Setup time step (we don't use SDL_GetTicks() because it is using millisecond resolution)
	frequency := sdl.GetPerformanceFrequency()
	currentTime := sdl.GetPerformanceCounter()
	if platform.time > 0 {
		platform.imguiIO.SetDeltaTime(float32(currentTime-platform.time) / float32(frequency))
	} else {
		platform.imguiIO.SetDeltaTime(1.0 / 60.0)
	}
	platform.time = currentTime

	// If a mouse press event came, always pass it as "mouse held this frame", so we don't miss click-release events that are shorter than 1 frame.
	x, y, state := sdl.GetMouseState()
	platform.imguiIO.SetMousePosition(imgui.Vec2{X: float32(x), Y: float32(y)})
	for i, button := range []uint32{sdl.BUTTON_LEFT, sdl.BUTTON_RIGHT, sdl.BUTTON_MIDDLE} {
		platform.imguiIO.SetMouseButtonDown(i, platform.buttonsDown[i] || (state&sdl.Button(button)) != 0)
		platform.buttonsDown[i] = false
	}
}

// PostRender performs a buffer swap.
func (platform *SDL) PostRender() {
	platform.window.GLSwap()
}

func (platform *SDL) setKeyMapping() {
	keys := map[int]int{
		imgui.KeyTab:        sdl.SCANCODE_TAB,
		imgui.KeyLeftArrow:  sdl.SCANCODE_LEFT,
		imgui.KeyRightArrow: sdl.SCANCODE_RIGHT,
		imgui.KeyUpArrow:    sdl.SCANCODE_UP,
		imgui.KeyDownArrow:  sdl.SCANCODE_DOWN,
		imgui.KeyPageUp:     sdl.SCANCODE_PAGEUP,
		imgui.KeyPageDown:   sdl.SCANCODE_PAGEDOWN,
		imgui.KeyHome:       sdl.SCANCODE_HOME,
		imgui.KeyEnd:        sdl.SCANCODE_END,
		imgui.KeyInsert:     sdl.SCANCODE_INSERT,
		imgui.KeyDelete:     sdl.SCANCODE_DELETE,
		imgui.KeyBackspace:  sdl.SCANCODE_BACKSPACE,
		imgui.KeySpace:      sdl.SCANCODE_BACKSPACE,
		imgui.KeyEnter:      sdl.SCANCODE_RETURN,
		imgui.KeyEscape:     sdl.SCANCODE_ESCAPE,
		imgui.KeyA:          sdl.SCANCODE_A,
		imgui.KeyC:          sdl.SCANCODE_C,
		imgui.KeyV:          sdl.SCANCODE_V,
		imgui.KeyX:          sdl.SCANCODE_X,
		imgui.KeyY:          sdl.SCANCODE_Y,
		imgui.KeyZ:          sdl.SCANCODE_Z,
	}

	// Keyboard mapping. ImGui will use those indices to peek into the io.KeysDown[] array.
	for imguiKey, nativeKey := range keys {
		platform.imguiIO.KeyMap(imguiKey, nativeKey)
	}
}

func (platform *SDL) processEvent(event sdl.Event) {
	switch event.GetType() {
	case sdl.QUIT:
		platform.shouldStop = true
	case sdl.MOUSEWHEEL:
		wheelEvent := event.(*sdl.MouseWheelEvent)
		var deltaX, deltaY float32
		if wheelEvent.X > 0 {
			deltaX++
		} else if wheelEvent.X < 0 {
			deltaX--
		}
		if wheelEvent.Y > 0 {
			deltaY++
		} else if wheelEvent.Y < 0 {
			deltaY--
		}
		platform.imguiIO.AddMouseWheelDelta(deltaX, deltaY)
	case sdl.MOUSEBUTTONDOWN:
		buttonEvent := event.(*sdl.MouseButtonEvent)
		switch buttonEvent.Button {
		case sdl.BUTTON_LEFT:
			platform.buttonsDown[0] = true
		case sdl.BUTTON_RIGHT:
			platform.buttonsDown[1] = true
		case sdl.BUTTON_MIDDLE:
			platform.buttonsDown[2] = true
		}
	case sdl.TEXTINPUT:
		inputEvent := event.(*sdl.TextInputEvent)
		platform.imguiIO.AddInputCharacters(string(inputEvent.Text[:]))
	case sdl.KEYDOWN:
		keyEvent := event.(*sdl.KeyboardEvent)
		platform.imguiIO.KeyPress(int(keyEvent.Keysym.Scancode))
		modState := int(sdl.GetModState())
		platform.imguiIO.KeyShift(modState&sdl.KMOD_LSHIFT, modState&sdl.KMOD_RSHIFT)
		platform.imguiIO.KeyCtrl(modState&sdl.KMOD_LCTRL, modState&sdl.KMOD_RCTRL)
		platform.imguiIO.KeyAlt(modState&sdl.KMOD_LALT, modState&sdl.KMOD_RALT)
	case sdl.KEYUP:
		keyEvent := event.(*sdl.KeyboardEvent)
		platform.imguiIO.KeyRelease(int(keyEvent.Keysym.Scancode))
		modState := int(sdl.GetModState())
		platform.imguiIO.KeyShift(modState&sdl.KMOD_LSHIFT, modState&sdl.KMOD_RSHIFT)
		platform.imguiIO.KeyCtrl(modState&sdl.KMOD_LCTRL, modState&sdl.KMOD_RCTRL)
		platform.imguiIO.KeyAlt(modState&sdl.KMOD_LALT, modState&sdl.KMOD_RALT)
	}
}
