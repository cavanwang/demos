package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}
)

func middleToSuffix(s string) ([]string, error) {
	var result []string
	var stack []string
	var i int
	for ; i < len(s); i++ {
		current := string(s[i])
		if asciiSpace[current[0]] == 1 {
			continue
		}

		if isNum(current) {
			l := getNumStrLen(s, i)
			result = append(result, s[i:l+i])
			i = l + i - 1
			continue
		}

		if current == "(" {
			stack = append(stack, current)
			continue
		}

		if current == ")" {
			found := false
			for len(stack) > 0 {
				index := len(stack) - 1
				if stack[index] != "(" {
					result = append(result, string(stack[index]))
					stack = stack[:index]
					continue
				}
				found = true
				stack = stack[:index]
				break
			}
			if !found {
				return nil, fmt.Errorf("i=%d stack left=%+v not found '(", i, stack)
			}
			continue
		}

		op := false
		for isOp(current) {
			op = true
			if len(stack) == 0 || stack[len(stack)-1] == "(" || opPriorityLargerThan(current, stack[len(stack)-1]) {
				stack = append(stack, current)
				break
			}
			result = append(result, stack[len(stack)-1])
			stack = stack[:len(stack)-1]
		}
		if !op {
			return nil, fmt.Errorf("invalid symbol %q", current)
		}
	}

	for len(stack) > 0 {
		result = append(result, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}
	return result, nil
}

func computeSuffix(s []string) (float64, error) {
	var result []float64
	for i := 0; i < len(s); i++ {
		if !isOp(s[i]) {
			sint, err := strconv.ParseFloat(s[i], 63)
			if err != nil {
				return 0, err
			}
			result = append(result, sint)
		} else {
			if len(result) < 2 {
				return 0, fmt.Errorf("left number count less than 2 for %q: %+v", s[i], result)
			}
			if s[i] == "+" {
				result[len(result)-2] = result[len(result)-2] + result[len(result)-1]
			} else if s[i] == "-" {
				result[len(result)-2] = result[len(result)-2] - result[len(result)-1]
			} else if s[i] == "*" {
				result[len(result)-2] = result[len(result)-2] * result[len(result)-1]
			} else if s[i] == "/" {
				result[len(result)-2] = result[len(result)-2] / result[len(result)-1]
			} else {
				return 0, fmt.Errorf("invalid symbol %q", s[i])
			}
			result = result[:len(result)-1]
		}
	}
	if len(result) != 1 {
		return 0, fmt.Errorf("finnaly expect result length is 1, but got: %+v", result)
	}
	return result[len(result)-1], nil
}

func getNumStrLen(s string, index int) (length int) {
	expect := "n"
	for i := index; i < len(s); i++ {
		if expect == "n" {
			if !isNum(string(s[i])) {
				return length
			}
			length++
			expect = ".n"
			continue
		}
		//expect == ".n"
		if isNum(string(s[i])) {
			length++
		} else if s[i] == '.' {
			length++
			expect = "n"
		} else {
			return length
		}
	}
	return length
}

func isNum(c string) (r bool) {
	for i := 0; i < len(c); i++ {
		if c[i] < '0' || c[i] > '9' {
			return false
		}
	}
	return len(c) != 0
}

func isOp(c string) bool {
	return c == "+" || c == "-" || c == "*" || c == "/"
}

func opPriorityLargerThan(op1, op2 string) bool {
	if op1 == "+" || op1 == "-" {
		return false
	}
	return op2 == "+" || op2 == "-"

}

func nextWhiteSpaceCount(s string, startIndex int) int {
	r := 0
	for i := startIndex; i < len(s); i++ {
		r += int(asciiSpace[s[i]])
	}
	return r
}

func main() {
	s := strings.Join(os.Args[1:], " ")
	got, err := middleToSuffix(s)
	if err != nil {
		fmt.Println("middleToSuffix error:", err.Error())
		os.Exit(1)
	}
	fmt.Printf("%q resolve to %v\n", s, got)

	result, err := computeSuffix(got)
	if err != nil {
		fmt.Println("computeSuffix error:", err.Error())
		os.Exit(1)
	}
	fmt.Printf("%v compute got: %f\n", got, result)
}
