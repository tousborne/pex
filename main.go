package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
)

// Parse the given image data and return the three most common colored pixels in the
// form of "#RRGGBB" with hex digits corresponding to red, green, and blue.
func parseImage(data image.Image) (string, string, string) {
	histogram := make(map[string]uint64)
	var first, second, third uint64
	var firstRGB, secondRGB, thirdRGB string

	// Recommended by the image library to loop on y then x for best performance.
	bounds := data.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var r, g, b uint8

			pixel := data.At(x, y)

			// Extract the RGB values from the different color model types.
			switch pixel.(type) {
			case color.YCbCr:
				temp := pixel.(color.YCbCr)
				r, g, b = color.YCbCrToRGB(temp.Y, temp.Cb, temp.Cr)

			case color.NRGBA:
				temp := pixel.(color.NRGBA)
				r = temp.R
				g = temp.G
				b = temp.B
			}

			// Convert the color to hex representation
			hex := fmt.Sprintf("#%02X%02X%02X", r, g, b)

			count := histogram[hex] + 1
			histogram[hex] = count

			// Check the current pixel against the leaders and either adjust counts or
			// flip leaderboard positions as needed.
			if count > first {
				if firstRGB == hex {
					first = count
					continue
				}

				second = first
				secondRGB = firstRGB
				first = count
				firstRGB = hex

				continue
			}

			if count > second {
				if secondRGB == hex {
					second = count
					continue
				}

				third = second
				thirdRGB = secondRGB
				second = count
				secondRGB = hex

				continue
			}

			if count > third {
				third = count
				thirdRGB = hex

				continue
			}
		}
	}

	// If there aren't three colors, fill in with the other leaders.
	if secondRGB == "" {
		secondRGB = firstRGB
	}

	if thirdRGB == "" {
		thirdRGB = secondRGB
	}

	return firstRGB, secondRGB, thirdRGB
}

// Download an image image from the given url and return the decoded image.
func downloadImage(url string) (image.Image, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Error downloading image: %s", err)
	}

	image, _, err := image.Decode(response.Body)
	return image, err
}

// Run the program logic.
func run(inPath, outPath string) error {
	inFile, err := os.Open(inPath)
	if err != nil {
		return fmt.Errorf("Failed to open %s: %s", inPath, err)
	}
	defer inFile.Close()

	scanner := bufio.NewScanner(inFile)

	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("Failed to open %s: %s", outPath, err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)

	// Defaults to tokenizing on newlines
	for scanner.Scan() {
		url := scanner.Text()
		image, err := downloadImage(url)
		if err != nil {
			return fmt.Errorf("Error scanning %s: %s", url, err)
		}

		first, second, third := parseImage(image)
		csv := []string{url, first, second, third}
		//fmt.Printf("%v\n", csv)

		err = writer.Write(csv)
		if err != nil {
			return fmt.Errorf("Error writing to file: %s", err)
		}
	}

	writer.Flush()

	return nil
}

// Print the program usage string
func usage() {
	fmt.Printf("usage: %s [] file_list out_file\n", os.Args[0])
	fmt.Printf("\tParses the file_list, which is a list of URLs to images, downloads all\n")
	fmt.Printf("\timages and determines the three most common RGB hex colors.  A list of\n")
	fmt.Printf("\tthese data points is then written to a out_file.\n")
}

// Parse the arguments and call run().
func main() {
	flag.Parse()

	args := flag.Args()

	if len(args) != 2 {
		usage()
		os.Exit(-1)
	}

	err := run(args[0], args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(-1)
	}
}
