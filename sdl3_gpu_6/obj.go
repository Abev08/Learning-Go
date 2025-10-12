package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
)

func parseObj(path string) ObjectData {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	object := ObjectData{
		Vertices: make([]VertexData, 0),
		Indices:  make([]uint16, 0),
	}

	positions := make([]mgl32.Vec3, 0)
	uvs := make([]mgl32.Vec2, 0)
	idx := uint16(0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		switch line[0] {
		case 'v':
			switch line[1] {
			case ' ': // parse vertex position (vec3)
				tmp := strings.Fields(line[2:])
				p := mgl32.Vec3{}
				for i := 0; i < 3; i++ {
					f, err := strconv.ParseFloat(tmp[i], 32)
					if err != nil {
						panic(err)
					}
					p[i] = float32(f)
				}
				positions = append(positions, p)
			case 't': // parse texture uv (vec2)
				tmp := strings.Fields(line[3:])
				uv := mgl32.Vec2{}
				for i := 0; i < 2; i++ {
					f, err := strconv.ParseFloat(tmp[i], 32)
					if err != nil {
						panic(err)
					}
					if i == 1 {
						uv[i] = 1 - float32(f)
					} else {
						uv[i] = float32(f)
					}
				}
				uvs = append(uvs, uv)
			}
		case 'f': // parse face
			// Face is described as "f vertexIdx/uvIdx/normal", multiple triplets (v/uv/n) can be present in single line
			tmp := strings.Fields(line[2:])
			for i := 0; i < len(tmp); i++ {
				tmp2 := strings.Split(tmp[i], "/")
				vertexIdx, err := strconv.ParseUint(tmp2[0], 10, 16)
				if err != nil {
					panic(err)
				}
				uv := mgl32.Vec2{}
				if len(tmp2[1]) > 0 {
					uvIdx, err := strconv.ParseUint(tmp2[1], 10, 16)
					if err != nil {
						panic(err)
					}
					uv = uvs[uvIdx-1]
				}

				object.Vertices = append(object.Vertices, VertexData{
					Position: positions[vertexIdx-1],
					Color:    WHITE,
					UV:       uv,
				})
				object.Indices = append(object.Indices, idx)
				idx++
			}
		}
	}

	return object
}
