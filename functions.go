package main

import (
	"fmt"
	str "strings"
)

var core = map[string]func(*List, *Context) LispValue{
	"%": func(input *List, c *Context) LispValue {
		if len(input.children) != 2 {
			return Atom{t: "error", value: "Wrong number of arguments for function '%'"}
		}
		if input.children[0].Type() != "int" {
			return Atom{t: "error", value: fmt.Sprintf("'%%' function requires arguments of type 'int', not '%s'", input.children[0].Type())}
		}
		if input.children[1].Type() != "int" {
			return Atom{t: "error", value: fmt.Sprintf("'%%' function requires arguments of type 'int', not '%s'", input.children[1].Type())}
		}
		return Atom{t: "int", value: input.children[0].Value().(int64) % input.children[1].Value().(int64)}
	},
	"+": func(input *List, c *Context) LispValue {
		sum := input.children[0].(Atom)
		for _, n := range input.children[1:] {
			if !(n.Type() == "int" || n.Type() == "float") {
				return Atom{t: "error", value: fmt.Sprintf("'+' function requires arguments of type 'int' or 'float', not '%s'", input.children[1].Type())}
			}
			sum = Add(sum, n.(Atom))
		}
		return sum
	},
	"-": func(input *List, c *Context) LispValue {
		if len(input.children) != 2 {
			return Atom{t: "error", value: "Wrong number of arguments for function '-'"}
		}
		return Add(input.children[0].(Atom), Negate(input.children[1].(Atom)))
	},
	"*": func(input *List, c *Context) LispValue {
		sum := input.children[0].(Atom)
		for _, n := range input.children[1:] {
			if !(n.Type() == "int" || n.Type() == "float") {
				return Atom{t: "error", value: fmt.Sprintf("'+' function requires arguments of type 'int' or 'float', not '%s'", input.children[1].Type())}
			}
			sum = Multiply(sum, n.(Atom))
		}
		return sum
	},
	"/": func(input *List, c *Context) LispValue {
		if len(input.children) != 2 {
			return Atom{t: "error", value: "Wrong number of arguments for function '-'"}
		}
		return Divide(input.children[0].(Atom), input.children[1].(Atom))
	},
	"print": func(input *List, c *Context) LispValue {
		if len(input.children) < 1 {
			fmt.Println()
		} else {
			output := []string{}
			for _, n := range input.children {
				r := n.Eval(c)
				if r.Type() == "error" {
					return r
				}
				output = append(output, str.Trim(r.String(), "\""))
			}
			fmt.Println(str.Join(output, " "))
		}
		return NIL
	},
	"cat": func(input *List, c *Context) LispValue {
		if len(input.children) < 1 {
			return Atom{t: "error", value: "Not enough arguments for function 'cat'"}
		}
		text := ""
		for _, s := range input.children {
			text += str.Trim(s.String(), "\"")
		}
		return Atom{t: "string", value: text}
	},
	"head": func(input *List, c *Context) LispValue {
		if len(input.children) < 1 {
			return Atom{t: "error", value: "Not enough arguments for function 'head'"}
		} else if input.children[0].Type() != "list" {
			return Atom{t: "error", value: fmt.Sprintf("Cannot use function 'head' on non-list type argument ('%s')", input.children[0].Type())}
		}
		return input.children[0].(*List).children[0]
	},
	"tail": func(input *List, c *Context) LispValue {
		if len(input.children) < 1 {
			return Atom{t: "error", value: "Not enough arguments for function 'tail'"}
		} else if input.children[0].Type() != "list" {
			return Atom{t: "error", value: fmt.Sprintf("Cannot use function 'tail' on non-list type argument ('%s')", input.children[0].Type())}
		}
		return List{children: input.children[0].(*List).children[1:]}
	},
	"cons": func(input *List, c *Context) LispValue {
		if len(input.children) != 2 {
			return Atom{t: "error", value: "Not enough arguments for function 'cons'"}
		} else if input.children[1].Type() != "list" {
			return Atom{t: "error", value: fmt.Sprintf("Cannot use function 'cons' on non-list type argument ('%s')", input.children[1].Type())}
		}
		ls := input.children[1].(*List)
		ls.children = append(ls.children, input.children[0])
		return ls
	},
	"rev": func(input *List, c *Context) LispValue {
		n := input.children[0]
		if n.Type() != "list" {
			return Atom{t: "error", value: fmt.Sprintf("Cannot use function 'rev' on non-list type argument ('%s')", n.Type())}
		}
		l := n.(*List)
		out := []LispValue{}
		if len(l.children) > 0 {
			for i := len(l.children) - 1; i >= 0; i -= 1 {
				out = append(out, l.children[i])
			}
		}
		return List{children: out}
	},
	"len": func(input *List, c *Context) LispValue {
		if input.children[0].Type() == "list" {
			return input.children[0].(*List).Length()
		}
		if input.children[0].Type() == "string" {
			return input.children[0].(Atom).Length()
		}
		return Atom{t: "error", value: fmt.Sprintf("Cannot use function 'len' on non-list or string type argument ('%s')", input.children[0].Type())}
	},
	"eq": func(input *List, c *Context) LispValue {
		first := input.children[0].Eval(c)
		if first.Type() == "error" {
			return first
		}
		for _, o := range input.children[1:] {
			r := o.Eval(c)
			if r.Type() == "error" {
				return r
			}
			if !Compare(first, r) {
				return FALSE
			}
		}
		return TRUE
	},
	"neq": func(input *List, c *Context) LispValue {
		first := input.children[0].Eval(c)
		if first.Type() == "error" {
			return first
		}
		for _, o := range input.children[1:] {
			r := o.Eval(c)
			if r.Type() == "error" {
				return r
			}
			if Compare(first, r) {
				return FALSE
			}
		}
		return TRUE
	},
	"and": func(input *List, c *Context) LispValue {
		for _, n := range input.children[1:] {
			r := n.Eval(c)
			if r.Type() == "error" {
				return r
			}
			if !Boolean(r) {
				return FALSE
			}
		}
		return TRUE
	},
	"or": func(input *List, c *Context) LispValue {
		for _, n := range input.children[1:] {
			r := n.Eval(c)
			if r.Type() == "error" {
				return r
			}
			if Boolean(r) {
				return TRUE
			}
		}
		return FALSE
	},
	"xor": func(input *List, c *Context) LispValue {
		true_seen := false
		for _, n := range input.children[1:] {
			r := n.Eval(c)
			if r.Type() == "error" {
				return r
			}
			if Boolean(r) {
				if true_seen {
					return FALSE
				} else {
					true_seen = true
				}
			}
		}
		return TRUE
	},
	"not": func(input *List, c *Context) LispValue {
		return Atom{t: "bool", value: !Boolean(input.children[0].Eval(c))}
	},
	"lt": func(input *List, c *Context) LispValue {
		if !(input.children[0].Type() == "int" || input.children[0].Type() == "float") {
			return Atom{t: "error", value: fmt.Sprintf("Value of type '%s' cannot be treated as a number", input.children[0].Type())}
		}
		if !(input.children[1].Type() == "int" || input.children[1].Type() == "float") {
			return Atom{t: "error", value: fmt.Sprintf("Value of type '%s' cannot be treated as a number", input.children[1].Type())}
		}
		return Atom{t: "bool", value: CompareNum(input.children[0].(Atom), input.children[1].(Atom)) == -1}
	},
	"lte": func(input *List, c *Context) LispValue {
		if !(input.children[0].Type() == "int" || input.children[0].Type() == "float") {
			return Atom{t: "error", value: fmt.Sprintf("Value of type '%s' cannot be treated as a number", input.children[0].Type())}
		}
		if !(input.children[1].Type() == "int" || input.children[1].Type() == "float") {
			return Atom{t: "error", value: fmt.Sprintf("Value of type '%s' cannot be treated as a number", input.children[1].Type())}
		}
		return Atom{t: "bool", value: CompareNum(input.children[0].(Atom), input.children[1].(Atom)) <= 0}
	},
	"gt": func(input *List, c *Context) LispValue {
		if !(input.children[0].Type() == "int" || input.children[0].Type() == "float") {
			return Atom{t: "error", value: fmt.Sprintf("Value of type '%s' cannot be treated as a number", input.children[0].Type())}
		}
		if !(input.children[1].Type() == "int" || input.children[1].Type() == "float") {
			return Atom{t: "error", value: fmt.Sprintf("Value of type '%s' cannot be treated as a number", input.children[1].Type())}
		}
		return Atom{t: "bool", value: CompareNum(input.children[0].(Atom), input.children[1].(Atom)) == 1}
	},
	"gte": func(input *List, c *Context) LispValue {
		if !(input.children[0].Type() == "int" || input.children[0].Type() == "float") {
			return Atom{t: "error", value: fmt.Sprintf("Value of type '%s' cannot be treated as a number", input.children[0].Type())}
		}
		if !(input.children[1].Type() == "int" || input.children[1].Type() == "float") {
			return Atom{t: "error", value: fmt.Sprintf("Value of type '%s' cannot be treated as a number", input.children[1].Type())}
		}
		return Atom{t: "bool", value: CompareNum(input.children[0].(Atom), input.children[1].(Atom)) >= 0}
	},
}

var aliases = map[string]string{
	"first": "head",
	"rest":  "tail",
	"add":   "+",
	"mod":   "%",
	"sub":   "-",
	"div":   "/",
	"mul":   "*",
	"equal": "eq",
	"=":     "eq",
	"!":     "not",
	"<":     "lt",
	">":     "gt",
}

func Bootstrap(c *Context) {
	for name, fn := range core {
		c.scope[name] = NewFunction(fn)
	}
	for name, a := range aliases {
		c.scope[name] = c.scope[a]
	}
}
