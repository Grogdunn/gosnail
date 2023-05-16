package font

import (
	"image"
)

type Font interface {
	GetCharAt(position int) (image.Image, error)
}
