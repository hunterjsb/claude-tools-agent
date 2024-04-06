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

	// Pastel
	"pastel_red":     "#FF6666",
	"pastel_orange":  "#FFB266",
	"pastel_yellow":  "#FFFF66",
	"pastel_green":   "#B2FF66",
	"pastel_blue":    "#66B2FF",
	"pastel_purple":  "#B266FF",
	"pastel_pink":    "#FF66B2",
	"pastel_cyan":    "#66FFFF",
	"pastel_gray":    "#B2B2B2",
	"pastel_lime":    "#B2FF66",
	"pastel_teal":    "#66FFB2",
	"pastel_indigo":  "#8066FF",
	"pastel_brown":   "#CC9966",
	"pastel_magenta": "#FF66FF",
	"pastel_white":   "#F0F0F0",
	"pastel_black":   "#333333",

	// Vintage
	"vintage_red":     "#D2691E",
	"vintage_blue":    "#6495ED",
	"vintage_green":   "#556B2F",
	"vintage_pink":    "#DB7093",
	"vintage_purple":  "#9932CC",
	"vintage_orange":  "#FF7F50",
	"vintage_yellow":  "#F0E68C",
	"vintage_brown":   "#8B4513",
	"vintage_gray":    "#696969",
	"vintage_teal":    "#5F9EA0",
	"vintage_cyan":    "#00CED1",
	"vintage_lime":    "#32CD32",
	"vintage_indigo":  "#4169E1",
	"vintage_magenta": "#FF1493",
	"vintage_white":   "#FFFAF0",
	"vintage_black":   "#191970",
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

func PrintAllColors() {
	for colorName, hexCode := range colorMap {
		CPrintHex(hexCode, colorName)
	}
}
