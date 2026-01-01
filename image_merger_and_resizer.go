package main

import (
	"bufio"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"os/user"
	"regexp"
	"strconv"
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

	if len(os.Args) <= 3 || (strings.ToLower(os.Args[2]) != "scale" && strings.ToLower(os.Args[2]) != "merge") {
		fmt.Println("Usage: image_merger_and_resizer <image_file> <scale|merge> [scale_factor|merge_file] [output_file]")
		os.Exit(1)
	}

	imageFile := os.Args[1]
	programCommand := os.Args[2]

	var scaleFactor = -1
	var mergeFile = ""
	var expandedMergeFile = ""
	var err error

	if programCommand == "scale" {
		scaleFactor, err = strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Println("Error converting <scale_factor> to a number. Error: ", err.Error())
			os.Exit(1)
		}
	} else if programCommand == "merge" {
		mergeFile = os.Args[3]
		currentDir, _ := user.Current()
		homeDir := currentDir.HomeDir
		expandedMergeFile = strings.ReplaceAll(mergeFile, "~", homeDir)
	}

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

	var decodedImage image.Image
	var decodeErr error
	if strings.ToLower(fileExtension) == "gif" {
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

	fmt.Println("This image has color model: ", decodedImage.ColorModel())

	var resultImage image.Image
	var encodeErr error
	if programCommand == "scale" {
		resultImage, encodeErr = scaleImage(decodedImage, scaleFactor)
	} else if programCommand == "merge" {
		resultImage, encodeErr = mergeImage(decodedImage, expandedMergeFile)
	}

	if encodeErr != nil {
		fmt.Println("Error encoding image: ", encodeErr.Error())
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
	outputFile, err = os.Open(outputFilePath)
	if err != nil {
		fmt.Println("Error opening file: ", err.Error())
		os.Exit(1)
	}
	outputFileWriter := bufio.NewWriter(outputFile)

	var encodingError error
	if strings.ToLower(outputFileExtension) == "gif" {
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
		fmt.Println("Not applying any encoding...")
	}

	if encodingError != nil {
		fmt.Println("Error encoding file: ", encodingError.Error())
	}
}

func scaleImage(image image.Image, scaleFactor int) (image.Image, error) {

	fmt.Println("Scaling image...")
	fmt.Println("ScaleFactor: ", scaleFactor)
	return nil, nil
}

func mergeImage(image image.Image, mergeFilePath string) (image.Image, error) {

	fmt.Println("Merging image...")
	fmt.Println("Merge-file path: ", mergeFilePath)
	return nil, nil
}
