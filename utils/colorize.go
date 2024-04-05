package utils

import (
	"fmt"
	"strings"
)

var colorMap = map[string]string{
	"black":   "#000000",
	"red":     "#FF0000",
	"green":   "#00FF00",
	"yellow":  "#FFFF00",
	"blue":    "#0000FF",
	"magenta": "#FF00FF",
	"cyan":    "#00FFFF",
	"white":   "#FFFFFF",
	"gray":    "#808080",
	"purple":  "#800080",
	"orange":  "#FFA500",
	"brown":   "#A52A2A",
	"pink":    "#FFC0CB",
	"lime":    "#00FF00",
	"teal":    "#008080",
	"indigo":  "#4B0082",
}

func CPrint(color string, a ...interface{}) {
	color = strings.ToLower(color)
	if hexCode, ok := colorMap[color]; ok {
		CPrintHex(hexCode, a...)
	} else {
		fmt.Println(a...)
	}
}

func CPrintHex(hexColor string, a ...interface{}) {
	colorCode := "\033[38;2;" + hexToRGB(hexColor) + "m"
	resetCode := "\033[0m"
	fmt.Print(colorCode)
	fmt.Println(a...)
	fmt.Print(resetCode)
}

func hexToRGB(hexColor string) string {
	if hexColor[0] == '#' {
		hexColor = hexColor[1:]
	}
	if len(hexColor) == 3 {
		hexColor = string([]byte{hexColor[0], hexColor[0], hexColor[1], hexColor[1], hexColor[2], hexColor[2]})
	}
	if len(hexColor) != 6 {
		panic("Invalid hex color code")
	}
	r := hexColor[0:2]
	g := hexColor[2:4]
	b := hexColor[4:6]
	return fmt.Sprintf("%d;%d;%d", hexToDecimal(r), hexToDecimal(g), hexToDecimal(b))
}

func hexToDecimal(hexValue string) int {
	var result int
	for _, digit := range hexValue {
		result *= 16
		if digit >= '0' && digit <= '9' {
			result += int(digit - '0')
		} else if digit >= 'A' && digit <= 'F' {
			result += int(digit - 'A' + 10)
		} else if digit >= 'a' && digit <= 'f' {
			result += int(digit - 'a' + 10)
		} else {
			panic("Invalid hex digit")
		}
	}
	return result
}
