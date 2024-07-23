package main

import (
	"fmt"
	"log"

	"github.com/veandco/go-sdl2/sdl"
)

// SDL2 and OS information provided by SDL2 library.
// For this example SDL2.dll is required.

func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		log.Fatal(err)
	}
	defer sdl.Quit()

	// SDL
	fmt.Printf("SDL revision: %s\n", sdl.GetRevision())
	var v sdl.Version
	sdl.GetVersion(&v)
	fmt.Printf("SDL version: %d.%d.%d\n", v.Major, v.Minor, v.Patch)
	fmt.Printf("SDL compiled version: %d\n", sdl.COMPILEDVERSION())

	// Operating system
	fmt.Printf("OS: %s\n", sdl.GetPlatform())
	fmt.Printf("RAM: %d GB\n", sdl.GetSystemRAM())
	fmt.Printf("CPU cores count: %d\n", sdl.GetCPUCount())
	fmt.Printf("CPU supports RDTSC: %t\n", sdl.HasRDTSC())
	fmt.Printf("CPU supports AltiVec: %t\n", sdl.HasAltiVec())
	fmt.Printf("CPU supports MMX: %t\n", sdl.HasMMX())
	fmt.Printf("CPU supports 3DNow: %t\n", sdl.Has3DNow())
	fmt.Printf("CPU supports SSE: %t\n", sdl.HasSSE())
	fmt.Printf("CPU supports SSE2: %t\n", sdl.HasSSE2())
	fmt.Printf("CPU supports SSE3: %t\n", sdl.HasSSE3())
	fmt.Printf("CPU supports SSE41: %t\n", sdl.HasSSE41())
	fmt.Printf("CPU supports SSE42: %t\n", sdl.HasSSE42())
	fmt.Printf("CPU supports AVX: %t\n", sdl.HasAVX())
	fmt.Printf("CPU supports AVX512F: %t\n", sdl.HasAVX512F())
	fmt.Printf("CPU supports AVX2: %t\n", sdl.HasAVX2())
	fmt.Printf("CPU supports NEON: %t\n", sdl.HasNEON())
	fmt.Printf("Current device is a tablet?: %t\n", sdl.IsTablet())

	// Power state
	powerState, powerSecs, powerCharge := sdl.GetPowerInfo()
	var ps string
	switch powerState {
	case sdl.POWERSTATE_UNKNOWN:
		ps = "unknown"
	case sdl.POWERSTATE_ON_BATTERY:
		ps = "using battery"
	case sdl.POWERSTATE_NO_BATTERY:
		ps = "battery missing"
	case sdl.POWERSTATE_CHARGING:
		ps = "battery charging"
	case sdl.POWERSTATE_CHARGED:
		ps = "battery charged"
	}
	fmt.Printf("Power info: %s, charged %d%%, time left [s] %d\n", ps, powerCharge, powerSecs) // "%%" escapes "%" symbol

	// Video drivers
	fmt.Println("Available video drivers:")
	currVideoDriver, _ := sdl.GetCurrentVideoDriver()
	for i := 0; true; i++ {
		s := sdl.GetVideoDriver(i)
		if len(s) <= 0 {
			break
		}
		fmt.Printf(" - %s%s\n", s, Tif(currVideoDriver == s, "*", ""))
	}

	// Render drivers
	fmt.Println("Available renderers:")
	for i := 0; true; i++ {
		rendererInfo := sdl.RendererInfo{}
		idx, err := sdl.GetRenderDriverInfo(i, &rendererInfo)
		if err != nil {
			break
		}
		fmt.Printf(" - [%d] %s\n", idx, rendererInfo.Name)
	}

	// Resolutions
	fmt.Println("Available resolutions:")
	for display := 0; true; display++ {
		currMode, err := sdl.GetCurrentDisplayMode(display)
		if err != nil {
			break
		}
		for mode := 0; true; mode++ {
			ds, err := sdl.GetDisplayMode(display, mode)
			if err != nil {
				break
			}
			fmt.Printf(" - display %d, %dx%d px %d Hz%s\n", display, ds.W, ds.H, ds.RefreshRate,
				Tif(currMode.Format == ds.Format && currMode.W == ds.W && currMode.H == ds.H && currMode.RefreshRate == ds.RefreshRate, "*", ""))
		}
	}

	// Audio drivers
	fmt.Println("Available audio drivers:")
	currAudioDriver := sdl.GetCurrentAudioDriver()
	for i := 0; true; i++ {
		s := sdl.GetAudioDriver(i)
		if len(s) <= 0 {
			break
		}
		fmt.Printf(" - %s%s\n", s, Tif(currAudioDriver == s, "*", ""))
	}
}

// Ternary if statement
func Tif[T any](condition bool, vTrue, vFalse T) T {
	if condition {
		return vTrue
	}
	return vFalse
}
