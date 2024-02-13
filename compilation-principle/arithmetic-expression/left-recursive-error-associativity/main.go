package main

import "fmt"

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
*/

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
	if c1.ResolvedLen < len(input) && (input[c1.ResolvedLen] == '*' || input[c1.ResolvedLen] == '/') {
		c11 := matchTerm(input[c1.ResolvedLen+1:])
		if c11 != nil {
			return &AST{
				Type:        term,
				Input:       input,
				ResolvedLen: c1.ResolvedLen + 1 + c11.ResolvedLen,
				Resolve:     input[:c1.ResolvedLen+1+c11.ResolvedLen],
				Children: []*AST{
					c1, NewOP(string(input[c1.ResolvedLen])), c11,
				},
			}
		}
	} else {
		return &AST{
			ID:          genID(),
			Type:        term,
			Input:       input,
			ResolvedLen: c1.ResolvedLen,
			Resolve:     input[:c1.ResolvedLen],
			Children:    []*AST{c1},
		}
	}
	return nil
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

	if c1.ResolvedLen < len(input) && (input[c1.ResolvedLen] == '+' || input[c1.ResolvedLen] == '-') {
		c11 := matchExpr(input[c1.ResolvedLen+1:])
		if c11 != nil {
			return &AST{
				ID:          genID(),
				Input:       input,
				Type:        expr,
				ResolvedLen: c1.ResolvedLen + 1 + c11.ResolvedLen,
				Resolve:     input[:c1.ResolvedLen+1+c11.ResolvedLen],
				Children:    []*AST{c1, NewOP(string(input[c1.ResolvedLen])), c11},
			}
		}
	} else {
		return &AST{
			ID:          genID(),
			Input:       input,
			Type:        expr,
			ResolvedLen: c1.ResolvedLen,
			Resolve:     input[:c1.ResolvedLen],
			Children:    []*AST{c1},
		}
	}
	return nil
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
	got.Print()
}
