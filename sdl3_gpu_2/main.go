package main

import (
	"os"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jupiterrider/purego-sdl3/sdl"
)

// Followed: https://www.youtube.com/watch?v=9zrHmy3b0x0

// The example draws rotating triangle.
//  - vertices for the triangle are hardcoded in vertex shader,
//  - color of the triangle is hardcoded in the fragment shader,
//  - vertex uniform buffer is used to transfer MVP (model * view * projection) matrix to vertex shader,
//  - mvp matrix transforms hardcoded vertices resulting in rotating triangle,

// Shaders are written in GLSL and requires Vulkan SDK to compile them into SPIRV - shading language for Vulkan
// glslc -fshader-stage=vertex shaders/vertex.glsl -o shaders/vertex.spv
// glslc -fshader-stage=fragment shaders/fragment.glsl -o shaders/fragment.spv

func main() {
	sdl.SetLogPriorities(sdl.LogPriorityVerbose)

	defer sdl.Quit()
	if !sdl.Init(sdl.InitVideo) {
		panic(sdl.GetError())
	}

	device := sdl.CreateGPUDevice(sdl.GPUShaderFormatSPIRV, true, "")
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

	// Load already compiled shaders and create graphics pipeline
	vertexShader := loadShader(device, "shaders/vertex.spv", sdl.GPUShaderStageVertex, 1)
	fragmentShader := loadShader(device, "shaders/fragment.spv", sdl.GPUShaderStageFragment, 0)
	pipeline := sdl.CreateGPUGraphicsPipeline(device, &sdl.GPUGraphicsPipelineCreateInfo{
		VertexShader:   vertexShader,
		FragmentShader: fragmentShader,
		PrimitiveType:  sdl.GPUPrimitiveTypeTrianglelist,
		TargetInfo: sdl.GPUGraphicsPipelineTargetInfo{
			ColorTargetDescriptions: &sdl.GPUColorTargetDescription{Format: sdl.GetGPUSwapchainTextureFormat(device, window)},
			NumColorTargets:         1,
		},
	})
	sdl.ReleaseGPUShader(device, vertexShader)
	sdl.ReleaseGPUShader(device, fragmentShader)
	defer sdl.ReleaseGPUGraphicsPipeline(device, pipeline)

	var rotation, rotationSpeed float32 = 0, mgl32.DegToRad(5000)
	previousTick := sdl.GetTicks()
	running := true
	for running {
		tick := sdl.GetTicks()
		dt := float32(tick-previousTick) / 1000
		previousTick = tick

		var event sdl.Event
		for sdl.PollEvent(&event) {
			switch event.Type() {
			case sdl.EventQuit, sdl.EventWindowCloseRequested:
				running = false
			}
		}

		rotation += rotationSpeed * dt

		// Acquire GPU commandBuffer buffer, it's a buffer of commands for GPU to execute.
		// Instead of waiting for every command to be executed you pack them in a buffer and send them to GPU to be executed without waiting.
		// Every single frame new buffer has to be created and submitted at the end of the frame.
		commandBuffer := sdl.AcquireGPUCommandBuffer(device)

		var swapchainTexture *sdl.GPUTexture
		var w, h uint32
		sdl.WaitAndAcquireGPUSwapchainTexture(commandBuffer, window, &swapchainTexture, &w, &h)

		// The acquired swapchain texture may be nil for example when the window is minimized
		if swapchainTexture != nil {
			// Start render pass
			renderPass := sdl.BeginGPURenderPass(commandBuffer, []sdl.GPUColorTargetInfo{{
				Texture:    swapchainTexture,
				ClearColor: sdl.FColor{R: 0.2, G: 0.2, B: 0.2, A: 1},
				LoadOp:     sdl.GPULoadOpClear,
				StoreOp:    sdl.GPUStoreOpStore,
			}}, nil)

			sdl.BindGPUGraphicsPipeline(renderPass, pipeline)

			// Create MVP matrix and push it to uniform vertex buffer
			model := mgl32.Rotate3DY(mgl32.DegToRad(rotation)).Mat4()
			proj := mgl32.Perspective(mgl32.DegToRad(70), float32(w)/float32(h), 0, 1000)
			ubo := UBO{mvp: proj.Mul4(mgl32.Translate3D(0, 0, -1.5).Mul4(model))}
			sdl.PushGPUVertexUniformData(commandBuffer, 0, unsafe.Pointer(&ubo), uint32(unsafe.Sizeof(UBO{})))

			sdl.DrawGPUPrimitives(renderPass, 3, 1, 0, 0) // Order the GPU to render 3 vertices, vertex data is hardcoded in vertex shader

			sdl.EndGPURenderPass(renderPass)
		}

		sdl.SubmitGPUCommandBuffer(commandBuffer)
	}
}

func loadShader(device *sdl.GPUDevice, path string, stage sdl.GPUShaderStage, numUniformBuffers uint32) *sdl.GPUShader {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return sdl.CreateGPUShader(device, &sdl.GPUShaderCreateInfo{
		CodeSize:          uint64(len(data)),
		Code:              &data[0],
		Format:            sdl.GPUShaderFormatSPIRV,
		Stage:             stage,
		NumUniformBuffers: numUniformBuffers,
	})
}

type UBO struct {
	mvp mgl32.Mat4
}
