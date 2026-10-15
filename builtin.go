package main

import (
	"fmt"
	"maps"
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
			panic(fmt.Sprintf("invalid comparison of 2 different types [%s, %s]", a.t, b.t))
		}
	}
}

func strBuiltins() (string, map[string]stackFunc) {
	procedures := map[string]stackFunc{
		"chars": func(s *stack, ri *RhaeInstructions) {
			itemV, _ := s.Pop()

			itemV.assertType("chars", STRING)
			listV := RheaList{}

			for _, ch := range itemV.v.(string) {
				listV.add(stackValue{int(ch), NUMERIC})
			}

			s.Push(&listV, LIST)
		},
		"str": func(s *stack, ri *RhaeInstructions) {
			input, _ := s.Pop()

			input.assertType("str", LIST, NUMERIC)

			listO, isList := input.v.(*RheaList)

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
			listV := RheaList{}
			s.Push(&listV, LIST)
		},
		"add": func(s *stack, ri *RhaeInstructions) {
			itemV, _ := s.Pop()
			listV, _ := s.Pop()

			listV.assertType("add", LIST)

			listO := listV.v.(*RheaList)
			listO.add(itemV)
		},
		"get": func(s *stack, ri *RhaeInstructions) {
			posV, _ := s.Pop()
			listV, _ := s.Pop()

			listV.assertType("get", LIST)
			posV.assertType("get", NUMERIC)

			listO := listV.v.(*RheaList)
			out := listO.get(int(posV.v.(float32)))
			s.Push(out.v, out.t)
		},
		"set": func(s *stack, ri *RhaeInstructions) {
			itemV, _ := s.Pop()
			posV, _ := s.Pop()
			listV, _ := s.Pop()

			listV.assertType("set", LIST)
			posV.assertType("set", NUMERIC)

			listO := listV.v.(*RheaList)
			pos := int(posV.v.(float32))
			listO.set(pos, itemV)
		},
		"size": func(s *stack, ri *RhaeInstructions) {
			listV, _ := s.Pop()

			listV.assertType("size", LIST)

			listO := listV.v.(*RheaList)
			s.Push(listO.size(), NUMERIC)
		},
		"pop": func(s *stack, ri *RhaeInstructions) {
			posV, _ := s.Pop()
			listV, _ := s.Pop()

			listV.assertType("pop", LIST)
			posV.assertType("pop", NUMERIC)

			listO := listV.v.(*RheaList)
			listO.pop(asInt(posV.v))
		},
	}

	return "list", procedures
}

func initBuiltInProcedures() map[string]stackFunc {
	procedures := map[string]stackFunc{
		"print": func(s *stack, ri *RhaeInstructions) {
			p, _ := s.Pop()
			if p.t == STRING {
				fmt.Printf("%s", p.v)
			} else {
				fmt.Printf("%v", p.v)
			}
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
		"end":  func(s *stack, ri *RhaeInstructions) {},
		"loop": func(s *stack, ri *RhaeInstructions) {},
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

	strNamespace, strProcedures := strBuiltins()
	_ = strNamespace // ignoring namespacing for now
	maps.Copy(procedures, strProcedures)

	return procedures
}
