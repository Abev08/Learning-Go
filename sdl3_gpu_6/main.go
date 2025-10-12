package main

import (
	"fmt"
	"math"
	"os"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jupiterrider/purego-sdl3/img"
	"github.com/jupiterrider/purego-sdl3/sdl"
)

// Followed: https://www.youtube.com/watch?v=1tTcQ9zLY7E (loading obj file)
//       and https://www.youtube.com/watch?v=NXY00vKi3iA (depth testing)
//       and https://www.youtube.com/watch?v=cVL1Ih2xf0Q (camera movement)

// Cube model and texture downloaded from: https://github.com/garykac/3d-cubes

// Textured 3D model (from .obj file).
// Based on previous example.
//  - .obj file parser was added that creates object data,
//  - depth texture was created and attached to the pipeline for correct depth testing,
//  - camera movement was added (WASD + Space/Ctrl + mouse),

// Shaders are written in GLSL and requires Vulkan SDK to compile them into SPIRV - shading language for Vulkan
// glslc -fshader-stage=vertex shaders/vertex.glsl -o shaders/vertex.spv
// glslc -fshader-stage=fragment shaders/fragment.glsl -o shaders/fragment.spv

var WHITE = sdl.FColor{R: 1, G: 1, B: 1, A: 1} // White color
var KeyboardState = make(map[sdl.Keycode]bool) // State of the keyboard keys, true = pressed

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
	sdl.SetWindowRelativeMouseMode(window, true)

	if !sdl.ClaimWindowForGPUDevice(device, window) {
		panic(sdl.GetError())
	}

	// Load image into surface
	textureSurface := img.Load("cube.png")
	if textureSurface == nil {
		panic(sdl.GetError())
	} else if textureSurface.Format != sdl.PixelFormatRGBA32 {
		panic("We need R8G8B8A8 pixel format")
	}
	defer sdl.DestroySurface(textureSurface)

	// Create GPU texture with the size of loaded image
	texture := sdl.CreateGPUTexture(device, &sdl.GPUTextureCreateInfo{
		Type:              sdl.GPUTextureType2D,
		Format:            sdl.GPUTextureFormatR8G8B8A8Unorm,
		Usage:             sdl.GPUTextureUsageSampler,
		Width:             uint32(textureSurface.W),
		Height:            uint32(textureSurface.H),
		LayerCountOrDepth: 1,
		NumLevels:         1,
	})
	if texture == nil {
		panic(sdl.GetError())
	}
	defer sdl.ReleaseGPUTexture(device, texture)

	// Create transfer buffer separate for the texture, it's easier that way to define it's size than reuse different one
	textureTransferBuffer := sdl.CreateGPUTransferBuffer(device, &sdl.GPUTransferBufferCreateInfo{
		Usage: sdl.GPUTransferBufferUsageUpload,
		Size:  uint32(textureSurface.W * textureSurface.H * 4), // 4 bytes per pixel * amount of pixels
	})
	if textureTransferBuffer == nil {
		panic(sdl.GetError())
	}

	// Copy data into transfer buffer
	ptr := sdl.MapGPUTransferBuffer(device, textureTransferBuffer, false)
	sourcePixels := unsafe.Slice((*byte)(textureSurface.Pixels), textureSurface.W*textureSurface.H*4) // 4 bytes per pixel * amount of pixels
	destinationPixels := unsafe.Slice((*byte)(ptr), textureSurface.W*textureSurface.H*4)
	copy(destinationPixels, sourcePixels)
	sdl.UnmapGPUTransferBuffer(device, textureTransferBuffer)

	// Transfer texture data to GPU before render loop - the texture data won't change and there is no point to reupload it every single frame
	cb := sdl.AcquireGPUCommandBuffer(device)
	cp := sdl.BeginGPUCopyPass(cb)
	sdl.UploadToGPUTexture(cp,
		&sdl.GPUTextureTransferInfo{TransferBuffer: textureTransferBuffer},
		&sdl.GPUTextureRegion{
			Texture: texture,
			W:       uint32(textureSurface.W),
			H:       uint32(textureSurface.H),
			D:       1,
		},
		false)
	sdl.EndGPUCopyPass(cp)
	if !sdl.SubmitGPUCommandBuffer(cb) {
		panic(sdl.GetError())
	}
	sdl.ReleaseGPUTransferBuffer(device, textureTransferBuffer) // Release texture transfer buffer because it won't be required any more

	// 3D object definition
	object := parseObj("cube.obj")
	fmt.Printf("Loaded model has: %d vertices", len(object.Vertices))

	// Create texture sampler for the fragment shader
	sampler := sdl.CreateGPUSampler(device, &sdl.GPUSamplerCreateInfo{}) // Default values are ok
	defer sdl.ReleaseGPUSampler(device, sampler)

	// Load already compiled shaders and create graphics pipeline
	vertexShader := loadShader(device, "shaders/vertex.spv", sdl.GPUShaderStageVertex, 0, 1)
	fragmentShader := loadShader(device, "shaders/fragment.spv", sdl.GPUShaderStageFragment, 1, 0)
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
				sdl.GPUVertexAttribute{ // 3rd VertexData attribute (uv)
					Location: 2,
					Format:   sdl.GPUVertexElementFormatFloat2,
					Offset:   uint32(unsafe.Offsetof(VertexData{}.UV)),
				},
			}[0],
			NumVertexAttributes: 3,
		},
		FragmentShader: fragmentShader,
		PrimitiveType:  sdl.GPUPrimitiveTypeTrianglelist,
		DepthStencilState: sdl.GPUDepthStencilState{ // Enable depth testing
			CompareOp:        sdl.GPUCompareOpLess,
			EnableDepthTest:  true,
			EnableDepthWrite: true,
		},
		TargetInfo: sdl.GPUGraphicsPipelineTargetInfo{
			ColorTargetDescriptions: &sdl.GPUColorTargetDescription{Format: sdl.GetGPUSwapchainTextureFormat(device, window)},
			NumColorTargets:         1,
			DepthStencilFormat:      sdl.GPUTextureFormatD24Unorm, // Has to match format of created gpu depth texture
			HasDepthStencilTarget:   true,
		},
	})
	sdl.ReleaseGPUShader(device, vertexShader)
	sdl.ReleaseGPUShader(device, fragmentShader)
	defer sdl.ReleaseGPUGraphicsPipeline(device, pipeline)

	// Create vertex buffer object
	vboInfo := sdl.GPUBufferCreateInfo{
		Usage: sdl.GPUBufferUsageVertex,
		Size:  uint32(len(object.Vertices) * int(unsafe.Sizeof(VertexData{}))),
	}
	vbo := sdl.CreateGPUBuffer(device, &vboInfo)
	if vbo == nil {
		panic(sdl.GetError())
	}
	defer sdl.ReleaseGPUBuffer(device, vbo)

	// Create index buffer object
	iboInfo := sdl.GPUBufferCreateInfo{
		Usage: sdl.GPUBufferUsageIndex,
		Size:  uint32(len(object.Indices) * int(unsafe.Sizeof(object.Indices[0]))),
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

	ptr = sdl.MapGPUTransferBuffer(device, transferBuffer, false)
	gpuVertices := unsafe.Slice((*VertexData)(ptr), len(object.Vertices)) // Vertex data starts at the beginning of the transfer buffer
	copy(gpuVertices, object.Vertices)
	ptr = unsafe.Pointer(uintptr(ptr) + uintptr(len(object.Vertices))*unsafe.Sizeof(object.Vertices[0])) // Index data starts after vertex data in the transfer buffer - some pointer arithmetics is needed
	gpuIndices := unsafe.Slice((*uint16)(ptr), len(object.Indices))
	copy(gpuIndices, object.Indices)
	sdl.UnmapGPUTransferBuffer(device, transferBuffer)

	var depthTexture *sdl.GPUTexture

	camera := NewCamera(mgl32.Vec3{0, 0, -3})
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
			case sdl.EventWindowResized:
				if depthTexture != nil {
					sdl.ReleaseGPUTexture(device, depthTexture)
				}
				depthTexture = nil
			case sdl.EventMouseMotion:
				e := event.Motion()
				camera.Rotate(e.Xrel, e.Yrel)
			case sdl.EventKeyDown, sdl.EventKeyUp:
				e := event.Key()
				KeyboardState[e.Key] = e.Down
			}
		}

		rotation += rotationSpeed * dt
		camera.Update(dt)

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

			if depthTexture == nil {
				// Create depth texture
				depthTexture = sdl.CreateGPUTexture(device, &sdl.GPUTextureCreateInfo{
					Format:            sdl.GPUTextureFormatD24Unorm,
					Usage:             sdl.GPUTextureUsageDepthStencilTarget,
					Width:             w,
					Height:            h,
					LayerCountOrDepth: 1,
					NumLevels:         1,
				})
				if depthTexture == nil {
					panic(sdl.GetError())
				}
				defer sdl.ReleaseGPUTexture(device, depthTexture)
			}

			// Start render pass
			renderPass := sdl.BeginGPURenderPass(commandBuffer, []sdl.GPUColorTargetInfo{{
				Texture:    swapchainTexture,
				ClearColor: sdl.FColor{R: 0.2, G: 0.2, B: 0.2, A: 1},
				LoadOp:     sdl.GPULoadOpClear,
				StoreOp:    sdl.GPUStoreOpStore,
			}}, &sdl.GPUDepthStencilTargetInfo{
				Texture:    depthTexture,
				ClearDepth: 1,
				LoadOp:     sdl.GPULoadOpClear,
				StoreOp:    sdl.GPUStoreOpDontCare,
			})

			sdl.BindGPUGraphicsPipeline(renderPass, pipeline)
			sdl.BindGPUVertexBuffers(renderPass, 0, &sdl.GPUBufferBinding{Buffer: vbo, Offset: 0}, 1)
			sdl.BindGPUIndexBuffer(renderPass, &sdl.GPUBufferBinding{Buffer: ibo, Offset: 0}, sdl.GPUIndexElementSize16Bit)
			sdl.BindGPUFragmentSamplers(renderPass, 0, &sdl.GPUTextureSamplerBinding{Texture: texture, Sampler: sampler}, 1)

			// Create MVP matrix and push it to uniform vertex buffer
			// The cube vertices are in range of (0, 1) instead of (-1, 1), so the model is aligned to one of the edges of the cube
			// instead of the center. Moving the cube by the {-0.5, -0.5, -0.5} vector moves it's origin to the center.
			model := mgl32.Rotate3DY(mgl32.DegToRad(rotation)).Mat4()                                // Rotate the cube in the Y axis
			model = model.Mul4(mgl32.Rotate3DZ(mgl32.DegToRad(rotation / 2)).Mat4())                 // Rotate the cube in the Z axis but slower than in the Y axis
			model = model.Mul4(mgl32.Translate3D(-0.5, -0.5, -0.5))                                  // Move the cube origin to its center
			view := camera.View                                                                      // Look at the cube using the camera
			projection := mgl32.Perspective(mgl32.DegToRad(70), float32(w)/float32(h), 0.0001, 1000) // Apply perspective projection
			ubo := UBO{mvp: projection.Mul4(view.Mul4(model))}
			sdl.PushGPUVertexUniformData(commandBuffer, 0, unsafe.Pointer(&ubo), uint32(unsafe.Sizeof(UBO{})))

			sdl.DrawGPUIndexedPrimitives(renderPass, uint32(len(object.Indices)), 1, 0, 0, 0) // Order the GPU to render using indices

			sdl.EndGPURenderPass(renderPass)
		}

		sdl.SubmitGPUCommandBuffer(commandBuffer)
	}
}

func loadShader(device *sdl.GPUDevice, path string, stage sdl.GPUShaderStage, numSamplers, numUniformBuffers uint32) *sdl.GPUShader {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return sdl.CreateGPUShader(device, &sdl.GPUShaderCreateInfo{
		CodeSize:          uint64(len(data)),
		Code:              &data[0],
		Format:            sdl.GPUShaderFormatSPIRV,
		Stage:             stage,
		NumSamplers:       numSamplers,
		NumUniformBuffers: numUniformBuffers,
	})
}

type VertexData struct {
	Position mgl32.Vec3
	Color    sdl.FColor
	UV       mgl32.Vec2
}

// Uniform buffer object
type UBO struct {
	mvp mgl32.Mat4
}

type ObjectData struct {
	Vertices []VertexData
	Indices  []uint16
}

type Camera struct {
	position mgl32.Vec3 // Position of the camera
	target   mgl32.Vec3 // Target position at which the camera is looking

	yaw   float32 // Rotation around Y axis
	pitch float32 // Rotation around X axis

	forward mgl32.Vec3 // Forward vector - from position to target
	right   mgl32.Vec3 // Right vector - from position to target
	up      mgl32.Vec3 // Up vector - from position to target

	View mgl32.Mat4 // 4x4 camera view matrix
}

func NewCamera(position mgl32.Vec3) Camera {
	c := Camera{
		position: position,

		yaw: 90,

		forward: mgl32.Vec3{0, 0, 1},
		right:   mgl32.Vec3{1, 0, 0},
		up:      mgl32.Vec3{0, 1, 0},
	}
	c.target = c.position.Add(c.forward)
	c.updateViewMatrix()

	return c
}

func (c *Camera) Rotate(xDiff, yDiff float32) {
	const cameraRotateSpeed float32 = 0.5

	c.yaw += xDiff * cameraRotateSpeed // left/right mouse movement rotates camera around Y axis
	for c.yaw < 0 {
		c.yaw += 360
	}
	for c.yaw > 360 {
		c.yaw -= 360
	}

	c.pitch -= yDiff * cameraRotateSpeed // up/down mouse movement rotates camera around X axis
	if c.pitch < -89 {
		c.pitch = -89
	} else if c.pitch > 89 {
		c.pitch = 89
	}

	sin, cos := math.Sincos(float64(mgl32.DegToRad(c.yaw)))
	sinYaw, cosYaw := float32(sin), float32(cos)
	sin, cos = math.Sincos(float64(mgl32.DegToRad(c.pitch)))
	sinPitch, cosPitch := float32(sin), float32(cos)

	c.forward = mgl32.Vec3{cosYaw * cosPitch, sinPitch, sinYaw * cosPitch}.Normalize() // Forward vector
	c.right = c.forward.Cross(mgl32.Vec3{0, 1, 0}).Normalize()                         // Right vector
	c.up = c.right.Cross(c.forward).Normalize()                                        // Up vector

	c.target = c.position.Add(c.forward)
	c.updateViewMatrix()
}

func (c *Camera) Update(dt float32) {
	const cameraMoveSpeed float32 = 5

	dir := mgl32.Vec3{}
	if KeyboardState[sdl.KeycodeW] {
		dir[0] += 1
	}
	if KeyboardState[sdl.KeycodeS] {
		dir[0] -= 1
	}
	if KeyboardState[sdl.KeycodeA] {
		dir[1] -= 1
	}
	if KeyboardState[sdl.KeycodeD] {
		dir[1] += 1
	}
	if KeyboardState[sdl.KeycodeSpace] {
		dir[2] += 1
	}
	if KeyboardState[sdl.KeycodeLCtrl] {
		dir[2] -= 1
	}

	if dir[0] == 0 && dir[1] == 0 && dir[2] == 0 {
		return
	}

	move := c.forward.Mul(dir[0])                    // Forward/backward movement - Z axis
	move = move.Add(c.right.Mul(dir[1]))             // Left/right movement - X axis
	move = move.Add(mgl32.Vec3{0, 1, 0}.Mul(dir[2])) // Up/down movement - relative to word "up"
	move = move.Normalize()
	move = move.Mul(cameraMoveSpeed * dt) // Apply camera speed

	c.position = c.position.Add(move)
	c.target = c.position.Add(c.forward)
	c.updateViewMatrix()
}

func (c *Camera) updateViewMatrix() {
	c.View = mgl32.LookAtV(c.position, c.target, c.up)
}
