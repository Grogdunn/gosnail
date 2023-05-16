package osd

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

const headerBytes = 40
const frameBytes = 2124

const timestampBytes = 4
const bytesPerGlyph = 2
const gridWidth = 53
const gridHeight = 20

type FileOsd struct {
	Fc       string
	Duration time.Duration
	Frames   []Frame
}
type Frame struct {
	TimeMillis uint32
	Glyphs     []Glyph
}
type Glyph struct {
	Index    uint16
	position Position
}
type Position struct {
	X, Y uint32
}

func Parse(file string) *FileOsd {
	var toReturn FileOsd
	osdFile, err := os.Open(file)
	if err != nil {
		panic("cannot open file") //TODO
	}
	defer osdFile.Close()
	reader := bufio.NewReader(osdFile)
	header := make([]byte, headerBytes)
	readFrame, err := io.ReadFull(reader, header)
	if err != nil || readFrame < headerBytes {
		panic("readFrame header problem") //TODO
	}
	fc := header[0:4]
	toReturn.Fc = string(fc)
	toReturn.Frames = make([]Frame, 0)
	for {
		frame := make([]byte, frameBytes)
		readFrame, err = io.ReadFull(reader, frame)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Finito tutto")
			} else {
				fmt.Printf("Errore %v\n", err)
			}
			break
		}
		timeMillis := binary.LittleEndian.Uint32(frame[:timestampBytes])
		glyphReader := bytes.NewReader(frame[timestampBytes:])
		chunkId := 0
		glyphs := make([]Glyph, 0)
		for {
			glyphBytes := make([]byte, bytesPerGlyph)
			_, err = glyphReader.Read(glyphBytes)
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Errore glifo %v\n", err)
				}
				break
			}
			chunkId++
			var index uint16
			index = binary.LittleEndian.Uint16(glyphBytes)
			if index == 0x00 || index == 0x20 {
				continue
			}
			glyphs = append(glyphs, Glyph{
				Index: index,
				position: Position{
					X: uint32(chunkId % gridWidth),
					Y: uint32(chunkId / gridWidth),
				},
			})
		}

		f := Frame{
			TimeMillis: timeMillis,
			Glyphs:     glyphs,
		}
		toReturn.Frames = append(toReturn.Frames, f)
	}
	lastFrame := toReturn.Frames[len(toReturn.Frames)-1]
	frameInterval := float32(lastFrame.TimeMillis-toReturn.Frames[0].TimeMillis) / float32(len(toReturn.Frames)-1)
	toReturn.Duration = time.Duration(lastFrame.TimeMillis)*time.Millisecond + (time.Duration(frameInterval) * time.Millisecond)
	return &toReturn
}
