package main

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"os/user"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/image/bmp"
)

const fileRegexp = "\\w*\\.(\\w*)"

var compiledRegex, regexpCompileErr = regexp.Compile(fileRegexp)

func main() {

	if regexpCompileErr != nil {
		fmt.Println("Error compiling regex: ", regexpCompileErr.Error())
		os.Exit(1)
	}

	if len(os.Args) <= 3 || (strings.ToLower(os.Args[2]) != "scale" && strings.ToLower(os.Args[2]) != "merge") {
		fmt.Println("Usage: image_merger_and_resizer <image_file> <scale|merge> [scale_factor|merge_file] [output_file]")
		os.Exit(1)
	}

	imageFile := os.Args[1]
	programCommand := os.Args[2]

	var scaleFactor = -1
	var mergeFile = ""
	var expandedMergeFile = ""
	var err error = nil

	if programCommand == "scale" {
		scaleFactor, err = strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Println("Error converting <scale_factor> to a number. Error: ", err.Error())
			os.Exit(1)
		}
	} else if programCommand == "merge" {
		mergeFile = os.Args[3]
		expandedMergeFile = expandFilePath(mergeFile)
	}

	regexpMatches := compiledRegex.FindSubmatch([]byte(imageFile))
	if len(regexpMatches) <= 1 {
		fmt.Println("This file does not have an extension...")
		os.Exit(1)
	}

	fileExtension := string(regexpMatches[1])

	expandedImageFile := expandFilePath(imageFile)

	fmt.Println("Reading from file:", expandedImageFile)
	fileReader, err := os.Open(expandedImageFile)
	defer func() {
		if fileReader != nil {
			err = fileReader.Close()
			if err != nil {
				fmt.Println("Error closing file: ", err.Error())
			}
		}
	}()

	if err != nil {
		fmt.Println("Error opening file: ", err.Error())
		os.Exit(1)
	}

	var decodedImage image.Image
	var decodeErr error
	if strings.ToLower(fileExtension) == "bmp" {
		fmt.Println("Decoding bmp...")
		decodedImage, decodeErr = bmp.Decode(fileReader)
	} else if strings.ToLower(fileExtension) == "gif" {
		fmt.Println("Decoding gif...")
		decodedImage, decodeErr = gif.Decode(fileReader)
	} else if strings.ToLower(fileExtension) == "jpeg" {
		fmt.Println("Decoding jpeg...")
		decodedImage, decodeErr = jpeg.Decode(fileReader)
	} else if strings.ToLower(fileExtension) == "png" {
		fmt.Println("Decoding png...")
		decodedImage, decodeErr = png.Decode(fileReader)
	} else {
		fmt.Println("Unknown file extension:", fileExtension)
		os.Exit(1)
	}

	if decodeErr != nil {
		fmt.Println("Error decoding file: ", decodeErr.Error())
		os.Exit(1)
	}

	if decodedImage == nil {
		fmt.Println("No image was read from the specified file...")
		os.Exit(1)
	}

	fmt.Printf("This image has color model: %X\n", decodedImage.ColorModel())

	var resultImage image.Image = nil
	var processErr error = nil
	if programCommand == "scale" {
		resultImage, processErr = scaleImage(decodedImage, scaleFactor)
	} else if programCommand == "merge" {
		resultImage, processErr = mergeImage(decodedImage, expandedMergeFile)
	}

	if processErr != nil {
		fmt.Println("Error encoding image: ", processErr.Error())
		os.Exit(1)
	} else if resultImage == nil {
		fmt.Println("No image was encoded...")
		os.Exit(1)
	}

	var outputFilePath string
	var outputFileExtension string
	if len(os.Args) >= 5 {
		outputFilePath = os.Args[4]
		regexpMatches = compiledRegex.FindSubmatch([]byte(outputFilePath))
		if len(regexpMatches) > 1 {
			outputFileExtension = string(regexpMatches[1])
		}
	}
	var outputFile *os.File
	if outputFilePath == "" {
		outputFile = os.Stdout
		fileExtension = "bmp"
	} else {
		outputFile, err = os.OpenFile(expandFilePath(outputFilePath), os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			fmt.Println("Error opening file: ", err.Error())
			os.Exit(1)
		}
	}
	outputFileWriter := bufio.NewWriter(outputFile)

	var encodingError error
	if strings.ToLower(fileExtension) == "bmp" {
		fmt.Println("Encoding bmp...")
		encodingError = bmp.Encode(outputFileWriter, resultImage)
	} else if strings.ToLower(outputFileExtension) == "gif" {
		fmt.Println("Encoding gif...")
		encodingError = gif.Encode(outputFileWriter, resultImage, nil)
	} else if strings.ToLower(outputFileExtension) == "jpeg" {
		fmt.Println("Encoding jpeg...")
		encodingError = jpeg.Encode(outputFileWriter, resultImage, nil)
	} else if strings.ToLower(outputFileExtension) == "png" {
		fmt.Println("Encoding png...")
		encodingError = png.Encode(outputFileWriter, resultImage)
	} else {
		fmt.Println("Unknown file extension:", fileExtension)
		encodingError = errors.New("Unknown file extension" + fileExtension)
	}

	if encodingError != nil {
		fmt.Println("Error encoding file: ", encodingError.Error())
		os.Exit(1)
	}

	if programCommand == "merge" && len(os.Args) > 4 {
		fmt.Println("Successfully merged two images!")
	}

	if programCommand == "scale" && len(os.Args) > 4 {
		fmt.Println("Successfully scaled image!")
	}
}

func expandFilePath(path string) string {
	currentDir, _ := user.Current()
	homeDir := currentDir.HomeDir
	return strings.ReplaceAll(path, "~", homeDir)
}

type scaledImage struct {
	originalImage image.Image
	scaleFactor   int
}

func (s scaledImage) ColorModel() color.Model {
	return s.originalImage.ColorModel()
}
func (s scaledImage) Bounds() image.Rectangle {
	scaledMax := image.Point{X: s.originalImage.Bounds().Max.X / s.scaleFactor, Y: s.originalImage.Bounds().Max.Y / s.scaleFactor}
	return image.Rectangle{Min: s.originalImage.Bounds().Min, Max: scaledMax}
}
func (s scaledImage) At(x, y int) color.Color {
	return s.originalImage.At(x*s.scaleFactor, y*s.scaleFactor)
}

type mergedImage struct {
	imageLeft  image.Image
	imageRight image.Image
}

func (m mergedImage) ColorModel() color.Model {
	return m.imageLeft.ColorModel()
}
func (m mergedImage) Bounds() image.Rectangle {
	return image.Rectangle{Min: m.imageLeft.Bounds().Min, Max: image.Point{X: max(m.imageLeft.Bounds().Max.X, m.imageRight.Bounds().Max.X), Y: max(m.imageLeft.Bounds().Max.Y, m.imageRight.Bounds().Max.Y)}}
}
func (m mergedImage) At(x, y int) color.Color {
	leftARGB := m.imageLeft.At(x, y)
	rightARGB := m.imageRight.At(x, y)
	_, _, _, a2 := rightARGB.RGBA()
	if a2 != 0 {
		return rightARGB
	}

	return leftARGB
}

func scaleImage(image image.Image, scaleFactor int) (image.Image, error) {

	fmt.Println("Scaling image...")
	fmt.Println("Scale factor: ", scaleFactor)

	return scaledImage{image, scaleFactor}, nil
}

func mergeImage(baseImage image.Image, mergeFilePath string) (image.Image, error) {

	fmt.Println("Merging image...")
	fmt.Println("Merge-file path: ", mergeFilePath)

	regexpMatches := compiledRegex.FindSubmatch([]byte(mergeFilePath))
	if len(regexpMatches) <= 1 {
		fmt.Println("This file does not have an extension...")
		os.Exit(1)
	}
	fileExtension := string(regexpMatches[1])

	fmt.Println("Reading from file:", mergeFilePath)
	fileReader, err := os.Open(mergeFilePath)
	defer func() {
		if fileReader != nil {
			err = fileReader.Close()
			if err != nil {
				fmt.Println("Error closing file: ", err.Error())
			}
		}
	}()

	var decodedMergeImage image.Image
	var decodeErr error
	if strings.ToLower(fileExtension) == "gif" {
		fmt.Println("Decoding gif...")
		decodedMergeImage, decodeErr = gif.Decode(fileReader)
	} else if strings.ToLower(fileExtension) == "jpeg" {
		fmt.Println("Decoding jpeg...")
		decodedMergeImage, decodeErr = jpeg.Decode(fileReader)
	} else if strings.ToLower(fileExtension) == "png" {
		fmt.Println("Decoding png...")
		decodedMergeImage, decodeErr = png.Decode(fileReader)
	} else {
		fmt.Println("Unknown file extension:", fileExtension)
		return nil, errors.New("Unknown file extension: " + fileExtension)
	}

	if decodeErr != nil {
		fmt.Println("Error decoding file merge file: ", decodeErr.Error())
		return nil, decodeErr
	}

	mergedImage := mergedImage{baseImage, decodedMergeImage}

	return mergedImage, nil
}
