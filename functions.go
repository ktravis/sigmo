package main

import (
	"fmt"
	str "strings"
)

var core = map[string]LispFunction{
	"%": NewFunction("%", "int,int", func(input *List, c *Context) LispValue {
		return Atom{t: "int", value: input.children[0].Value().(int64) % input.children[1].Value().(int64)}
	}),
	"+": NewFunction("+", "int|float,+", func(input *List, c *Context) LispValue {
		sum := input.children[0].(Atom)
		for _, n := range input.children[1:] {
			sum = Add(sum, n.(Atom))
		}
		return sum
	}),
	"-": NewFunction("-", "int|float,int|float", func(input *List, c *Context) LispValue {
		return Add(input.children[0].(Atom), Negate(input.children[1].(Atom)))
	}),
	"*": NewFunction("*", "int|float,+", func(input *List, c *Context) LispValue {
		sum := input.children[0].(Atom)
		for _, n := range input.children[1:] {
			sum = Multiply(sum, n.(Atom))
		}
		return sum
	}),
	"/": NewFunction("/", "int|float,int|float", func(input *List, c *Context) LispValue {
		return Divide(input.children[0].(Atom), input.children[1].(Atom))
	}),
	"print": NewFunction("print", "**", func(input *List, c *Context) LispValue {
		output := []string{}
		for _, n := range input.children {
			if n.Type() == "string" {
				output = append(output, n.Value().(string))
			} else {
				output = append(output, n.String())
			}
		}
		fmt.Printf(str.Join(output, " "))
		return NIL
	}),
	"println": NewFunction("println", "**", func(input *List, c *Context) LispValue {
		if len(input.children) < 1 {
			fmt.Println()
		} else {
			output := []string{}
			for _, n := range input.children {
				if n.Type() == "string" {
					output = append(output, n.Value().(string))
				} else {
					output = append(output, n.String())
				}
			}
			fmt.Println(str.Join(output, " "))
		}
		return NIL
	}),
	"cat": NewFunction("cat", "string,+", func(input *List, c *Context) LispValue {
		text := ""
		for _, s := range input.children {
			text += str.Trim(s.String(), "\"")
		}
		return Atom{t: "string", value: text}
	}),
	"head": NewFunction("head", "list", func(input *List, c *Context) LispValue {
		return input.children[0].(*List).children[0]
	}),
	"tail": NewFunction("tail", "list", func(input *List, c *Context) LispValue {
		return &List{children: input.children[0].(*List).children[1:]}
	}),
	"cons": NewFunction("cons", "*,*", func(input *List, c *Context) LispValue {
		var left, right *List
		if input.children[0].Type() != "list" {
			left = &List{children: []LispValue{input.children[0]}}
		} else {
			left = input.children[0].(*List)
		}
		if input.children[1].Type() != "list" {
			right = &List{children: []LispValue{input.children[1]}}
		} else {
			right = input.children[1].(*List)
		}
		left.children = append(left.children, right.children...)
		return left
	}),
	"rev": NewFunction("rev", "list|string", func(input *List, c *Context) LispValue {
		l := input.children[0].(*List)
		out := []LispValue{}
		if len(l.children) > 0 {
			for i := len(l.children) - 1; i >= 0; i -= 1 {
				out = append(out, l.children[i])
			}
		}
		return &List{children: out}
	}),
	"len": NewFunction("len", "list|string", func(input *List, c *Context) LispValue {
		if input.children[0].Type() == "list" {
			return input.children[0].(*List).Length()
		} else {
			return input.children[0].(Atom).Length()
		}
	}),
	"eq": NewFunction("eq", "*,*", func(input *List, c *Context) LispValue {
		if !Compare(input.children[0], input.children[1]) {
			return FALSE
		}
		return TRUE
	}),
	"neq": NewFunction("neq", "*,*", func(input *List, c *Context) LispValue {
		if Compare(input.children[0], input.children[1]) {
			return FALSE
		}
		return TRUE
	}),
	"and": NewFunction("and", "bool,+", func(input *List, c *Context) LispValue {
		for _, n := range input.children {
			if !Boolean(n) {
				return FALSE
			}
		}
		return TRUE
	}),
	"or": NewFunction("or", "bool,+", func(input *List, c *Context) LispValue {
		for _, n := range input.children {
			if Boolean(n) {
				return TRUE
			}
		}
		return FALSE
	}),
	"xor": NewFunction("xor", "bool,bool", func(input *List, c *Context) LispValue {
		true_seen := false
		for _, n := range input.children {
			if Boolean(n) {
				if true_seen {
					return FALSE
				} else {
					true_seen = true
				}
			}
		}
		return TRUE
	}),
	"not": NewFunction("not", "bool", func(input *List, c *Context) LispValue {
		return Atom{t: "bool", value: !Boolean(input.children[0])}
	}),
	"lt": NewFunction("lt", "int|float,int|float", func(input *List, c *Context) LispValue {
		return Atom{t: "bool", value: CompareNum(input.children[0].(Atom), input.children[1].(Atom)) == -1}
	}),
	"lte": NewFunction("lte", "int|float,int|float", func(input *List, c *Context) LispValue {
		return Atom{t: "bool", value: CompareNum(input.children[0].(Atom), input.children[1].(Atom)) <= 0}
	}),
	"gt": NewFunction("gt", "int|float,int|float", func(input *List, c *Context) LispValue {
		return Atom{t: "bool", value: CompareNum(input.children[0].(Atom), input.children[1].(Atom)) == 1}
	}),
	"gte": NewFunction("gte", "int|float,int|float", func(input *List, c *Context) LispValue {
		return Atom{t: "bool", value: CompareNum(input.children[0].(Atom), input.children[1].(Atom)) >= 0}
	}),
	"exec": NewFunction("exec", "list", func(input *List, c *Context) LispValue {
		x := input.children[0].Copy().(*List)
		x.Quoted = false
		return x.Eval(c)
	}),
	"eval": NewFunction("eval", "string", func(input *List, c *Context) LispValue {
		var last LispValue = NIL
		x := input.children[0].Value().(string)
		for _, n := range Parse(Tokenize(x)) {
			last = n.Eval(c)
		}
		return last
	}),
	"trim": NewFunction("trim", "string,string", func(input *List, c *Context) LispValue {
		return Atom{t: "string", value: str.Trim(input.children[0].Value().(string), input.children[1].Value().(string))}
	}),
	"split": NewFunction("split", "string,string", func(input *List, c *Context) LispValue {
		return Atom{t: "string", value: str.Split(input.children[0].Value().(string), input.children[1].Value().(string))}
	}),
	"split-n": NewFunction("split-n", "string,string,int", func(input *List, c *Context) LispValue {
		return Atom{t: "string", value: str.SplitN(input.children[0].Value().(string), input.children[1].Value().(string), input.children[2].Value().(int))}
	}),
	"at": NewFunction("at", "int,list", func(input *List, c *Context) LispValue {
		i := input.children[0].Value().(int)
		l := input.children[1].(*List)
		for i < 0 {
			i += len(l.children)
		}
		if i < len(l.children) {
			return l.children[i]
		}
		return Atom{t: "error", value: fmt.Sprintf("Index '%d' out of list bounds.", i)}
	}),
	"get": NewFunction("get", "string|symbol,hash", func(input *List, c *Context) LispValue {
		s := input.children[0].Value().(string)
		h := input.children[1].(*Hash)
		if input.children[0].Type() == "string" {
			return h.vals[s]
		} else {
			return h.sym_vals[s]
		}
	}),
	"hset!": NewFunction("hset!", "string|symbol,*,hash", func(input *List, c *Context) LispValue {
		s := input.children[0].Value().(string)
		v := input.children[1]
		h := input.children[2].(*Hash)
		if input.children[0].Type() == "string" {
			h.vals[s] = v
		} else {
			h.sym_vals[s] = v
		}
		return h
	}),
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
		c.Set(name, fn)
	}
	for name, a := range aliases {
		c.Set(name, c.Get(a))
	}
}
