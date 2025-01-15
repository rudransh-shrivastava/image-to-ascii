package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const ASCII_ARRAY = "@%#*+=-:. "

var isColourful = false
var colourMatrix [][]color.Color

func displayHelp() {
	fmt.Println("Usage: go run main.go <image> <output-txt-name> <output-image-name>")
	if len(os.Args) == 2 {
		if os.Args[1] == "help" {
			fmt.Println("This program converts an image to an ascii art text file and an image")
			fmt.Println("Usage: go run main.go <image> <output-txt-name> <output-image-name> -<c>")
			fmt.Println("-c makes the image colourful")
			fmt.Println("Example: go run main.go image.jpg output.txt output.png")
		}
	}
}
func main() {
	if len(os.Args) == 5 {
		if os.Args[4] == "-c" {
			isColourful = true
		} else {
			displayHelp()
			return
		}
		if len(os.Args) < 4 {
			displayHelp()
			return
		}
	}
	imagePath := os.Args[1]
	outputTxtPath := os.Args[2]
	outputImgPath := os.Args[3]

	outputDirectory := "output/"
	if _, err := os.Stat(outputDirectory); os.IsNotExist(err) {
		os.Mkdir(outputDirectory, 0755)
	}
	outputTxtPath = outputDirectory + outputTxtPath
	outputImgPath = outputDirectory + outputImgPath

	err := imageToASCIIfile(imagePath, outputTxtPath, isColourful)
	if err != nil {
		panic(err)
	}

	err = txtToImage(outputTxtPath, outputImgPath, 9, 9, color.Black, color.White, isColourful)
	if err != nil {
		panic(err)
	}
	fmt.Printf("image successfully converted to ascii art \n output text file: %s \n output image file: %s \n", outputTxtPath, outputImgPath)
}

// convert a color to an ascii character
func colorToASCII(c color.Color) string {
	gray := color.GrayModel.Convert(c)
	grayValue := gray.(color.Gray).Y
	asciiIndex := int(grayValue) * (len(ASCII_ARRAY) - 1) / 255
	return string(ASCII_ARRAY[asciiIndex])
}

// convert an image to an ascii text file

func imageToASCIIfile(imagePath, outputTxtPath string, isColourful bool) error {
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Println("Error opening image file")
		return err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Println("Error decoding the image file")
		return err
	}

	outputTxtFile, err := os.Create(outputTxtPath)
	if err != nil {
		fmt.Println("Error creating the output file")
		return err
	}
	defer outputTxtFile.Close()

	bounds := img.Bounds()

	colourMatrix = make([][]color.Color, bounds.Max.Y)
	if isColourful {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			colourMatrix[y] = make([]color.Color, bounds.Max.X)
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				colourMatrix[y][x] = img.At(x, y)
			}
		}
	}
	// write ascii chars to txt file
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			asciiChar := colorToASCII(img.At(x, y))
			_, err := outputTxtFile.WriteString(asciiChar)
			if err != nil {
				fmt.Println("Error while writing to output text file")
				return err
			}
		}
		outputTxtFile.WriteString("\n")
	}
	return nil
}

// convert a text file to an image
func txtToImage(inputFileName string, outputFileName string, charWidth, charHeight int, textColor, bgColor color.Color, isColourful bool) error {
	file, err := os.Open(inputFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	imgWidth := len(lines[0]) * charWidth
	imgHeight := len(lines) * charHeight

	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bgColor}, image.Point{}, draw.Src)

	point := fixed.Point26_6{}
	face := basicfont.Face7x13
	for y, line := range lines {
		for x, char := range line {
			point.X = fixed.I(x * charWidth)
			point.Y = fixed.I((y + 1) * charHeight)
			if isColourful {
				textColor = colourMatrix[y][x]
			}
			d := &font.Drawer{
				Dst:  img,
				Src:  image.NewUniform(textColor),
				Face: face,
				Dot:  point,
			}
			d.DrawString(string(char))
		}
	}

	outputFile, err := os.Create(outputFileName)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	return png.Encode(outputFile, img)
}
