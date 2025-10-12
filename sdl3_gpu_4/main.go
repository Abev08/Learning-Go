package main

import (
	"os"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jupiterrider/purego-sdl3/sdl"
)

// Followed: https://www.youtube.com/watch?v=QltHibRYF8w

// The example draws rotating quad.
//  - vertex data describing the quad consists of position and color,
//  - indices describe how to create 2 triangles from vertex data,
//  - vertex buffer object (vbo) and index buffer object (ibo) is created to describe to GPU how the data is laid out,
//  - transfer buffer is used to transfer vertex and index data to GPU,
//  - color of the quad vertices is passed from vertex shader to fragment shader,
//  - vertex uniform buffer is used to transfer MVP (model * view * projection) matrix to vertex shader,
//  - mvp matrix transforms vertex data in the vertex shader resulting in rotating quad,

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

	// Quad definition
	quad := struct {
		Vertices []VertexData
		Indices  []uint16
	}{
		Vertices: []VertexData{
			VertexData{Position: mgl32.Vec3{-0.5, 0.5, 0}, Color: sdl.FColor{R: 0.8, G: 0, B: 0.2, A: 1}}, // top left
			VertexData{Position: mgl32.Vec3{0.5, 0.5, 0}, Color: sdl.FColor{R: 1, G: 1, B: 0, A: 1}},      // top right
			VertexData{Position: mgl32.Vec3{-0.5, -0.5, 0}, Color: sdl.FColor{R: 0, G: 0, B: 1, A: 1}},    // bottom left
			VertexData{Position: mgl32.Vec3{0.5, -0.5, 0}, Color: sdl.FColor{R: 0, G: 1, B: 0, A: 1}},     // bottom right
		},
		Indices: []uint16{
			0, 2, 1,
			1, 2, 3,
		},
	}

	// Load already compiled shaders and create graphics pipeline
	vertexShader := loadShader(device, "shaders/vertex.spv", sdl.GPUShaderStageVertex, 1)
	fragmentShader := loadShader(device, "shaders/fragment.spv", sdl.GPUShaderStageFragment, 0)
	pipeline := sdl.CreateGPUGraphicsPipeline(device, &sdl.GPUGraphicsPipelineCreateInfo{
		VertexShader: vertexShader,
		VertexInputState: sdl.GPUVertexInputState{
			VertexBufferDescriptions: &sdl.GPUVertexBufferDescription{ // Describe that vertex buffer consists of VertexData objects
				Slot:  0,
				Pitch: uint32(unsafe.Sizeof(VertexData{})),
			},
			NumVertexBuffers: 1,
			VertexAttributes: &[]sdl.GPUVertexAttribute{ // Describe that VertexData object consists multiple attributes, this has to match to what's declared in the vertex shader
				sdl.GPUVertexAttribute{ // 1st VertexData attribute (position)
					Location: 0,
					Format:   sdl.GPUVertexElementFormatFloat3,
					Offset:   uint32(unsafe.Offsetof(VertexData{}.Position)),
				},
				sdl.GPUVertexAttribute{ // 2nd VertexData attribute (color)
					Location: 1,
					Format:   sdl.GPUVertexElementFormatFloat4,
					Offset:   uint32(unsafe.Offsetof(VertexData{}.Color)),
				},
			}[0],
			NumVertexAttributes: 2,
		},
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

	// Create vertex buffer object
	vboInfo := sdl.GPUBufferCreateInfo{
		Usage: sdl.GPUBufferUsageVertex,
		Size:  uint32(len(quad.Vertices) * int(unsafe.Sizeof(VertexData{}))),
	}
	vbo := sdl.CreateGPUBuffer(device, &vboInfo)
	if vbo == nil {
		panic(sdl.GetError())
	}
	defer sdl.ReleaseGPUBuffer(device, vbo)

	// Create index buffer object
	iboInfo := sdl.GPUBufferCreateInfo{
		Usage: sdl.GPUBufferUsageIndex,
		Size:  uint32(len(quad.Indices) * int(unsafe.Sizeof(quad.Indices[0]))),
	}
	ibo := sdl.CreateGPUBuffer(device, &iboInfo)
	if ibo == nil {
		panic(sdl.GetError())
	}
	defer sdl.ReleaseGPUBuffer(device, ibo)

	// Create transfer buffer
	transferBuffer := sdl.CreateGPUTransferBuffer(device, &sdl.GPUTransferBufferCreateInfo{
		Usage: sdl.GPUTransferBufferUsageUpload,
		Size:  vboInfo.Size + iboInfo.Size,
	})
	if transferBuffer == nil {
		panic(sdl.GetError())
	}
	defer sdl.ReleaseGPUTransferBuffer(device, transferBuffer)

	ptr := sdl.MapGPUTransferBuffer(device, transferBuffer, false)
	gpuVertices := unsafe.Slice((*VertexData)(ptr), len(quad.Vertices)) // Vertex data starts at the beginning of the transfer buffer
	copy(gpuVertices, quad.Vertices)
	ptr = unsafe.Pointer(uintptr(ptr) + uintptr(len(quad.Vertices))*unsafe.Sizeof(quad.Vertices[0])) // Index data starts after vertex data in the transfer buffer - some pointer arithmetics is needed
	gpuIndices := unsafe.Slice((*uint16)(ptr), len(quad.Indices))
	copy(gpuIndices, quad.Indices)
	sdl.UnmapGPUTransferBuffer(device, transferBuffer)

	var rotation, rotationSpeed float32 = 0, mgl32.DegToRad(3000)
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

		// Transfer data to GPU
		copyPass := sdl.BeginGPUCopyPass(commandBuffer)
		sdl.UploadToGPUBuffer(copyPass,
			&sdl.GPUTransferBufferLocation{TransferBuffer: transferBuffer, Offset: 0},
			&sdl.GPUBufferRegion{Buffer: vbo, Offset: 0, Size: vboInfo.Size},
			false)
		sdl.UploadToGPUBuffer(copyPass,
			&sdl.GPUTransferBufferLocation{TransferBuffer: transferBuffer, Offset: vboInfo.Size},
			&sdl.GPUBufferRegion{Buffer: ibo, Offset: 0, Size: iboInfo.Size},
			false)
		sdl.EndGPUCopyPass(copyPass)

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
			sdl.BindGPUVertexBuffers(renderPass, 0, &sdl.GPUBufferBinding{Buffer: vbo, Offset: 0}, 1)
			sdl.BindGPUIndexBuffer(renderPass, &sdl.GPUBufferBinding{Buffer: ibo, Offset: 0}, sdl.GPUIndexElementSize16Bit)

			// Create MVP matrix and push it to uniform vertex buffer
			model := mgl32.Rotate3DY(mgl32.DegToRad(rotation)).Mat4()
			proj := mgl32.Perspective(mgl32.DegToRad(70), float32(w)/float32(h), 0.0001, 1000)
			ubo := UBO{mvp: proj.Mul4(mgl32.Translate3D(0, 0, -1.5).Mul4(model))}
			sdl.PushGPUVertexUniformData(commandBuffer, 0, unsafe.Pointer(&ubo), uint32(unsafe.Sizeof(UBO{})))

			sdl.DrawGPUIndexedPrimitives(renderPass, uint32(len(quad.Indices)), 1, 0, 0, 0) // Order the GPU to render using indices

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

type VertexData struct {
	Position mgl32.Vec3
	Color    sdl.FColor
}

type UBO struct {
	mvp mgl32.Mat4
}
