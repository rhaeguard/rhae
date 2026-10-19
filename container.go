package main

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"os"
	"slices"
	"strings"
)

// Custom containers
type RhaeList struct {
	elements []stackValue
}

func (rl RhaeList) String() string {
	return fmt.Sprint(rl.elements)
}

func (rl *RhaeList) add(e stackValue) {
	rl.elements = append(rl.elements, e)
}

func (rl *RhaeList) get(pos int) stackValue {
	return rl.elements[pos]
}

func (rl *RhaeList) set(pos int, e stackValue) {
	rl.elements[pos] = e
}

func (rl *RhaeList) size() int {
	return len(rl.elements)
}

func (rl *RhaeList) pop(pos int) {
	rl.elements = slices.Delete(rl.elements, pos, pos+1)
}

type RhaeDict struct {
	dict map[uint32]stackValue
	keys map[uint32]stackValue
}

func NewDict() RhaeDict {
	return RhaeDict{
		dict: map[uint32]stackValue{},
		keys: map[uint32]stackValue{},
	}
}

func (rd *RhaeDict) String() string {
	out := strings.Builder{}

	out.WriteString("{")
	for hk, v := range rd.dict {
		k := rd.keys[hk]
		pair := fmt.Sprintf("%s: %s,", k.String(), v.String())
		out.WriteString(pair)
	}
	out.WriteString("}")
	return out.String()
}

func (rd *RhaeDict) hash(sv *stackValue) uint32 {
	var s string
	switch sv.t {
	case NUMERIC, STRING:
		s = fmt.Sprint(sv.v)
	case BOOL:
		s = "0"
		if sv.v.(bool) {
			s = "1"
		}
	default:
		panic(fmt.Sprintf("unhashable type: t[%s] - v[%v]", sv.t, sv.v))
	}

	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	if err != nil {
		panic(fmt.Sprintf("[hash]: %s", err.Error()))
	}
	return h.Sum32()
}

func (rd *RhaeDict) putKv(k stackValue, v stackValue) {
	hk := rd.hash(&k)
	rd.dict[hk] = v
	rd.keys[hk] = k
}

func (rd *RhaeDict) getValue(k stackValue) stackValue {
	hk := rd.hash(&k)
	if out, ok := rd.dict[hk]; ok {
		return out
	}
	panic(fmt.Sprintf("[dict] key [%s] does not exist", k.String()))
}

func (rd *RhaeDict) deleteKey(k stackValue) {
	hk := rd.hash(&k)
	delete(rd.dict, hk)
	delete(rd.keys, hk)
}

func (rd *RhaeDict) sizeDict() int {
	return len(rd.dict)
}

type RhaeFile struct {
	file    *os.File
	scanner *bufio.Scanner
}

func NewRhaeFile(filepath string) *RhaeFile {
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(fmt.Sprintf("opening a file failed with error: %s", err.Error()))
	}
	scanner := bufio.NewScanner(file) // ScanLines is the default split function
	return &RhaeFile{
		file:    file,
		scanner: scanner,
	}
}

func RhaeFileFrom(file *os.File) *RhaeFile {
	return &RhaeFile{file, bufio.NewScanner(file)}
}

func (rd *RhaeFile) String() string {
	return "|just a file|"
}

func (rf *RhaeFile) readline() string {
	ok := rf.scanner.Scan()
	if ok {
		return rf.scanner.Text()
	}
	return ""
}

func (rf *RhaeFile) write(str string) {
	_, err := rf.file.WriteString(str)
	if err != nil {
		panic(fmt.Sprintf("writing to a file failed with error: %s", err.Error()))
	}
}

func (rf *RhaeFile) close() {
	_ = rf.file.Close()
}

// Custom Containers - end
