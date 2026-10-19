package main

import (
	"fmt"
	"maps"
	"math"
	"os"
	"slices"
	"strings"
)

type stackFunc = func(*stack, *RhaeInstructions)

func arithmeticF(fn func(a, b float32) float32) func(*stack, *RhaeInstructions) {
	return func(s *stack, ri *RhaeInstructions) {
		b, _ := s.Pop()
		a, _ := s.Pop()

		left := asFloat(a.v)
		right := asFloat(b.v)

		var res float32 = fn(left, right)

		s.Push(res, NUMERIC)
	}
}

func arithmeticI(fn func(a, b int) int) func(*stack, *RhaeInstructions) {
	return func(s *stack, ri *RhaeInstructions) {
		b, _ := s.Pop()
		a, _ := s.Pop()

		left := asInt(a.v)
		right := asInt(b.v)

		var res float32 = float32(fn(left, right))

		s.Push(res, NUMERIC)
	}
}

func compareNumeric(a, b float32) int {
	if a > b {
		return 1
	} else if a < b {
		return -1
	} else {
		return 0
	}
}

func comparisonF(
	generalCmp func(a int) bool,
	boolCmp func(a, b bool) bool,
) func(*stack, *RhaeInstructions) {
	return func(s *stack, ri *RhaeInstructions) {
		b, _ := s.Pop()
		a, _ := s.Pop()

		if a.t == NUMERIC && b.t == NUMERIC {
			left := asFloat(a.v)
			right := asFloat(b.v)

			rcmp := compareNumeric(left, right)

			var res bool = generalCmp(rcmp)

			s.Push(res, BOOL)
		} else if a.t == STRING && b.t == STRING {
			left := a.v.(string)
			right := b.v.(string)

			rcmp := strings.Compare(left, right)

			var res bool = generalCmp(rcmp)

			s.Push(res, BOOL)
		} else if a.t == BOOL && b.t == BOOL {
			left := a.v.(bool)
			right := b.v.(bool)

			var res bool = boolCmp(left, right)

			s.Push(res, BOOL)
		} else {
			panic(
				fmt.Sprintf("invalid comparison of 2 different types [%s, %s] - (%s, %s)",
					a.t,
					b.t,
					a.String(),
					b.String(),
				),
			)
		}
	}
}

func strBuiltins() (string, map[string]stackFunc) {
	procedures := map[string]stackFunc{
		"chars": func(s *stack, ri *RhaeInstructions) {
			itemV, _ := s.Pop()

			itemV.assertType("chars", STRING)
			listV := RhaeList{}

			for _, ch := range itemV.v.(string) {
				listV.add(stackValue{int(ch), NUMERIC})
			}

			s.Push(&listV, LIST)
		},
		"str": func(s *stack, ri *RhaeInstructions) {
			input, _ := s.Pop()

			input.assertType("str", LIST, NUMERIC)

			listO, isList := input.v.(*RhaeList)

			var out string
			if isList {
				temp := []rune{}

				for i := range listO.size() {
					ch := asInt(listO.get(i).v)
					temp = append(temp, rune(ch))
				}

				out = string(temp)
			} else {
				charCode := rune(asInt(input.v))
				out = string(charCode)
			}
			s.Push(out, STRING)
		},
	}

	return "str", procedures
}

func listBuiltins() (string, map[string]stackFunc) {
	procedures := map[string]stackFunc{
		"list": func(s *stack, ri *RhaeInstructions) {
			listV := RhaeList{}
			s.Push(&listV, LIST)
		},
		"add": func(s *stack, ri *RhaeInstructions) {
			itemV, _ := s.Pop()
			listV, _ := s.Pop()

			listV.assertType("add", LIST)

			listO := listV.v.(*RhaeList)
			listO.add(itemV)
		},
		"get": func(s *stack, ri *RhaeInstructions) {
			posV, _ := s.Pop()
			listV, _ := s.Pop()

			listV.assertType("get", LIST)
			posV.assertType("get", NUMERIC)

			listO := listV.v.(*RhaeList)
			out := listO.get(int(posV.v.(float32)))
			s.Push(out.v, out.t)
		},
		"set": func(s *stack, ri *RhaeInstructions) {
			itemV, _ := s.Pop()
			posV, _ := s.Pop()
			listV, _ := s.Pop()

			listV.assertType("set", LIST)
			posV.assertType("set", NUMERIC)

			listO := listV.v.(*RhaeList)
			pos := int(posV.v.(float32))
			listO.set(pos, itemV)
		},
		"size": func(s *stack, ri *RhaeInstructions) {
			listV, _ := s.Pop()

			listV.assertType("size", LIST)

			listO := listV.v.(*RhaeList)
			s.Push(listO.size(), NUMERIC)
		},
		"pop": func(s *stack, ri *RhaeInstructions) {
			posV, _ := s.Pop()
			listV, _ := s.Pop()

			listV.assertType("pop", LIST)
			posV.assertType("pop", NUMERIC)

			listO := listV.v.(*RhaeList)
			listO.pop(asInt(posV.v))
		},
	}

	return "list", procedures
}

func dictBuiltins() (string, map[string]stackFunc) {
	procedures := map[string]stackFunc{
		"dict": func(s *stack, ri *RhaeInstructions) {
			dictV := NewDict()
			s.Push(&dictV, DICT)
		},
		"putKv": func(s *stack, ri *RhaeInstructions) {
			// dict k v putKv
			valueV, _ := s.Pop()
			keyV, _ := s.Pop()
			dictV, _ := s.Pop()

			dictV.assertType("putKv", DICT)

			dictO := dictV.v.(*RhaeDict)
			dictO.putKv(keyV, valueV)
		},
		"getValue": func(s *stack, ri *RhaeInstructions) {
			// dict k getValue
			keyV, _ := s.Pop()
			dictV, _ := s.Pop()

			dictV.assertType("getValue", DICT)

			dictO := dictV.v.(*RhaeDict)
			out := dictO.getValue(keyV)
			s.Push(out.v, out.t)
		},
		"deleteKey": func(s *stack, ri *RhaeInstructions) {
			// dict k deleteKey
			keyV, _ := s.Pop()
			dictV, _ := s.Pop()

			dictV.assertType("deleteKey", DICT)

			dictO := dictV.v.(*RhaeDict)
			dictO.deleteKey(keyV)
		},
		"sizeDict": func(s *stack, ri *RhaeInstructions) {
			dictV, _ := s.Pop()

			dictV.assertType("sizeDict", DICT)

			dictO := dictV.v.(*RhaeDict)
			s.Push(dictO.sizeDict(), NUMERIC)
		},
	}

	return "dict", procedures
}

func ioBuiltins() (string, map[string]stackFunc) {
	procedures := map[string]stackFunc{
		"stdout": func(s *stack, ri *RhaeInstructions) {
			file := RhaeFileFrom(os.Stdout)
			s.Push(file, FILE)
		},
		"stderr": func(s *stack, ri *RhaeInstructions) {
			file := RhaeFileFrom(os.Stderr)
			s.Push(file, FILE)
		},
		"stdin": func(s *stack, ri *RhaeInstructions) {
			file := RhaeFileFrom(os.Stdin)
			s.Push(file, FILE)
		},
		"open": func(s *stack, ri *RhaeInstructions) {
			stringV, _ := s.Pop()

			stringV.assertType("open", STRING)

			file := NewRhaeFile(stringV.v.(string))
			s.Push(file, FILE)
		},
		"write": func(s *stack, ri *RhaeInstructions) {
			stringV, _ := s.Pop()
			fileV, _ := s.Pop()

			stringV.assertType("write", STRING)
			fileV.assertType("write", FILE)

			fileO := fileV.v.(*RhaeFile)
			fileO.write(stringV.v.(string))
		},
		"readline": func(s *stack, ri *RhaeInstructions) {
			fileV, _ := s.Pop()

			fileV.assertType("readline", FILE)

			fileO := fileV.v.(*RhaeFile)
			outS := fileO.readline()
			s.Push(outS, STRING)
		},
		"close": func(s *stack, ri *RhaeInstructions) {
			fileV, _ := s.Pop()

			fileV.assertType("close", FILE)

			fileO := fileV.v.(*RhaeFile)
			fileO.close()
		},
	}

	return "io", procedures
}

func initBuiltInProcedures() map[string]stackFunc {
	procedures := map[string]stackFunc{
		"print": func(s *stack, ri *RhaeInstructions) {
			p, _ := s.Pop()
			fmt.Printf("%s", p.String())
		},
		"jump0": func(s *stack, ri *RhaeInstructions) {
			offsetV, _ := s.Pop()
			resultV, _ := s.Pop()

			offset := asInt(offsetV.v)
			conditionResult := resultV.v.(bool)

			if conditionResult == false {
				ri.Seek(offset)
			}
		},
		"jump": func(s *stack, ri *RhaeInstructions) {
			offsetV, _ := s.Pop()
			offset := asInt(offsetV.v)
			ri.Seek(offset)
		},
		"floor": func(s *stack, ri *RhaeInstructions) {
			itemV, _ := s.Pop()
			value := asFloat(itemV.v)
			value = float32(math.Floor(float64(value)))
			s.Push(value, NUMERIC)
		},
		"end":  func(s *stack, ri *RhaeInstructions) {},
		"loop": func(s *stack, ri *RhaeInstructions) {},
		":clear": func(s *stack, ri *RhaeInstructions) {
			for true {
				_, ok := s.Pop()
				if !ok {
					break
				}
			}
		},
		":func": func(s *stack, ri *RhaeInstructions) {
			keys := slices.Collect(maps.Keys(ri.procedures))

			slices.Sort(keys)

			for _, k := range keys {
				fmt.Printf("[debug] %s\n", k)
			}
		},
	}

	// arithmetics
	procedures["+"] = arithmeticF(func(a, b float32) float32 { return a + b })
	procedures["-"] = arithmeticF(func(a, b float32) float32 { return a - b })
	procedures["*"] = arithmeticF(func(a, b float32) float32 { return a * b })
	procedures["/"] = arithmeticF(func(a, b float32) float32 { return a / b })
	procedures["%"] = arithmeticI(func(a, b int) int { return a % b })
	// bitwise
	procedures["&"] = arithmeticI(func(a, b int) int { return a & b })
	procedures["^"] = arithmeticI(func(a, b int) int { return a ^ b })
	procedures["|"] = arithmeticI(func(a, b int) int { return a | b })
	procedures["<<"] = arithmeticI(func(a, b int) int { return a << b })
	procedures[">>"] = arithmeticI(func(a, b int) int { return a >> b })

	// comparison
	panickerB := func(op string) func(bool, bool) bool {
		return func(bool, bool) bool { panic(fmt.Sprintf("unsupported operation: '%s'", op)) }
	}

	panicker := func(op string) func(int) bool {
		return func(int) bool { panic(fmt.Sprintf("unsupported operation: '%s'", op)) }
	}

	procedures["<"] = comparisonF(func(a int) bool { return a < 0 }, panickerB("<"))
	procedures["<="] = comparisonF(func(a int) bool { return a <= 0 }, panickerB("<="))
	procedures[">"] = comparisonF(func(a int) bool { return a > 0 }, panickerB(">"))
	procedures[">="] = comparisonF(func(a int) bool { return a >= 0 }, panickerB(">="))
	procedures["=="] = comparisonF(func(a int) bool { return a == 0 }, func(a, b bool) bool { return a == b })
	procedures["!="] = comparisonF(func(a int) bool { return a != 0 }, func(a, b bool) bool { return a != b })
	procedures["and"] = comparisonF(panicker("and"), func(a, b bool) bool { return a && b })
	procedures["&&"] = comparisonF(panicker("&&"), func(a, b bool) bool { return a && b })
	procedures["or"] = comparisonF(panicker("or"), func(a, b bool) bool { return a || b })
	procedures["||"] = comparisonF(panicker("||"), func(a, b bool) bool { return a || b })

	// list functions
	listNamespace, listProcedures := listBuiltins()
	_ = listNamespace // ignoring namespacing for now
	maps.Copy(procedures, listProcedures)

	// dict functions
	dictNamespace, dictProcedures := dictBuiltins()
	_ = dictNamespace // ignoring namespacing for now
	maps.Copy(procedures, dictProcedures)

	strNamespace, strProcedures := strBuiltins()
	_ = strNamespace // ignoring namespacing for now
	maps.Copy(procedures, strProcedures)

	ioNamespace, ioProcedures := ioBuiltins()
	_ = ioNamespace // ignoring namespacing for now
	maps.Copy(procedures, ioProcedures)

	return procedures
}
