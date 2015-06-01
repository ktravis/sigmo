package main

import (
	"fmt"
	str "strings"
)

var Special = make(map[string]func(*List, *Context) LispValue)

var TRUE = Atom{t: "bool", value: true}
var FALSE = Atom{t: "bool", value: false}
var NIL = Atom{t: "nil"}

type Context struct {
	scope  map[string]LispValue
	parent *Context
}

func (c *Context) get(identifier string) LispValue {
	a, ok := c.scope[identifier]
	if ok {
		return a
	} else if c.parent != nil {
		return c.parent.get(identifier)
	}
	v := LispValue(Atom{t: "error", value: fmt.Sprintf("Unknown identifier '%s'", identifier)})
	return v
}

func NewContext(parent *Context) *Context {
	c := Context{
		scope:  make(map[string]LispValue),
		parent: parent,
	}
	return &c
}

type LispValue interface {
	String() string
	Eval(*Context) LispValue
	Value() interface{}
	Type() string
}

type List struct {
	children []LispValue
}

func (l List) String() string {
	elms := []string{}
	for _, c := range l.children {
		elms = append(elms, c.String())
	}
	return fmt.Sprintf("(%s)", str.Join(elms, " "))
}

func (l List) Eval(c *Context) LispValue {
	output := List{}
	if len(l.children) > 0 {
		first := l.children[0]
		if first.Type() == "identifier" {
			f, special := Special[first.Value().(string)]
			if special {
				return f(&l, c)
			}
		}
		for _, a := range l.children {
			e := a.Eval(c)
			if e.Type() == "error" {
				return e
			}
			output.children = append(output.children, e)
		}
		n := output.children[0]
		if n.Type() == "function" {
			l := List{children: output.children[1:]}
			return n.(LispFunction).Call(&l, c)
		}
	}
	return &output
}

func (l List) Value() interface{} {
	return l.children
}

func (l List) Type() string {
	return "list"
}

func (l List) Length() Atom {
	return Atom{t: "int", value: len(l.children)}
}

type Atom struct {
	t     string
	value interface{}
}

func (a Atom) String() string {
	switch a.t {
	case "int":
		return fmt.Sprintf("%d", a.value)
	case "float":
		return fmt.Sprintf("%f", a.value)
	case "bool":
		return fmt.Sprintf("%t", a.value)
	case "identifier":
		return a.value.(string)
	case "string":
		return fmt.Sprintf("\"%s\"", a.value)
	}
	return a.t
}

func (a Atom) Eval(c *Context) LispValue {
	if a.t == "identifier" {
		return c.get(a.value.(string))
	}
	return a
}

func (a Atom) Value() interface{} {
	return a.value
}

func (a Atom) Type() string {
	return a.t
}

func (a Atom) Length() Atom {
	switch a.t {
	case "identifier":
		return Atom{t: "int", value: len(a.value.(string))}
	case "string":
		return Atom{t: "int", value: len(a.value.(string))}
	default:
		return NIL
	}
}

type LispFunction struct {
	arc      int
	argtypes []string
	value    func(*List, *Context) LispValue
}

func (f LispFunction) String() string {
	return "{fn}"
}

func (f LispFunction) Eval(c *Context) LispValue {
	return f
}

func (f LispFunction) Value() interface{} {
	return NIL
}

func (f LispFunction) Type() string {
	return "function"
}

func (f LispFunction) Call(args *List, c *Context) LispValue {
	// do some argc checking
	return f.value(args, c)
}

func NewFunction(fn func(*List, *Context) LispValue) LispFunction {
	return LispFunction{value: fn}
}

func Setup(c *Context) {
	Special["lambda"] = func(form *List, c *Context) LispValue {
		return NewFunction(func(args *List, outer *Context) LispValue {
			inner := NewContext(outer)
			argnames := form.children[1].(List)
			for i, a := range args.children {
				inner.scope[argnames.children[i].Value().(string)] = a
			}
			return form.children[2].Eval(inner)
		})
	}
	Special["def"] = func(form *List, c *Context) LispValue {
		if len(form.children) != 3 {
			return &Atom{t: "error", value: "Wrong number of arguments to 'def'"}
		}
		if form.children[1].Type() != "identifier" {
			return &Atom{t: "error", value: fmt.Sprintf("def expected argument 0 of type 'identifier', got type '%s'", form.children[1].Type())}
		}
		v := form.children[2].Eval(c)
		c.scope[form.children[1].Value().(string)] = v
		return v
	}
	Special["defn"] = func(form *List, c *Context) LispValue {
		if len(form.children) != 4 {
			return &Atom{t: "error", value: "Wrong number of arguments to 'defn'"}
		}
		if form.children[1].Type() != "identifier" {
			return Atom{t: "error", value: fmt.Sprintf("defn expected argument 0 of type 'identifier', got type '%s'", form.children[1].Type())}
		}
		v := NewFunction(func(args *List, outer *Context) LispValue {
			inner := NewContext(outer)
			argnames := form.children[2].(List)
			for i, a := range args.children {
				inner.scope[argnames.children[i].Value().(string)] = a
			}
			return form.children[3].Eval(inner)
		})
		c.scope[form.children[1].Value().(string)] = v
		return v
	}
	Special["do"] = func(form *List, c *Context) LispValue {
		var last LispValue = NIL
		for _, n := range form.children[1:] {
			last = n.Eval(c)
		}
		return last
	}
	Special["if"] = func(form *List, c *Context) LispValue {
		if Boolean(form.children[1].Eval(c)) {
			return form.children[2].Eval(c)
		}
		if len(form.children) > 3 {
			return form.children[3].Eval(c)
		}
		return NIL // should this be false?
	}
	Special["while"] = func(form *List, c *Context) LispValue {
		var last LispValue = NIL
		for Boolean(form.children[1].Eval(c)) {
			last = form.children[2].Eval(c)
		}
		return last
	}
	Special["assert"] = func(form *List, c *Context) LispValue {
		if Boolean(form.children[1].Eval(c)) {
			return TRUE
		}
		return Atom{t: "error", value: fmt.Sprintf("Assert failed '%s'", form.children[1].String())}
	}
}
