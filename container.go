package main

import "slices"

// Custom containers
type RheaList struct {
	elements []stackValue
}

func (rl *RheaList) add(e stackValue) {
	rl.elements = append(rl.elements, e)
}

func (rl *RheaList) get(pos int) stackValue {
	return rl.elements[pos]
}

func (rl *RheaList) set(pos int, e stackValue) {
	rl.elements[pos] = e
}

func (rl *RheaList) size() int {
	return len(rl.elements)
}

func (rl *RheaList) pop(pos int) {
	rl.elements = slices.Delete(rl.elements, pos, pos+1)
}

// Custom Containers - end
