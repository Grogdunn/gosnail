package font

import (
	"bufio"
	"errors"
	"image"
	"image/color"
	"os"
)

type McmFont struct {
	Font
	file    string
	charset []image.Image //TODO
}

func NewMcmFont(file string) *McmFont {
	return &McmFont{file: file}
}

const charHeight = 18
const charWidth = 12
const lineGarbage = 10
const nibbleLen = 2

func (m McmFont) GetCharAt(position int) (image.Image, error) {
	if m.charset == nil {
		m.charset = make([]image.Image, 0)

		fontFile, err := os.Open(m.file)
		if err != nil {
			panic("cannot open file")
		}
		defer fontFile.Close()

		scanner := bufio.NewScanner(fontFile)
		scanner.Split(bufio.ScanLines)

		if scanner.Scan() {
			line := scanner.Text()
			if line != "MAX7456" {
				panic("header row mismatch expected MAX7456")
			}
		}
		for scanner.Scan() {
			rgba := image.NewRGBA(image.Rect(0, 0, charWidth, charHeight))
			m.charset = append(m.charset, rgba)

			var line string
			for i := 0; i < charWidth*charHeight*nibbleLen/8; i++ {
				line = line + scanner.Text()
				if !scanner.Scan() {
					break
				}
			}

			if len(line) < charWidth*charHeight*nibbleLen {
				panic("incomplete line")
			}

			for charY := 0; charY < charHeight; charY++ {
				for charX := 0; charX < charWidth; charX++ {
					charOffset := (charY * charWidth * nibbleLen) + (charX * nibbleLen)
					s := line[charOffset : charOffset+nibbleLen]
					switch s {
					case "00":
						rgba.Set(charX, charY, color.Black)
					case "01":
						rgba.Set(charX, charY, color.Transparent)
					case "10":
						rgba.Set(charX, charY, color.White)
					case "11": //White transparent
						rgba.Set(charX, charY, color.RGBA{
							R: 255,
							G: 255,
							B: 255,
							A: 0,
						})
					}
				}

			}
			for i := 0; i < lineGarbage-1; i++ {
				scanner.Scan()
			}
		}
	}
	if position > len(m.charset) {
		return nil, errors.New("char not found")
	}
	return m.charset[position], nil

}
