package main

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/jupiterrider/purego-sdl3/sdl"
)

// Followed: https://hamdy-elzanqali.medium.com/let-there-be-triangles-sdl-gpu-edition-bd82cf2ef615

// The example draws single triangle.
//  - vertex data of the triangle is stored in Vertex3D object, that is position and color of each of the vertices,
//  - vertex data is stored in vertex buffer,
//  - vertex buffer is transferred to GPU via transfer buffer,
//  - graphics pipeline knows about vertex attributes in vertex buffer and how to interpret them (position and color),
//  - fragment uniform buffer is used to transfer time value to create color pulse effect,

// Shaders are written in GLSL and requires Vulkan SDK to compile them into SPIRV - shading language for Vulkan
// glslc -fshader-stage=vertex shaders/vertex.glsl -o shaders/vertex.spv
// glslc -fshader-stage=fragment shaders/fragment.glsl -o shaders/fragment.spv

func main() {
	sdl.SetLogPriorities(sdl.LogPriorityVerbose) // Sets SDL log priorities to be more verbose

	defer sdl.Quit()
	if !sdl.Init(sdl.InitVideo) {
		panic(sdl.GetError())
	}

	// Print available GPU drivers
	fmt.Print("Available GPU drivers: ")
	for i := int32(0); i < sdl.GetNumGPUDrivers(); i++ {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(sdl.GetGPUDriver(i))
	}
	fmt.Println()

	// Create new GPU device
	// Name of the GPU device may be left empty for SDL to choose the best one.
	// If specific driver is chosen appropriate for that driver shader format has to be used.
	device := sdl.CreateGPUDevice(sdl.GPUShaderFormatSPIRV, true, "")
	if device == nil {
		panic(sdl.GetError())
	}
	defer sdl.DestroyGPUDevice(device)
	fmt.Printf("Created GPU device uses %s driver, and shader format '%d'\n", sdl.GetGPUDeviceDriver(device), sdl.GetGPUShaderFormats(device))

	// Create window that would use created GPU device
	window := sdl.CreateWindow("GPU test", 1280, 720, sdl.WindowResizable)
	if window == nil {
		panic(sdl.GetError())
	}
	defer sdl.DestroyWindow(window)
	if !sdl.ClaimWindowForGPUDevice(device, window) {
		panic(sdl.GetError())
	}
	defer sdl.ReleaseWindowFromGPUDevice(device, window)

	// Disables VSync
	// sdl.SetGPUSwapchainParameters(device, window, sdl.GPUSwapchainCompositionSDR, sdl.GPUPresentModeImmediate)

	// Load vertex shader
	vertexCode, err := os.ReadFile("shaders/vertex.spv")
	if err != nil {
		panic(err)
	}
	vertexInfo := sdl.GPUShaderCreateInfo{
		CodeSize: uint64(len(vertexCode)),
		Code:     &vertexCode[0],
		Format:   sdl.GPUShaderFormatSPIRV,
		Stage:    sdl.GPUShaderStageVertex,
	}
	vertexInfo.SetEntryPoint("main")
	vertexShader := sdl.CreateGPUShader(device, &vertexInfo)
	vertexCode = nil // Free the file
	if vertexShader == nil {
		panic(sdl.GetError())
	}

	// Load fragment shader
	fragmentCode, err := os.ReadFile("shaders/fragment.spv")
	if err != nil {
		panic(err)
	}
	fragmentInfo := sdl.GPUShaderCreateInfo{
		CodeSize:          uint64(len(fragmentCode)),
		Code:              &fragmentCode[0],
		Format:            sdl.GPUShaderFormatSPIRV,
		Stage:             sdl.GPUShaderStageFragment,
		NumUniformBuffers: 1,
	}
	fragmentInfo.SetEntryPoint("main")
	fragmentShader := sdl.CreateGPUShader(device, &fragmentInfo)
	fragmentCode = nil
	if fragmentShader == nil {
		panic(sdl.GetError())
	}

	// Create graphics pipelineInfo
	pipelineInfo := sdl.GPUGraphicsPipelineCreateInfo{
		VertexShader:   vertexShader,
		FragmentShader: fragmentShader,

		PrimitiveType: sdl.GPUPrimitiveTypeTrianglelist, // Draw triangles
	}

	// Vertices to be drawn on the screen
	vertices := []Vertex3D{
		{X: 0.0, Y: 0.5, Z: 0.0, Color: sdl.FColor{R: 1.0, G: 0.0, B: 0.0, A: 1.0}},   // top vertex
		{X: -0.4, Y: -0.5, Z: 0.0, Color: sdl.FColor{R: 1.0, G: 1.0, B: 0.0, A: 1.0}}, // bottom left vertex
		{X: 0.4, Y: -0.5, Z: 0.0, Color: sdl.FColor{R: 1.0, G: 0.0, B: 1.0, A: 1.0}},  // bottom right vertex
	}

	// Vertex buffer is a place on GPU where vertex data is stored
	vertexBuffer := sdl.CreateGPUBuffer(device, &sdl.GPUBufferCreateInfo{
		Size:  uint32(len(vertices) * int(unsafe.Sizeof(Vertex3D{}))),
		Usage: sdl.GPUBufferUsageVertex,
	})
	if vertexBuffer == nil {
		panic(sdl.GetError())
	}
	defer sdl.ReleaseGPUBuffer(device, vertexBuffer)

	// Attach vertex buffer to graphics pipeline
	// Describe layout of the vertex buffer
	pipelineInfo.VertexInputState.VertexBufferDescriptions = &[]sdl.GPUVertexBufferDescription{
		{
			Slot:             0,
			InputRate:        sdl.GPUVertexInputRateVertex,
			InstanceStepRate: 0,
			Pitch:            uint32(unsafe.Sizeof(Vertex3D{})),
		},
	}[0]
	pipelineInfo.VertexInputState.NumVertexBuffers = 1
	// Describe layout of vertex attributes in the vertex buffer
	vertexAttributes := make([]sdl.GPUVertexAttribute, 2)
	// a_position
	vertexAttributes[0].BufferSlot = 0                            // fetch data from the buffer at slot 0
	vertexAttributes[0].Location = 0                              // layout (location = 0) in shader
	vertexAttributes[0].Format = sdl.GPUVertexElementFormatFloat3 // vec3
	vertexAttributes[0].Offset = 0                                // start from the first byte from current buffer position
	// a_color
	vertexAttributes[1].BufferSlot = 0                                 // use buffer at slot 0
	vertexAttributes[1].Location = 1                                   // layout (location = 1) in shader
	vertexAttributes[1].Format = sdl.GPUVertexElementFormatFloat4      // vec4
	vertexAttributes[1].Offset = 3 * uint32(unsafe.Sizeof(float32(0))) // 4th float from current buffer position
	pipelineInfo.VertexInputState.VertexAttributes = &vertexAttributes[0]
	pipelineInfo.VertexInputState.NumVertexAttributes = uint32(len(vertexAttributes))

	// Describe the color target
	colorTargetDescriptions := []sdl.GPUColorTargetDescription{
		{
			BlendState: sdl.GPUColorTargetBlendState{
				EnableBlend:         true,
				ColorBlendOp:        sdl.GPUBlendOpAdd,
				AlphaBlendOp:        sdl.GPUBlendOpAdd,
				SrcColorBlendFactor: sdl.GPUBlendFactorSrcAlpha,
				DstColorBlendFactor: sdl.GPUBlendFactorOneMinusSrcAlpha,
				SrcAlphaBlendFactor: sdl.GPUBlendFactorSrcAlpha,
				DstAlphaBlendFactor: sdl.GPUBlendFactorOneMinusSrcAlpha,
			},
			Format: sdl.GetGPUSwapchainTextureFormat(device, window),
		},
	}
	pipelineInfo.TargetInfo.NumColorTargets = uint32(len(colorTargetDescriptions))
	pipelineInfo.TargetInfo.ColorTargetDescriptions = &colorTargetDescriptions[0]

	// Create graphics pipeline
	pipeline := sdl.CreateGPUGraphicsPipeline(device, &pipelineInfo)
	// We don't need the shaders after creating the pipeline
	sdl.ReleaseGPUShader(device, vertexShader)
	sdl.ReleaseGPUShader(device, fragmentShader)
	if pipeline == nil {
		panic(sdl.GetError())
	}
	defer sdl.ReleaseGPUGraphicsPipeline(device, pipeline)

	// Transfer buffer allows moving data from CPU to GPU
	transferBuffer := sdl.CreateGPUTransferBuffer(device, &sdl.GPUTransferBufferCreateInfo{
		Size:  uint32(len(vertices) * int(unsafe.Sizeof(Vertex3D{}))),
		Usage: sdl.GPUTransferBufferUsageUpload,
	})
	if transferBuffer == nil {
		panic(sdl.GetError())
	}
	defer sdl.ReleaseGPUTransferBuffer(device, transferBuffer)

	// Copy Vertices into Vertex Buffer using Transfer Buffer
	// First copy the data into transfer buffer
	data := unsafe.Slice((*Vertex3D)(sdl.MapGPUTransferBuffer(device, transferBuffer, false)), len(vertices))
	copy(data, vertices)
	sdl.UnmapGPUTransferBuffer(device, transferBuffer)
	// Next transfer data from transfer buffer into vertex buffer, this needs to be done using command buffer so inside render loop

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
		if commandBuffer == nil {
			panic(sdl.GetError())
		}

		// Start a copy pass
		copyPass := sdl.BeginGPUCopyPass(commandBuffer)
		source := sdl.GPUTransferBufferLocation{ // Where the data is in transfer buffer
			TransferBuffer: transferBuffer,
			Offset:         0,
		}
		destination := sdl.GPUBufferRegion{ // Where to copy the data from transfer buffer
			Buffer: vertexBuffer,
			Size:   uint32(len(vertices) * int(unsafe.Sizeof(Vertex3D{}))),
			Offset: 0,
		}
		sdl.UploadToGPUBuffer(copyPass, &source, &destination, true) // Upload the data
		sdl.EndGPUCopyPass(copyPass)                                 // End the copy pass

		timeUniform.time = float32(sdl.GetTicksNS()) / 1000000000                         // The time since the app started in seconds
		sdl.PushGPUFragmentUniformData(commandBuffer, 0, unsafe.Pointer(&timeUniform), 4) // Update uniform data in the fragment shader

		var swapchainTexture *sdl.GPUTexture
		var width, height uint32
		if !sdl.WaitAndAcquireGPUSwapchainTexture(commandBuffer, window, &swapchainTexture, &width, &height) {
			panic(sdl.GetError())
		}

		// Swapchain texture can be nil in some situations (for example when window is minimized). So it needs to be validated before use.
		if swapchainTexture != nil {
			// GPUColorTargetInfo tells GPU where to draw something, with what color and what operation should be performed
			colorTargetInfo := sdl.GPUColorTargetInfo{
				Texture:    swapchainTexture,
				ClearColor: sdl.FColor{R: 0.2, G: 0.2, B: 0.2, A: 1},
				LoadOp:     sdl.GPULoadOpClear, // GPULoadOpClear to clear previous content or GPULoadOpLoad to keep the previous content
				StoreOp:    sdl.GPUStoreOpStore,
			}

			// Start render pass
			renderPass := sdl.BeginGPURenderPass(commandBuffer, []sdl.GPUColorTargetInfo{colorTargetInfo}, nil)

			// Bind the graphics pipeline
			sdl.BindGPUGraphicsPipeline(renderPass, pipeline)

			// Bind the vertex buffer
			bufferBindings := []sdl.GPUBufferBinding{
				{
					Buffer: vertexBuffer, // index 0 is slot 0 in this example
					Offset: 0,            // start from the first byte
				},
			}
			sdl.BindGPUVertexBuffers(renderPass, 0, &bufferBindings[0], 1) // bind one buffer starting from slot 0

			// Issue a draw call
			sdl.DrawGPUPrimitives(renderPass, 3, 1, 0, 0)

			// End the render pass
			sdl.EndGPURenderPass(renderPass)
		}

		// ALWAYS submit the command buffer
		if !sdl.SubmitGPUCommandBuffer(commandBuffer) {
			panic(sdl.GetError())
		}
	}
}

type Vertex3D struct {
	X, Y, Z float32 // Position
	Color   sdl.FColor
}

type UniformBuffer struct {
	time float32
	// You can add other properties here
}

var timeUniform = UniformBuffer{}
