package osd

import (
	"fmt"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	open, err := os.Open("20230427_AvatarG0002.osd")
	if err != nil {
		panic("non trovato!")
	}
	parse := Parse(open)
	fmt.Printf("%v", parse)
}
