package main

import (
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"os"
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

	Indices []int64

	// OutputMatrix [Z, [Y, [X]]]
	OutputMatrix [][][]int64

	XSize int64 // Width Size
	YSize int64 // Length Size
	ZSize int64 // Height Size
)

func main() {
	var timeStart = time.Now()

	var err error

	// My current screen size divided by the size of the tiles (16x16)
	flag.Int64Var(&XSize, "x", 2560/16, `X - Width Size`)
	flag.Int64Var(&YSize, "y", 1440/16, `Y - Length Size`)
	flag.Int64Var(&ZSize, "z", 1, `Z - Height Size`)

	flag.Parse()

	var structsFile *os.File
	structsFile, err = os.Open(`input.json`)

	if err != nil {
		log.Fatalln(err)
	}

	var jsonDecoder = json.NewDecoder(structsFile)
	err = jsonDecoder.Decode(&Structs)

	if err != nil {
		log.Fatalln(err)
	}

	rand.Seed(time.Now().Unix() / 10 * 102)

	GenerateRotations()

	OutputMatrix = make([][][]int64, ZSize)

	var z int64
	for z = 0; z < ZSize; z++ {
		OutputMatrix[z] = make([][]int64, YSize)

		var y int64
		for y = 0; y < YSize; y++ {
			OutputMatrix[z][y] = make([]int64, XSize)

			var x int64
			for x = 0; x < XSize; x++ {
				var leftPosYID int64
				var topNegXID int64
				var topPosZID int64
				var acceptedIndices []int64

				if x == 0 {
					acceptedIndices = Indices
				} else {
					var leftIndex = OutputMatrix[z][y][x-1]
					leftPosYID = IndexedPosY[leftIndex]

					acceptedIndices = NegYIndices[leftPosYID]
				}

				if y != 0 {
					var topIndex = OutputMatrix[z][y-1][x]
					topNegXID = IndexedNegX[topIndex]
					var posXIndices = PosXIndices[topNegXID]

					acceptedIndices = GetDuplications(acceptedIndices, posXIndices)
				}

				if z != 0 {
					var bottomIndex = OutputMatrix[z-1][y][x]
					topPosZID = IndexedPosZ[bottomIndex]
					var posZIndices = NegZIndices[topPosZID]

					acceptedIndices = GetDuplications(acceptedIndices, posZIndices)
				}

				OutputMatrix[z][y][x] = acceptedIndices[rand.Int63n(int64(len(acceptedIndices)))]
			}
		}
	}

	var outputFile *os.File
	outputFile, err = os.Create(`output.json`)

	if err != nil {
		log.Fatalln(err)
	}

	var jsonEncoder = json.NewEncoder(outputFile)
	err = jsonEncoder.Encode(struct {
		IndexedStructs map[int64]WFCStruct `json:"indexed_structs"`
		OutputMatrix   [][][]int64         `json:"output_matrix"`
	}{
		IndexedStructs: IndexedStructs,
		OutputMatrix:   OutputMatrix,
	})

	if err != nil {
		log.Fatalln(err)
	}

	log.Printf(`Input structures %d`, len(Structs))
	log.Printf(`Output structures %d`, len(IndexedStructs))
	log.Printf(`Finished in %s`, time.Since(timeStart).String())
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

	var _, ok = PosXIndices[wfcStruct.PosX]

	if !ok {
		PosXIndices[wfcStruct.PosX] = make([]int64, 0, 10)
	}

	PosXIndices[wfcStruct.PosX] = append(PosXIndices[wfcStruct.PosX], i)

	_, ok = PosYIndices[wfcStruct.PosY]

	if !ok {
		PosYIndices[wfcStruct.PosY] = make([]int64, 0, 10)
	}

	PosYIndices[wfcStruct.PosY] = append(PosYIndices[wfcStruct.PosY], i)

	_, ok = NegXIndices[wfcStruct.NegX]

	if !ok {
		NegXIndices[wfcStruct.NegX] = make([]int64, 0, 10)
	}

	NegXIndices[wfcStruct.NegX] = append(NegXIndices[wfcStruct.NegX], i)

	_, ok = NegYIndices[wfcStruct.NegY]

	if !ok {
		NegYIndices[wfcStruct.NegY] = make([]int64, 0, 10)
	}

	NegYIndices[wfcStruct.NegY] = append(NegYIndices[wfcStruct.NegY], i)

	_, ok = PosZIndices[wfcStruct.PosZ]

	if !ok {
		PosZIndices[wfcStruct.PosZ] = make([]int64, 0, 10)
	}

	PosZIndices[wfcStruct.PosZ] = append(PosZIndices[wfcStruct.PosZ], i)

	_, ok = NegZIndices[wfcStruct.NegZ]

	if !ok {
		NegZIndices[wfcStruct.NegZ] = make([]int64, 0, 10)
	}

	NegZIndices[wfcStruct.NegZ] = append(NegZIndices[wfcStruct.NegZ], i)

	return i + 1
}

func GenerateRotations() {
	var index int64 = 0

	for _, wfcStruct := range Structs {
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
			PosX:          wfcStruct.NegY,
			PosY:          wfcStruct.PosX,
			NegX:          wfcStruct.PosY,
			NegY:          wfcStruct.NegX,
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
			PosX:          wfcStruct.PosY,
			PosY:          wfcStruct.NegX,
			NegX:          wfcStruct.NegY,
			NegY:          wfcStruct.PosX,
			PosZ:          wfcStruct.PosZ,
			NegZ:          wfcStruct.NegZ,
		})
	}
}
