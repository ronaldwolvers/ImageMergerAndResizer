package main

import (
	"fmt"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
)

func main() {

	fmt.Println("Printing program arguments:")
	fmt.Println(os.Args)

	return

	gif.Decode(nil)
	jpeg.Decode(nil)
	png.Decode(nil)
}
