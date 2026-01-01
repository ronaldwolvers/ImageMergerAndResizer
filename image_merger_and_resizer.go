package main

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"os/user"
	"regexp"
	"strings"
)

const fileRegexp = "\\w*\\.(\\w*)"

var compiledRegex, regexpCompileErr = regexp.Compile(fileRegexp)

func main() {

	if regexpCompileErr != nil {
		fmt.Println("Error compiling regex: ", regexpCompileErr.Error())
		os.Exit(1)
	}

	fmt.Println("Printing program arguments:")
	fmt.Println(os.Args)

	if len(os.Args) <= 1 {
		fmt.Println("Usage: image_merger_and_resizer <image_file>")
		os.Exit(1)
	}

	imageFile := os.Args[1]

	regexpMatches := compiledRegex.FindSubmatch([]byte(imageFile))
	if len(regexpMatches) <= 1 {
		fmt.Println("This file does not have an extension...")
		os.Exit(1)
	}

	fileExtension := string(regexpMatches[1])

	currentDir, _ := user.Current()
	homeDir := currentDir.HomeDir
	expandedImageFile := strings.ReplaceAll(imageFile, "~", homeDir)

	fmt.Println("Reading from file:", expandedImageFile)
	fileReader, err := os.Open(expandedImageFile)
	defer func() {
		if fileReader != nil {
			fileReader.Close()
		}
	}()

	if err != nil {
		fmt.Println("Error opening file: ", err.Error())
		os.Exit(1)
	}

	var image image.Image
	var decodeErr error
	if strings.ToLower(fileExtension) == "gif" {
		fmt.Println("Decoding gif...")
		image, decodeErr = gif.Decode(fileReader)
	} else if strings.ToLower(fileExtension) == "jpeg" {
		fmt.Println("Decoding jpeg...")
		image, decodeErr = jpeg.Decode(fileReader)
	} else if strings.ToLower(fileExtension) == "png" {
		fmt.Println("Decoding png...")
		image, decodeErr = png.Decode(fileReader)
	} else {
		fmt.Println("Unknown file extension:", fileExtension)
		os.Exit(1)
	}

	if decodeErr != nil {
		fmt.Println("Error decoding file: ", decodeErr.Error())
		os.Exit(1)
	}
	if image == nil {
		fmt.Println("No image was read from the specified file...")
		os.Exit(1)
	}

	fmt.Println("This image has color model: ", image.ColorModel())
}
