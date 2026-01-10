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

// Set this flag to 'false' to disable all logging.
var enableLogging = true

type Logger interface {
	Printf(format string, a ...any)
	Println(a ...any)
}

type loggerStruct struct{}

func (logger loggerStruct) Println(a ...any) {
	if !enableLogging {
		return
	}
	fmt.Println(a...)
}
func (logger loggerStruct) Printf(format string, a ...any) {
	if !enableLogging {
		return
	}
	fmt.Printf(format, a...)
}

var logger Logger = loggerStruct{}

const fileRegexp = "\\w*\\.(\\w*)"
const mergeRegexp = "merge(\\:(\\d*))(\\:(\\d*))"

var compiledFileRegex, fileRegexpCompileErr = regexp.Compile(fileRegexp)
var compiledMergeRegex, mergeRegexpCompileErr = regexp.Compile(mergeRegexp)

func main() {

	if fileRegexpCompileErr != nil {
		logger.Println("Error compiling regex: ", fileRegexpCompileErr.Error())
		os.Exit(1)
	}

	if mergeRegexpCompileErr != nil {
		logger.Println("Error compiling regex: ", mergeRegexpCompileErr.Error())
		os.Exit(1)
	}

	if len(os.Args) <= 3 || (strings.ToLower(os.Args[2]) != "scale" && !compiledMergeRegex.MatchString(os.Args[2])) {
		logger.Println("Usage: image_merger_and_resizer <image_file> <scale|merge> [<scale_factor>|<merge_file>[:offsetX][:offsetY]]" +
			" [output_file]")
		os.Exit(1)
	}

	if len(os.Args) == 4 {
		//Do not log when writing to stdout.
		enableLogging = false
	}

	var imageFile = ""
	imageFile = os.Args[1]
	var programCommand = ""
	programCommand = os.Args[2]

	var scaleFactor = -1
	var mergeFile = ""
	var expandedMergeFile = ""
	var err error = nil

	if programCommand == "scale" {
		scaleFactor, err = strconv.Atoi(os.Args[3])
		if err != nil {
			logger.Println("Error converting <scale_factor> to a number. Error: ", err.Error())
			os.Exit(1)
		}
	} else if compiledMergeRegex.MatchString(programCommand) {
		mergeFile = os.Args[3]
		expandedMergeFile = expandFilePath(mergeFile)
	}

	var regexpMatches [][]byte = nil
	regexpMatches = compiledFileRegex.FindSubmatch([]byte(imageFile))
	if len(regexpMatches) <= 1 {
		logger.Println("This file does not have an extension...")
		os.Exit(1)
	}

	var fileExtension = ""
	fileExtension = string(regexpMatches[1])

	var expandedImageFile = ""
	expandedImageFile = expandFilePath(imageFile)

	var fileReader *os.File
	logger.Println("Reading from file:", expandedImageFile)
	fileReader, err = os.Open(expandedImageFile)
	defer func() {
		if fileReader != nil {
			err = fileReader.Close()
			if err != nil {
				logger.Println("Error closing file: ", err.Error())
			}
		}
	}()

	if err != nil {
		logger.Println("Error opening file: ", err.Error())
		os.Exit(1)
	}

	var decodedImage image.Image
	var decodeErr error
	if strings.ToLower(fileExtension) == "bmp" {
		logger.Println("Decoding bmp...")
		decodedImage, decodeErr = bmp.Decode(fileReader)
	} else if strings.ToLower(fileExtension) == "gif" {
		logger.Println("Decoding gif...")
		decodedImage, decodeErr = gif.Decode(fileReader)
	} else if strings.ToLower(fileExtension) == "jpeg" {
		logger.Println("Decoding jpeg...")
		decodedImage, decodeErr = jpeg.Decode(fileReader)
	} else if strings.ToLower(fileExtension) == "png" {
		logger.Println("Decoding png...")
		decodedImage, decodeErr = png.Decode(fileReader)
	} else {
		logger.Println("Unknown file extension:", fileExtension)
		os.Exit(1)
	}

	if decodeErr != nil {
		logger.Println("Error decoding file: ", decodeErr.Error())
		os.Exit(1)
	}

	if decodedImage == nil {
		logger.Println("No image was read from the specified file...")
		os.Exit(1)
	}

	logger.Printf("This image has color model: %X\n", decodedImage.ColorModel())

	var resultImage image.Image = nil
	var processErr error = nil
	if programCommand == "scale" {
		resultImage, processErr = scaleImage(decodedImage, scaleFactor)
	} else if compiledMergeRegex.MatchString(os.Args[2]) {
		var offsetX = 0
		var offsetY = 0
		regexpMatches = compiledMergeRegex.FindSubmatch([]byte(os.Args[2]))
		offsetX, _ = strconv.Atoi(string(regexpMatches[2]))
		offsetY, _ = strconv.Atoi(string(regexpMatches[4]))
		resultImage, processErr = mergeImage(decodedImage, expandedMergeFile, offsetX, offsetY)
	}

	if processErr != nil {
		logger.Println("Error encoding image: ", processErr.Error())
		os.Exit(1)
	} else if resultImage == nil {
		logger.Println("No image was encoded...")
		os.Exit(1)
	}

	var outputFilePath string
	var outputFileExtension string
	if len(os.Args) >= 5 {
		outputFilePath = os.Args[4]
		regexpMatches = compiledFileRegex.FindSubmatch([]byte(outputFilePath))
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
			logger.Println("Error opening file: ", err.Error())
			os.Exit(1)
		}
	}
	var outputFileWriter *bufio.Writer
	outputFileWriter = bufio.NewWriter(outputFile)

	var encodingError error
	if strings.ToLower(fileExtension) == "bmp" {
		logger.Println("Encoding bmp...")
		encodingError = bmp.Encode(outputFileWriter, resultImage)
	} else if strings.ToLower(outputFileExtension) == "gif" {
		logger.Println("Encoding gif...")
		encodingError = gif.Encode(outputFileWriter, resultImage, nil)
	} else if strings.ToLower(outputFileExtension) == "jpeg" {
		logger.Println("Encoding jpeg...")
		encodingError = jpeg.Encode(outputFileWriter, resultImage, nil)
	} else if strings.ToLower(outputFileExtension) == "png" {
		logger.Println("Encoding png...")
		encodingError = png.Encode(outputFileWriter, resultImage)
	} else {
		logger.Println("Unknown file extension:", fileExtension)
		encodingError = errors.New("Unknown file extension" + fileExtension)
	}

	if encodingError != nil {
		logger.Println("Error encoding file: ", encodingError.Error())
		os.Exit(1)
	}

	if compiledMergeRegex.MatchString(programCommand) {
		logger.Println("Successfully merged two images!")
	}

	if programCommand == "scale" {
		logger.Println("Successfully scaled image!")
	}
}

func expandFilePath(path string) string {
	var currentDir *user.User = nil
	currentDir, _ = user.Current()
	var homeDir = ""
	homeDir = currentDir.HomeDir
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
	var scaledMax image.Point
	scaledMax = image.Point{X: s.originalImage.Bounds().Max.X / s.scaleFactor, Y: s.originalImage.Bounds().Max.Y / s.scaleFactor}
	return image.Rectangle{Min: s.originalImage.Bounds().Min, Max: scaledMax}
}
func (s scaledImage) At(x, y int) color.Color {
	return s.originalImage.At(x*s.scaleFactor, y*s.scaleFactor)
}

type mergedImage struct {
	imageLeft         image.Image
	imageRight        image.Image
	imageRightOffsetX int
	imageRightOffsetY int
}

func (m mergedImage) ColorModel() color.Model {
	return m.imageLeft.ColorModel()
}
func (m mergedImage) Bounds() image.Rectangle {
	return m.imageLeft.Bounds()
}

func (m mergedImage) At(x, y int) color.Color {
	var leftColor color.Color
	leftColor = m.imageLeft.At(x, y)

	if x >= m.imageRight.Bounds().Min.X+m.imageRightOffsetX && x <= m.imageRight.Bounds().Max.X-m.imageRightOffsetX &&
		y >= m.imageRight.Bounds().Min.Y+m.imageRightOffsetY && y <= m.imageRight.Bounds().Max.Y-m.imageRightOffsetY {

		var rightColor color.Color
		rightColor = m.imageRight.At(x, y)
		var rightColorConverted color.Color
		rightColorConverted = m.imageLeft.ColorModel().Convert(rightColor)

		var rightColorAlpha uint32
		_, _, _, rightColorAlpha = rightColorConverted.RGBA()

		if rightColorAlpha > 0 {
			return rightColorConverted
		}
	}

	return leftColor
}

func scaleImage(image image.Image, scaleFactor int) (image.Image, error) {

	logger.Println("Scaling image...")
	logger.Println("Scale factor: ", scaleFactor)

	return scaledImage{image, scaleFactor}, nil
}

func mergeImage(baseImage image.Image, mergeFilePath string, offsetX int, offsetY int) (image.Image, error) {

	logger.Println("Merging image...")
	logger.Println("Merge-file path: ", mergeFilePath)

	var regexpMatches [][]byte
	regexpMatches = compiledFileRegex.FindSubmatch([]byte(mergeFilePath))
	if len(regexpMatches) <= 1 {
		logger.Println("This file does not have an extension...")
		os.Exit(1)
	}
	var fileExtension = ""
	fileExtension = string(regexpMatches[1])

	logger.Println("Reading from file:", mergeFilePath)
	var fileReader *os.File = nil
	var err error = nil
	fileReader, err = os.Open(mergeFilePath)
	defer func() {
		if fileReader != nil {
			err = fileReader.Close()
			if err != nil {
				logger.Println("Error closing file: ", err.Error())
			}
		}
	}()

	var decodedMergeImage image.Image
	var decodeErr error
	if strings.ToLower(fileExtension) == "gif" {
		logger.Println("Decoding gif...")
		decodedMergeImage, decodeErr = gif.Decode(fileReader)
	} else if strings.ToLower(fileExtension) == "jpeg" || strings.ToLower(fileExtension) == "jpg" {
		logger.Println("Decoding jpeg...")
		decodedMergeImage, decodeErr = jpeg.Decode(fileReader)
	} else if strings.ToLower(fileExtension) == "png" {
		logger.Println("Decoding png...")
		decodedMergeImage, decodeErr = png.Decode(fileReader)
	} else {
		logger.Println("Unknown file extension:", fileExtension)
		return nil, errors.New("Unknown file extension: " + fileExtension)
	}

	if decodeErr != nil {
		logger.Println("Error decoding file merge file: ", decodeErr.Error())
		return nil, decodeErr
	}

	var mergedImage_ mergedImage
	mergedImage_ = mergedImage{baseImage, decodedMergeImage, offsetX, offsetY}

	return mergedImage_, nil
}
