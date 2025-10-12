package main

import "github.com/jupiterrider/purego-sdl3/sdl"

// The simplest example that creates SDL GPU device, opens a window and clears the screen to specified color.

func main() {
	sdl.SetLogPriorities(sdl.LogPriorityVerbose)

	defer sdl.Quit()
	if !sdl.Init(sdl.InitVideo) {
		panic(sdl.GetError())
	}

	device := sdl.CreateGPUDevice(0b1111111111111111, true, "") // Any of the shader format because we won't use shaders in this example
	if device == nil {
		panic(sdl.GetError())
	}
	defer sdl.DestroyGPUDevice(device)

	window := sdl.CreateWindow("GPU test", 1280, 720, sdl.WindowResizable)
	if window == nil {
		panic(sdl.GetError())
	}
	defer sdl.DestroyWindow(window)

	if !sdl.ClaimWindowForGPUDevice(device, window) {
		panic(sdl.GetError())
	}

	running := true
	for running {
		var event sdl.Event
		for sdl.PollEvent(&event) {
			switch event.Type() {
			case sdl.EventQuit, sdl.EventWindowCloseRequested:
				running = false
			}
		}

		// Acquire GPU commandBuffer buffer, it's a buffer of commands for GPU to execute.
		// Instead of waiting for every command to be executed you pack them in a buffer and send them to GPU to be executed without waiting.
		// Every single frame new buffer has to be created and submitted at the end of the frame.
		commandBuffer := sdl.AcquireGPUCommandBuffer(device)

		var swapchainTexture *sdl.GPUTexture
		sdl.WaitAndAcquireGPUSwapchainTexture(commandBuffer, window, &swapchainTexture, nil, nil)

		// The acquired swapchain texture may be nil for example when the window is minimized
		if swapchainTexture != nil {
			// Start render pass
			renderPass := sdl.BeginGPURenderPass(commandBuffer, []sdl.GPUColorTargetInfo{{
				Texture:    swapchainTexture,
				ClearColor: sdl.FColor{R: 0.2, G: 0.2, B: 0.2, A: 1},
				LoadOp:     sdl.GPULoadOpClear,
				StoreOp:    sdl.GPUStoreOpStore,
			}}, nil)

			// More drawing operations should go here...
			// But first you would need to create shaders, load them, create graphics pipeline, bind the pipeline, etc.
			// Depending on the approach maybe vertex buffers and/or transfer buffers would be required. So you would need to create, bind, etc.

			sdl.EndGPURenderPass(renderPass)
		}

		sdl.SubmitGPUCommandBuffer(commandBuffer)
	}
}
