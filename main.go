package main

import (
	"encoding/json"
	"flag"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type WFCStruct struct {
	Texture       string `json:"texture"`
	Rotation      int8   `json:"rotation"`
	AllowRotation bool   `json:"allow_rotation,omitempty"`
	/*
				 posX
				/----\
		   negY |    | PosY
				\----/
				 negX
	*/
	PosX int64 `json:"pos_x"` // Front
	PosY int64 `json:"pos_y"` // Left
	NegX int64 `json:"neg_x"` // Back
	NegY int64 `json:"neg_y"` // Right
	PosZ int64 `json:"pos_z"` // Top
	NegZ int64 `json:"neg_z"` // Bottom
}

var (
	Structs []WFCStruct

	IndexedStructs = make(map[int64]WFCStruct)
	PosXIndices    = make(map[int64][]int64) // [PosX_ID] -> Index
	PosYIndices    = make(map[int64][]int64) // [PosY_ID] -> Index
	NegXIndices    = make(map[int64][]int64) // [NegX_ID] -> Index
	NegYIndices    = make(map[int64][]int64) // [NegY_ID] -> Index
	PosZIndices    = make(map[int64][]int64) // [PosZ_ID] -> Index
	NegZIndices    = make(map[int64][]int64) // [NegZ_ID] -> Index

	IndexedPosX = make(map[int64]int64) // [Index] -> PosX_ID
	IndexedPosY = make(map[int64]int64) // [Index] -> PosY_ID
	IndexedNegX = make(map[int64]int64) // [Index] -> NegX_ID
	IndexedNegY = make(map[int64]int64) // [Index] -> NegY_ID
	IndexedPosZ = make(map[int64]int64) // [Index] -> PosZ_ID
	IndexedNegZ = make(map[int64]int64) // [Index] -> NegZ_ID

	IndexedAssets = make(map[string][][]*color.RGBA64)

	Indices []int64

	// OutputMatrix [Z, [Y, [X]]]
	OutputMatrix [][][]int64

	XSize int64 // Width Size
	YSize int64 // Length Size
	ZSize int64 // Height Size

	InputJsonFilePath      string
	OutputMapImageFilePath string
)

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()

	// My current screen size divided by the size of the tiles (16x16)
	flag.Int64Var(&XSize, "x", 3840/16, `X - X amount of tiles`)
	flag.Int64Var(&YSize, "y", 2160/16, `Y - Y amount of tiles`)
	flag.Int64Var(&ZSize, "z", 1, `Z - Z amount of tiles`)
	flag.StringVar(&InputJsonFilePath, "input", "input.json", "Input json file path")
	flag.StringVar(&OutputMapImageFilePath, "output", "output.png", "Output map image file path")

	flag.Parse()
}

func RotateRGBA64Matrix(matrix [][]*color.RGBA64) [][]*color.RGBA64 {
	var i, j int

	var matrixLen = len(matrix)
	var temp *color.RGBA64

	for i = 0; i < matrixLen/2; i++ {
		for j = i; j < matrixLen-i-1; j++ {
			temp = matrix[i][j]

			matrix[i][j] = matrix[matrixLen-1-j][i]
			matrix[matrixLen-1-j][i] = matrix[matrixLen-1-i][matrixLen-1-j]
			matrix[matrixLen-1-i][matrixLen-1-j] = matrix[j][matrixLen-1-i]
			matrix[j][matrixLen-1-i] = temp
		}
	}

	return matrix
}

func GetDuplications(firstArray, secondArray []int64) (returnedArray []int64) {
	var allKeys = make(map[int64]int8)

	var number int64
	for _, number = range firstArray {
		allKeys[number] = 1
	}

	for _, number = range secondArray {
		var ok bool

		if _, ok = allKeys[number]; ok {
			returnedArray = append(returnedArray, number)
		}
	}

	return
}

func AddWFCStructIntoStructs(i int64, wfcStruct WFCStruct) int64 {
	IndexedStructs[i] = wfcStruct

	Indices = append(Indices, i)

	IndexedPosX[i] = wfcStruct.PosX
	IndexedPosY[i] = wfcStruct.PosY
	IndexedNegX[i] = wfcStruct.NegX
	IndexedNegY[i] = wfcStruct.NegY
	IndexedPosZ[i] = wfcStruct.PosZ
	IndexedNegZ[i] = wfcStruct.NegZ

	var ok bool

	if _, ok = PosXIndices[wfcStruct.PosX]; !ok {
		PosXIndices[wfcStruct.PosX] = make([]int64, 0, 10)
	}

	PosXIndices[wfcStruct.PosX] = append(PosXIndices[wfcStruct.PosX], i)

	if _, ok = PosYIndices[wfcStruct.PosY]; !ok {
		PosYIndices[wfcStruct.PosY] = make([]int64, 0, 10)
	}

	PosYIndices[wfcStruct.PosY] = append(PosYIndices[wfcStruct.PosY], i)

	if _, ok = NegXIndices[wfcStruct.NegX]; !ok {
		NegXIndices[wfcStruct.NegX] = make([]int64, 0, 10)
	}

	NegXIndices[wfcStruct.NegX] = append(NegXIndices[wfcStruct.NegX], i)

	if _, ok = NegYIndices[wfcStruct.NegY]; !ok {
		NegYIndices[wfcStruct.NegY] = make([]int64, 0, 10)
	}

	NegYIndices[wfcStruct.NegY] = append(NegYIndices[wfcStruct.NegY], i)

	if _, ok = PosZIndices[wfcStruct.PosZ]; !ok {
		PosZIndices[wfcStruct.PosZ] = make([]int64, 0, 10)
	}

	PosZIndices[wfcStruct.PosZ] = append(PosZIndices[wfcStruct.PosZ], i)

	if _, ok = NegZIndices[wfcStruct.NegZ]; !ok {
		NegZIndices[wfcStruct.NegZ] = make([]int64, 0, 10)
	}

	NegZIndices[wfcStruct.NegZ] = append(NegZIndices[wfcStruct.NegZ], i)

	return i + 1
}

func GenerateRotations() {
	var index int64 = 0

	var wfcStruct WFCStruct
	for _, wfcStruct = range Structs {
		index = AddWFCStructIntoStructs(index, WFCStruct{
			Texture:       wfcStruct.Texture,
			Rotation:      0,
			AllowRotation: false,
			PosX:          wfcStruct.PosX,
			PosY:          wfcStruct.PosY,
			NegX:          wfcStruct.NegX,
			NegY:          wfcStruct.NegY,
			PosZ:          wfcStruct.PosZ,
			NegZ:          wfcStruct.NegZ,
		})

		if !wfcStruct.AllowRotation {
			continue
		}

		index = AddWFCStructIntoStructs(index, WFCStruct{
			Texture:       wfcStruct.Texture,
			Rotation:      1,
			AllowRotation: false,
			PosX:          wfcStruct.PosY,
			PosY:          wfcStruct.NegX,
			NegX:          wfcStruct.NegY,
			NegY:          wfcStruct.PosX,
			PosZ:          wfcStruct.PosZ,
			NegZ:          wfcStruct.NegZ,
		})
		index = AddWFCStructIntoStructs(index, WFCStruct{
			Texture:       wfcStruct.Texture,
			Rotation:      2,
			AllowRotation: false,
			PosX:          wfcStruct.NegX,
			PosY:          wfcStruct.NegY,
			NegX:          wfcStruct.PosX,
			NegY:          wfcStruct.PosY,
			PosZ:          wfcStruct.PosZ,
			NegZ:          wfcStruct.NegZ,
		})
		index = AddWFCStructIntoStructs(index, WFCStruct{
			Texture:       wfcStruct.Texture,
			Rotation:      3,
			AllowRotation: false,
			PosX:          wfcStruct.NegY,
			PosY:          wfcStruct.PosX,
			NegX:          wfcStruct.PosY,
			NegY:          wfcStruct.NegX,
			PosZ:          wfcStruct.PosZ,
			NegZ:          wfcStruct.NegZ,
		})
	}
}

func LoadInput() (err error) {
	var structsFile *os.File
	if structsFile, err = os.Open(InputJsonFilePath); err != nil {
		return
	}

	var jsonDecoder = json.NewDecoder(structsFile)
	if err = jsonDecoder.Decode(&Structs); err != nil {
		return
	}

	return
}

func GenerateMap() (err error) {
	rand.Seed(time.Now().Unix() / 10 * 102)

	OutputMatrix = make([][][]int64, ZSize)

	var z, y, x int64
	for z = 0; z < ZSize; z++ {
		OutputMatrix[z] = make([][]int64, YSize)

		for y = 0; y < YSize; y++ {
			OutputMatrix[z][y] = make([]int64, XSize)

			for x = 0; x < XSize; x++ {
				var acceptedIndices []int64

				if x == 0 {
					acceptedIndices = Indices
				} else {
					var leftIndex = OutputMatrix[z][y][x-1]
					var leftPosYID = IndexedPosY[leftIndex]

					acceptedIndices = NegYIndices[leftPosYID]
				}

				if y != 0 {
					var topIndex = OutputMatrix[z][y-1][x]
					var topNegXID = IndexedNegX[topIndex]
					var posXIndices = PosXIndices[topNegXID]

					acceptedIndices = GetDuplications(acceptedIndices, posXIndices)
				}

				if z != 0 {
					var bottomIndex = OutputMatrix[z-1][y][x]
					var topPosZID = IndexedPosZ[bottomIndex]
					var posZIndices = NegZIndices[topPosZID]

					acceptedIndices = GetDuplications(acceptedIndices, posZIndices)
				}

				OutputMatrix[z][y][x] = acceptedIndices[rand.Int63n(int64(len(acceptedIndices)))]
			}
		}
	}

	return
}

func LoadAssets() (err error) {
	for _, wfcStruct := range Structs {
		var texturePath = wfcStruct.Texture
		var textureName = filepath.Base(texturePath)

		var textureFile *os.File
		if textureFile, err = os.Open(texturePath); err != nil {
			return
		}

		var textureImage image.Image
		if textureImage, err = png.Decode(textureFile); err != nil {
			return
		}

		var imageSize = textureImage.Bounds().Size()
		var textureMatrix = make([][]*color.RGBA64, imageSize.X)

		var x, y int
		for x = 0; x < imageSize.X; x++ {
			textureMatrix[x] = make([]*color.RGBA64, imageSize.Y)

			for y = 0; y < imageSize.Y; y++ {
				var rgba64At = textureImage.(*image.RGBA).RGBA64At(x, y)

				textureMatrix[x][y] = &rgba64At
			}
		}

		IndexedAssets[textureName] = textureMatrix
	}

	return
}

func GenerateMapImage() (err error) {
	var outputImage = image.NewRGBA64(image.Rectangle{Max: image.Point{X: int(XSize * 16), Y: int(YSize * 16)}})

	var z, y, x int
	for z = 0; z < len(OutputMatrix); z++ {
		for y = 0; y < len(OutputMatrix[z]); y++ {
			for x = 0; x < len(OutputMatrix[z][y]); x++ {
				var wfcStruct = IndexedStructs[OutputMatrix[z][y][x]]
				var textureName = filepath.Base(wfcStruct.Texture)
				var textureMatrix = make([][]*color.RGBA64, len(IndexedAssets[textureName]))

				var textureX, textureY int
				for textureX = 0; textureX < len(IndexedAssets[textureName]); textureX++ {
					textureMatrix[textureX] = make([]*color.RGBA64, len(IndexedAssets[textureName][textureX]))

					for textureY = 0; textureY < len(IndexedAssets[textureName][textureX]); textureY++ {
						textureMatrix[textureX][textureY] = IndexedAssets[textureName][textureX][textureY]
					}
				}

				var i int8
				for {
					if i == wfcStruct.Rotation {
						break
					}

					textureMatrix = RotateRGBA64Matrix(textureMatrix)

					i = i + 1
				}

				for textureX = 0; textureX < len(textureMatrix); textureX++ {
					for textureY = 0; textureY < len(textureMatrix[textureX]); textureY++ {
						outputImage.SetRGBA64((x*len(textureMatrix))+textureX, (y*len(textureMatrix[textureX]))+textureY, *textureMatrix[textureX][textureY])
					}
				}
			}
		}
	}

	var imageFile *os.File
	if imageFile, err = os.Create(OutputMapImageFilePath); err != nil {
		return
	}

	err = png.Encode(imageFile, outputImage)

	return
}

func main() {
	var err error
	var startTime = time.Now()
	var leadTime = time.Now()

	log.Println(`loading input.json...`)

	if err = LoadInput(); err != nil {
		log.Fatalln(err)
	}

	log.Printf(`loading input took %s`, time.Since(leadTime).String())

	log.Println(`loading assets...`)

	leadTime = time.Now()

	if err = LoadAssets(); err != nil {
		log.Fatalln(err)
	}

	log.Printf(`loading assets took %s`, time.Since(leadTime).String())

	log.Println(`generating rotations...`)

	leadTime = time.Now()

	GenerateRotations()

	log.Printf(`generate rotations took %s`, time.Since(leadTime).String())

	log.Println(`generating map...`)

	leadTime = time.Now()

	if err = GenerateMap(); err != nil {
		log.Fatalln(err)
	}

	log.Printf(`generating map took %s`, time.Since(leadTime).String())

	log.Println(`generating map image...`)

	leadTime = time.Now()

	if err = GenerateMapImage(); err != nil {
		log.Fatalln(err)
	}

	log.Printf(`generating map image took %s`, time.Since(leadTime).String())

	log.Printf(`input structures %d`, len(Structs))
	log.Printf(`output structures %d`, len(IndexedStructs))
	log.Printf(`finished in %s`, time.Since(startTime).String())
}
