package main

import (
	"fmt"
	"image/color"
	"os"
	"strconv"
)

func asFloat(v any) float32 {
	r, ok := v.(float32)
	if ok {
		return r
	}

	x, ok := v.(int)
	if ok {
		return float32(x)
	}
	panic(fmt.Sprintf("expected float32 got %-v", v))
}

func asInt(v any) int {
	r, ok := v.(int)
	if ok {
		return r
	}

	x, ok := v.(float32)
	if ok {
		return int(x)
	}
	panic(fmt.Sprintf("expected int got %-v", v))
}

func hexToRgba(hex string) color.RGBA {
	hex = hex[1:]
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	a, _ := strconv.ParseUint(hex[6:8], 16, 8)

	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
}

func GetFileContents(filename string) string {
	dat, err := os.ReadFile(filename)

	if err != nil {
		panic(fmt.Sprintf("could not the open file: %s", filename))
	}

	return string(dat)
}
