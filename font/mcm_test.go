package font

import (
	"fmt"
	"image/png"
	"os"
	"testing"
)

func TestParseFont(t *testing.T) {
	font := NewMcmFont("betaflight.mcm")

	for i := 0; i < 256; i++ {
		at, _ := font.GetCharAt(i)
		open, err := os.Create(fmt.Sprintf("/tmp/image-%v.png", i))
		if err != nil {
			t.Fatal("cannot create file", err)
		}
		defer open.Close()
		err = png.Encode(open, at)
		if err != nil {
			return
		}
	}
}
