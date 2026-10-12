package main

import "fmt"

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
