package main

import (
	"fmt"
	"strconv"
)

/*
expr -> term + expr
      | term - expr
      | term
term -> factor * term
      | factor / term
      | factor
factor -> (expr)
        | num
        | id
id -> [a-zA-Z][a-zA-Z0-9_]*
num -> [0-9][0-9]*

Above syntax can be converted into syntax:
expr -> term (+term | -term)...   // one term follows one or more +term|-term
term -> factor (*factor|/factor)... // one factor follows one or more *factor|/factor

This syntax can be resolved like following for-loop:
func matchExpr(input string) int {
	value, left := matchTerm(input)
	for len(left) > 0 && left[0] == '+' || left[0] == '-' {
		value2, left2 := matchTerm(left[1:])
		value = operator( value, left[0], value2)
	}
	return value
}
*/

const (
	expr   = "expr"
	term   = "term"
	factor = "factor"
	num    = "num"
	id     = "id"
	op     = "op"
)

var (
	initID = 0
)

type AST struct {
	ID          int
	ParentID    int
	Type        string
	Input       string
	ResolvedLen int
	Resolve     string
	Children    []*AST
}

func (a *AST) Print() {
	if a == nil {
		fmt.Println("null")
		return
	}
	a.ParentID = 0
	printStack := []*AST{a}
	var ff func([]*AST)
	f := func(ps []*AST) {
		if len(ps) == 0 {
			return
		}

		end := len(ps)
		var nextStack []*AST
		for j := 0; j < end; j++ {
			ast := ps[j]
			fmt.Printf("<type=%s resolve=%s id=%d parent=%d children=[", ast.Type, ast.Resolve, ast.ID, ast.ParentID)
			for i := 0; i < len(ast.Children); i++ {
				c := ast.Children[i]
				c.ParentID = ast.ID
				if i > 0 {
					fmt.Printf(" ")
				}
				if c.Type == op || c.Type == id || c.Type == num {
					fmt.Printf("%s", c.Resolve)
				} else {
					fmt.Printf("%s", c.Type)
				}
				nextStack = append(nextStack, c)
			}
			fmt.Printf("]>  ")
		}
		fmt.Printf("\n\n")
		ff(nextStack)
	}
	ff = f
	f(printStack)
}

func evalOP(opValue string, v1, v2 int64) int64 {
	switch opValue {
	case "+":
		return v1 + v2
	case "-":
		return v1 - v2
	case "*":
		return v1 * v2
	case "/":
		return v1 / v2
	default:
		panic("invalid op: " + opValue)
	}
}

func (a *AST) Compute() int64 {
	if a.Type == op || a.Type == id {
		panic("cannot compute an op or id")
	}
	var v int64
	var err error
	switch a.Type {
	case num:
		v, err = strconv.ParseInt(a.Resolve, 10, 64)
		if err != nil {
			panic(err)
		}
	case term, expr:
		v = a.Children[0].Compute()
		for i := 1; i < len(a.Children); i += 2 {
			v2 := a.Children[i+1].Compute()
			v = evalOP(a.Children[i].Resolve, v, v2)
		}
	case id:
		panic("cannot compute a id")
	case factor:
		if len(a.Children) == 3 {
			v = a.Children[1].Compute()
		} else {
			v, err = strconv.ParseInt(a.Resolve, 10, 64)
			if err != nil {
				panic(err)
			}
		}
	default:
		panic("invalid ast type for compute: " + a.Type)
	}
	fmt.Printf("resolve %s of %q got %d\n", a.Type, a.Resolve, v)
	return v
}

func NewOP(opStr string) *AST {
	return &AST{
		ID:          genID(),
		Type:        op,
		Input:       opStr,
		Resolve:     opStr,
		ResolvedLen: 1,
	}
}

func matchTerm(input string) *AST {
	c1 := matchFactor(input)
	if c1 == nil {
		return nil
	}

	c := &AST{
		Type:        term,
		ID:          genID(),
		Input:       input,
		ResolvedLen: c1.ResolvedLen,
		Children:    []*AST{c1},
	}
	for c.ResolvedLen < len(input) && (input[c.ResolvedLen] == '*' || input[c.ResolvedLen] == '/') {
		c11 := matchFactor(input[c1.ResolvedLen+1:])
		if c11 == nil {
			return nil
		}
		c.Children = append(c.Children, NewOP(string(input[c.ResolvedLen])))
		c.Children = append(c.Children, c11)
		c.ResolvedLen += c11.ResolvedLen + 1
	}
	c.Resolve = input[:c.ResolvedLen]
	return c
}

func matchFactor(input string) *AST {
	if len(input) == 0 {
		return nil
	}
	if input[0] == '(' {
		c1 := matchExpr(input[1:])
		if c1 == nil || c1.ResolvedLen+1 >= len(input)-1 || input[1+c1.ResolvedLen+1] != ')' {
			return nil
		}
		return &AST{
			ID:          genID(),
			Type:        factor,
			ResolvedLen: 1 + c1.ResolvedLen + 1,
			Resolve:     input[:1+c1.ResolvedLen+1],
			Input:       input,
			Children:    []*AST{NewOP("("), c1, NewOP(")")},
		}
	}
	c1 := matchNum(input)
	if c1 != nil {
		return &AST{
			ID:          genID(),
			Type:        factor,
			ResolvedLen: c1.ResolvedLen,
			Resolve:     input[:c1.ResolvedLen],
			Input:       input,
		}
	}
	c1 = matchID(input)
	if c1 != nil {
		return &AST{
			ID:          genID(),
			Type:        factor,
			ResolvedLen: c1.ResolvedLen,
			Resolve:     input[:c1.ResolvedLen],
			Input:       input,
		}
	}
	return nil
}

func matchNum(input string) *AST {
	i := 0
	for ; i < len(input); i++ {
		c := input[i]
		if c >= '0' && c <= '9' {
			continue
		}
		break
	}
	if i == 0 {
		return nil
	}
	return &AST{
		ID:          genID(),
		Type:        num,
		ResolvedLen: i,
		Resolve:     input[:i],
		Input:       input,
	}
}

func matchID(input string) *AST {
	i := 0
	for ; i < len(input); i++ {
		c := input[i]
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' {
			continue
		}
		if i >= 1 && c >= '0' && c <= '9' {
			continue
		}
		break
	}
	if i == 0 {
		return nil
	}
	return &AST{
		ID:          genID(),
		Input:       input,
		ResolvedLen: i,
		Resolve:     input[:i],
		Type:        id,
	}
}

func matchExpr(input string) *AST {
	if len(input) == 0 {
		return nil
	}
	c1 := matchTerm(input)
	if c1 == nil {
		return nil
	}

	c := &AST{
		Type:        expr,
		Input:       input,
		ID:          genID(),
		ParentID:    0,
		ResolvedLen: c1.ResolvedLen,
		Children:    []*AST{c1},
	}
	for c.ResolvedLen < len(input) && (input[c.ResolvedLen] == '+' || input[c.ResolvedLen] == '-') {
		c11 := matchTerm(input[c.ResolvedLen+1:])
		if c11 == nil {
			fmt.Printf("error: only resolved to %q\n", input[:c.ResolvedLen])
			return nil
		}
		c.Children = append(c.Children, NewOP(string(input[c.ResolvedLen])))
		c.Children = append(c.Children, c11)
		c.ResolvedLen += c11.ResolvedLen + 1
	}
	c.Resolve = input[:c.ResolvedLen]
	if c.ResolvedLen != len(input) {
		fmt.Printf("error: exited for-loop and only resolved to %q\n", input[:c.ResolvedLen])
		return nil
	}

	return c
}

func genID() int {
	initID++
	return initID
}

func main() {
	input := "a1+b1*cd2+ef3"
	got := matchExpr(input)
	fmt.Printf("got is %#v\n", got)
	println("ast tree is:")
	println("")
	got.Print()
	println("")
	println("")
	input = "1+2*3+4-5"
	fmt.Printf("will eval %s:\n", input)
	println("")
	got = matchExpr(input)
	v := got.Compute()
	fmt.Printf("result is: %d\n", input, v)
}
