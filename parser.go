package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
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
