package main

import (
	"fmt"
	"math"
	"strconv"
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
	"len": NewFunction("len", "list|string|hash", func(input *List, c *Context) LispValue {
		if input.children[0].Type() == "list" {
			return input.children[0].(*List).Length()
		} else if input.children[0].Type() == "hash" {
			return input.children[0].(*Hash).Length()
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
			if last.Type() == "error" {
				return last
			}
		}
		return last
	}),
	"trim": NewFunction("trim", "string,string", func(input *List, c *Context) LispValue {
		return Atom{t: "string", value: str.Trim(input.children[0].Value().(string), input.children[1].Value().(string))}
	}),
	"join": NewFunction("split", "list,string", func(input *List, c *Context) LispValue {
		elms := []string{}
		for _, c := range input.children[0].(*List).children {
			if c.Type() == "string" {
				elms = append(elms, c.Value().(string))
			} else {
				elms = append(elms, c.String())
			}
		}
		return Atom{t: "string", value: str.Join(elms, input.children[1].Value().(string))}
	}),
	"split": NewFunction("split", "string,string", func(input *List, c *Context) LispValue {
		out := &List{}
		for _, s := range str.Split(input.children[0].Value().(string), input.children[1].Value().(string)) {
			out.children = append(out.children, Atom{t: "string", value: s})
		}
		return out
	}),
	"split-n": NewFunction("split-n", "string,string,int", func(input *List, c *Context) LispValue {
		out := &List{}
		for _, s := range str.SplitN(input.children[0].Value().(string), input.children[1].Value().(string), input.children[2].Value().(int)) {
			out.children = append(out.children, Atom{t: "string", value: s})
		}
		return out
	}),
	"parse-int": NewFunction("parse-int", "string", func(input *List, c *Context) LispValue {
		s := input.children[0].Value().(string)
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return Atom{t: "error", value: fmt.Sprintf("Could not convert string '%s' to an integer.", s)}
		}
		return Atom{t: "int", value: int(i)}
	}),
	"parse-float": NewFunction("parse-float", "string", func(input *List, c *Context) LispValue {
		s := input.children[0].Value().(string)
		f, err := strconv.ParseFloat(s, 10)
		if err != nil {
			return Atom{t: "error", value: fmt.Sprintf("Could not convert string '%s' to a float.", s)}
		}
		return Atom{t: "float", value: f}
	}),
	"get": NewFunction("get", "list,int", func(input *List, c *Context) LispValue {
		l := input.children[0].(*List)
		i := input.children[1].Value().(int)
		for i < 0 {
			i += len(l.children)
		}
		if i < len(l.children) {
			return l.children[i]
		}
		return Atom{t: "error", value: fmt.Sprintf("Index '%d' out of list bounds.", i)}
	}),
	"hget": NewFunction("hget", "hash,string|symbol", func(input *List, c *Context) LispValue {
		h := input.children[0].(*Hash)
		s := input.children[1].Value().(string)
		if input.children[1].Type() == "string" {
			return h.vals[s]
		} else {
			return h.sym_vals[s]
		}
	}),
	"hset!": NewFunction("hset!", "hash,string|symbol,*", func(input *List, c *Context) LispValue {
		h := input.children[0].(*Hash)
		s := input.children[1].Value().(string)
		v := input.children[2]
		if input.children[1].Type() == "string" {
			h.vals[s] = v
		} else {
			h.sym_vals[s] = v
		}
		return h
	}),
	"hcontains": NewFunction("hcontains", "hash,string|symbol", func(input *List, c *Context) LispValue {
		h := input.children[0].(*Hash)
		s := input.children[1].Value().(string)
		if input.children[1].Type() == "string" {
			_, ok := h.vals[s]
			return Atom{t: "bool", value: ok}
		} else {
			_, ok := h.sym_vals[s]
			return Atom{t: "bool", value: ok}
		}
	}),
	"type": NewFunction("type", "*", func(input *List, c *Context) LispValue {
		return Atom{t: "type", value: input.children[0].Type()}
	}),
	"int": NewFunction("int", "int|float", func(input *List, c *Context) LispValue {
		if input.children[0].Type() == "int" {
			return input.children[0]
		}
		return Atom{t: "int", value: int(input.children[0].Value().(float64))}
	}),
	"float": NewFunction("float", "int|float", func(input *List, c *Context) LispValue {
		if input.children[0].Type() == "float" {
			return input.children[0]
		}
		return Atom{t: "float", value: float64(input.children[0].Value().(int))}
	}),
	"string": NewFunction("string", "*", func(input *List, c *Context) LispValue {
		if input.children[0].Type() == "string" {
			return input.children[0]
		}
		return Atom{t: "string", value: input.children[0].String()}
	}),
	"bool": NewFunction("string", "*", func(input *List, c *Context) LispValue {
		return Atom{t: "bool", value: Boolean(input.children[0])}
	}),
	"floor": NewFunction("floor", "float", func(input *List, c *Context) LispValue {
		return Atom{t: "float", value: math.Floor(input.children[0].Value().(float64))}
	}),
	"ceil": NewFunction("ceil", "float", func(input *List, c *Context) LispValue {
		return Atom{t: "float", value: math.Ceil(input.children[0].Value().(float64))}
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
