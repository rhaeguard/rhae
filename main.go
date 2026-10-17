package main

import (
	"embed"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

//go:embed std/*.rh
var stdDir embed.FS

const (
	numeric instructionType = "numeric"
	str     instructionType = "string"
	list    instructionType = "list"
	word    instructionType = "word"
	load    instructionType = "load"
	store   instructionType = "store"
)

type instructionType = string

type instruction struct {
	value any
	itype instructionType
}

type RhaeInstructions struct {
	ptr        int
	ins        []instruction
	procedures map[string]stackFunc
}

func getStdModules() string {
	var fileContents strings.Builder

	for _, filename := range []string{
		"std/list.rh",
		"std/str.rh",
		"std/utils.rh",
	} {
		dat, err := stdDir.ReadFile(filename)

		if err != nil {
			panic("could not open file")
		}
		fileContents.WriteString("\n\n")
		fileContents.WriteString(string(dat))
	}

	return fileContents.String() + "\n"
}

func NewRhaeInstructions(fileContents string) RhaeInstructions {
	this := RhaeInstructions{ptr: -1, ins: []instruction{}, procedures: initBuiltInProcedures()}
	ix := -1
	for line := range strings.Lines(getStdModules() + fileContents) {
		ix += 1
		b, _, found := strings.Cut(line, "//")
		if found {
			line = b
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		this.lex(line)
	}

	this.preprocess()

	return this
}

// returns index, success, instruction
func (this *RhaeInstructions) Next() (int, bool, instruction) {
	this.ptr += 1
	if this.ptr >= len(this.ins) {
		return -1, false, instruction{}
	}
	ins := this.ins[this.ptr]
	return this.ptr, true, ins
}

func (this *RhaeInstructions) Seek(progress int) {
	this.ptr += progress
}

func (this *RhaeInstructions) ResetPtr() {
	this.ptr = -1
}

func (this *RhaeInstructions) split(s string) []string {
	var quote = []rune(`'"`)
	var delim = []rune(" ")
	const none = '\000'
	open := none
	return strings.FieldsFunc(s, func(r rune) bool {
		switch open {
		case none:
			if strings.ContainsRune(string(quote), r) {
				open = r
				return false
			}
			return strings.ContainsRune(string(delim), r)
		case r:
			open = none
		}
		return false
	})
}

func (this *RhaeInstructions) lex(line string) {
	for _, t := range this.split(line) {
		var tt instruction
		t = strings.TrimSpace(t)
		if vf, err := strconv.ParseFloat(t, 32); err == nil {
			tt = instruction{value: float32(vf), itype: numeric}
		} else if (strings.HasPrefix(t, "\"") && strings.HasSuffix(t, "\"")) || strings.HasPrefix(t, "'") && strings.HasSuffix(t, "'") {
			t = t[1 : len(t)-1]
			tt = instruction{value: t, itype: str}
		} else {
			if len(t) > 1 && strings.HasSuffix(t, ">") {
				t = t[0 : len(t)-1]
				tt = instruction{value: t, itype: load}
			} else if len(t) > 1 && strings.HasSuffix(t, "<") {
				t = t[0 : len(t)-1]
				tt = instruction{value: t, itype: store}
			} else {
				tt = instruction{value: t, itype: word}
			}

		}
		this.ins = append(this.ins, tt)
	}
}

func (this *RhaeInstructions) Eval(stack *stack) {
	var variables = make(map[string]stackValue)
	for {
		_, ok, ins := this.Next()

		if !ok {
			break
		}

		switch ins.itype {
		case numeric:
			stack.Push(ins.value, NUMERIC)
		case str:
			stack.Push(ins.value, STRING)
		case word:
			fn, ok := this.procedures[ins.value.(string)]
			if !ok {
				panic(fmt.Sprintf("word '%s' is missing", ins.value.(string)))
			}
			fn(stack, this)
		case load:
			varName := ins.value.(string)
			if v, ok := variables[varName]; ok {
				stack.Push(v.v, v.t)
			} else {
				panic(fmt.Sprintf("loading unknown variable: [%s]", varName))
			}
		case store:
			varName := ins.value.(string)
			v, _ := stack.Pop()
			variables[varName] = v
		}

		// fmt.Printf("S: %v\n", *stack)
		// fmt.Printf("\tV: %v\n", variables)
	}
}

const (
	NUMERIC stackValueType = "NUMERIC"
	STRING  stackValueType = "STRING"
	BOOL    stackValueType = "BOOL"
	LIST    stackValueType = "LIST"
	DICT    stackValueType = "DICT"
	REF     stackValueType = "REF"
)

type stackValueType = string

type stackValue struct {
	v any
	t stackValueType
}

func (sv stackValue) String() string {
	switch sv.t {
	case NUMERIC:
		return fmt.Sprint(sv.v)
	case STRING:
		return fmt.Sprint(sv.v)
	case BOOL:
		return fmt.Sprintf("%t", sv.v)
	case LIST:
		return fmt.Sprint(*(sv.v.(*RhaeList)))
	}
	return fmt.Sprintf("%v", sv.v)
}

func (s stackValue) assertType(prefix string, assertedTypes ...stackValueType) {
	if slices.Contains(assertedTypes, s.t) {
		return
	}
	panic(fmt.Sprintf("[%s] expected %-v, got %s", prefix, assertedTypes, s.t))
}

type stack struct {
	pointer uint
	stack   []stackValue
}

func NewStack() stack {
	return stack{}
}

func (s *stack) Push(v any, t stackValueType) {
	switch t {
	case NUMERIC, STRING, BOOL, REF, LIST:
	default:
		panic(fmt.Sprintf("unknown stack entry type: [%s], value: [%v]", t, v))
	}
	s.stack = append(s.stack, stackValue{v, t})
}

func (s *stack) Pop() (stackValue, bool) {
	ix := len(s.stack) - 1
	if ix == -1 {
		return stackValue{}, false
	}

	sv := s.stack[ix]
	out := stackValue{sv.v, sv.t}

	s.stack = s.stack[:ix]

	return out, true
}

type offsetT = int
type offsetL = struct {
	label string
	start int
	end   int
}

func (this *RhaeInstructions) preprocess() {
	controlStack := NewStack()

	finalIns := []instruction{}

	insPos := func() int { return len(finalIns) - 1 }

	for {
		_, ok, ins := this.Next()

		if !ok {
			break
		}

		if ins.itype == word {
			fn := ins.value.(string)

			switch fn {
			case "if":
				var offset offsetT
				finalIns = append(finalIns,
					instruction{
						itype: "offset",
						value: &offset,
					},
					instruction{
						itype: word,
						value: "jump0",
					},
				)
				offset = insPos()
				controlStack.Push(&offset, REF)
			case "else":
				var elseOffset offsetT
				finalIns = append(finalIns,
					instruction{
						itype: "offset",
						value: &elseOffset,
					},
					instruction{
						itype: word,
						value: "jump",
					},
				)
				myIx := insPos()
				svalue, ok := controlStack.Pop()
				if !ok {
					panic("cannot pop this stack")
				}
				// update if offset value
				gap := myIx - *svalue.v.(*offsetT)
				finalIns = slices.Replace(finalIns, myIx-gap-1, myIx-gap, instruction{
					itype: numeric, value: gap,
				})

				// push else offset value
				elseOffset = myIx
				controlStack.Push(&elseOffset, REF)
			case "loop":
				finalIns = append(finalIns,
					instruction{itype: word, value: "loop"},
				)
				offset := offsetL{"loop", insPos(), -1}
				controlStack.Push(&offset, REF)
			case "func":
				finalIns = append(finalIns,
					instruction{itype: word, value: "func"},
				)
				offset := offsetL{"func", insPos(), -1}
				controlStack.Push(&offset, REF)
			case "do":
				svalue, ok := controlStack.Pop()
				if !ok {
					panic("cannot pop this stack")
				}
				labelOffset := svalue.v.(*offsetL)

				if labelOffset.label == "loop" {
					finalIns = append(finalIns,
						instruction{
							itype: "offset",
							value: labelOffset,
						},
						instruction{
							itype: word,
							value: "jump0",
						},
					)
					labelOffset.end = insPos()
					controlStack.Push(labelOffset, REF)
				} else {
					this.Seek(-2)
					_, _, funcName := this.Next()
					this.Next() // skip me
					fName := funcName.value.(string)
					labelOffset.label = fName
					labelOffset.start = insPos() + 1
					controlStack.Push(labelOffset, REF)
				}
			case "end":
				svalue, ok := controlStack.Pop()
				if !ok {
					panic("cannot pop this stack")
				}
				if vv, isSimpleOffset := svalue.v.(*offsetT); isSimpleOffset {
					// handles if/else/end
					finalIns = append(finalIns, instruction{
						itype: word,
						value: "end",
					})
					myIx := insPos()
					gap := myIx - int(*vv)

					finalIns = slices.Replace(finalIns, myIx-gap-1, myIx-gap, instruction{
						itype: numeric, value: gap,
					})
				} else if vv, isLabelOffset := svalue.v.(*offsetL); isLabelOffset {
					// handles loops
					lStart := vv.start
					lEnd := vv.end
					label := vv.label

					if label == "loop" {
						finalIns = append(finalIns,
							instruction{
								itype: numeric,
								value: lStart - (insPos() + 2), // this instructions position
							},
							instruction{
								itype: word,
								value: "jump",
							},
							instruction{
								itype: word,
								value: "end",
							},
						)

						myIx := insPos()
						gap := myIx - lEnd

						finalIns = slices.Replace(finalIns, myIx-gap-1, myIx-gap, instruction{
							itype: numeric, value: gap,
						})
					} else {
						lastInsPos := insPos()
						finalIns = slices.Replace(finalIns, lStart-2, lStart,
							instruction{
								itype: numeric, value: (lastInsPos - lStart) + 2, // +2 for offset and jump
							},
							instruction{
								itype: word, value: "jump",
							},
						)
						finalIns = append(finalIns,
							instruction{
								itype: word,
								value: "end",
							},
						)

						allFuncInstructions := finalIns[lStart : lastInsPos+1]

						this.procedures[label] = func(s *stack, r *RhaeInstructions) {
							t := RhaeInstructions{
								ptr:        -1,
								ins:        allFuncInstructions,
								procedures: r.procedures,
							}
							t.Eval(s)
						}
					}
				}

			default:
				finalIns = append(finalIns, ins)
			}
		} else {
			finalIns = append(finalIns, ins)
		}
	}

	this.ResetPtr()
	this.ins = finalIns
}

func main() {
	var filename string
	if len(os.Args) > 1 {
		filename = os.Args[1]
	} else {
		fmt.Println("please provide a file to run")
	}
	dat, err := os.ReadFile(filename)

	if err != nil {
		panic("could not open file")
	}

	fileContents := string(dat)

	rhInstructions := NewRhaeInstructions(fileContents)

	for false {
		ix, ok, ins := rhInstructions.Next()

		if !ok {
			break
		}

		fmt.Printf("%3d - nor: %-v\n", ix, ins)
	}

	fmt.Printf("---------------------|\n")

	rhInstructions.ResetPtr()
	stack := NewStack()
	rhInstructions.Eval(&stack)
}
