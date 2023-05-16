package osd

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	parse := Parse("20230427_AvatarG0002.osd")

	fmt.Printf("%v", parse)
}
