package main

import (
	"fmt"
	str "strings"
)

var Special = make(map[string]func(*List, Context) LispValue)

var TRUE = Atom{t: "bool", value: true}
var FALSE = Atom{t: "bool", value: false}
var NIL = Atom{t: "nil"}

type Context interface {
	Get(string) LispValue
	Set(string, LispValue)
	Parent() Context
}

type BaseContext struct {
	scope  map[string]LispValue
	parent Context
}

func (c *BaseContext) Get(identifier string) LispValue {
	a, ok := c.scope[identifier]
	if ok {
		return a
	}
	p := c.Parent()
	if p != nil {
		return p.Get(identifier)
	}
	v := LispValue(Atom{t: "error", value: fmt.Sprintf("Unknown identifier '%s'", identifier)})
	return v
}

func (c *BaseContext) Set(identifier string, l LispValue) {
	c.scope[identifier] = l
}

func (c *BaseContext) Parent() Context {
	return c.parent
}

func NewContext(parent Context) Context {
	c := BaseContext{
		scope:  make(map[string]LispValue),
		parent: parent,
	}
	return &c
}

// passthrough, can't "set"
type ReadContext struct {
	scope  map[string]LispValue
	parent Context
}

func (c *ReadContext) Get(identifier string) LispValue {
	a, ok := c.scope[identifier]
	if ok {
		return a
	}
	return c.Parent().Get(identifier)
}

func (c *ReadContext) Set(identifier string, l LispValue) {
	fmt.Println("set stuff", identifier, l)
	c.Parent().Set(identifier, l)
}

func (c *ReadContext) Parent() Context {
	return c.parent
}

func NewReadContext(m map[string]LispValue, parent Context) Context {
	if parent == nil {
		panic("Cannot create a ReadContext with no parent.")
	}
	return &ReadContext{scope: m, parent: parent}
}

type LispValue interface {
	String() string
	Eval(Context) LispValue
	Value() interface{}
	Type() string
	Copy() LispValue
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

func (l List) Eval(c Context) LispValue {
	output := List{}
	if len(l.children) > 0 {
		first := l.children[0]
		if first.Type() == "identifier" {
			f, special := Special[first.Value().(string)]
			if special {
				return f(&l, c)
			}
			m := first.Eval(c)
			if m.Type() == "macro" {
				return m.(LispMacro).Call(&l, c)
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

func (l List) Copy() LispValue {
	n := List{}
	for _, el := range l.children {
		n.children = append(n.children, el.Copy())
	}
	return n
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

func (a Atom) Eval(c Context) LispValue {
	if a.t == "identifier" {
		return c.Get(a.value.(string))
	}
	return a
}

func (a Atom) Value() interface{} {
	return a.value
}

func (a Atom) Copy() LispValue {
	return Atom{t: a.t, value: a.value}
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
	argc     int
	argtypes []string
	value    func(*List, Context) LispValue
}

func (f LispFunction) String() string {
	return "{fn}"
}

func (f LispFunction) Eval(c Context) LispValue {
	return f
}

func (f LispFunction) Value() interface{} {
	return NIL
}

func (f LispFunction) Copy() LispValue {
	return LispFunction{argc: f.argc, argtypes: f.argtypes, value: f.value}
}

func (f LispFunction) Type() string {
	return "function"
}

func (f LispFunction) Call(args *List, c Context) LispValue {
	// do some argc checking
	return f.value(args, c)
}

func NewFunction(fn func(*List, Context) LispValue) LispFunction {
	return LispFunction{value: fn}
}

type LispMacro struct {
	argc     int
	argtypes []string
	value    func(*List, Context) LispValue
}

func (f LispMacro) String() string {
	return "{macro}"
}

func (f LispMacro) Eval(c Context) LispValue {
	return f
}

func (f LispMacro) Value() interface{} {
	return NIL
}

func (f LispMacro) Copy() LispValue {
	return LispMacro{argc: f.argc, argtypes: f.argtypes, value: f.value}
}

func (f LispMacro) Type() string {
	return "macro"
}

func (f LispMacro) Call(args *List, c Context) LispValue {
	// do some argc checking
	return f.value(args, c)
}

func NewMacro(fn func(*List, Context) LispValue) LispMacro {
	return LispMacro{value: fn}
}

func Setup(c Context) {
	Special["lambda"] = func(form *List, c Context) LispValue {
		return NewFunction(func(args *List, outer Context) LispValue {
			inner := NewContext(outer)
			argnames := form.children[1].(List)
			for i, a := range args.children {
				inner.Set(argnames.children[i].Value().(string), a)
			}
			return form.children[2].Eval(inner)
		})
	}
	Special["def"] = func(form *List, c Context) LispValue {
		if len(form.children) != 3 {
			return &Atom{t: "error", value: "Wrong number of arguments to 'def'"}
		}
		if form.children[1].Type() != "identifier" {
			return &Atom{t: "error", value: fmt.Sprintf("def expected argument 0 of type 'identifier', got type '%s'", form.children[1].Type())}
		}
		v := form.children[2].Eval(c)
		c.Set(form.children[1].Value().(string), v)
		return v
	}
	Special["do"] = func(form *List, c Context) LispValue {
		var last LispValue = NIL
		for _, n := range form.children[1:] {
			last = n.Eval(c)
		}
		return last
	}
	Special["if"] = func(form *List, c Context) LispValue {
		if Boolean(form.children[1].Eval(c)) {
			return form.children[2].Eval(c)
		}
		if len(form.children) > 3 {
			return form.children[3].Eval(c)
		}
		return NIL // should this be false?
	}
	Special["while"] = func(form *List, c Context) LispValue {
		var last LispValue = NIL
		for Boolean(form.children[1].Eval(c)) {
			last = form.children[2].Eval(c)
		}
		return last
	}
	Special["for"] = func(form *List, c Context) LispValue {
		out := List{}
		if form.children[1].Type() != "list" {
			return Atom{t: "error", value: "First argument to 'for' must be a list of form '(identifier list)'"}
		}
		params := form.children[1].(List)
		if params.children[0].Type() != "identifier" {
			return Atom{t: "error", value: "First argument to 'for' must be a list of form '(identifier list)'"}
		}
		ident := params.children[0].Value().(string)
		inner := NewContext(c)
		temp := params.children[1].Eval(inner)
		if temp.Type() != "list" {
			return Atom{t: "error", value: "Second argument of 'for' parameters must evaluate to a list"}
		}
		ls := temp.(*List)
		for _, n := range ls.children {
			inner.Set(ident, n.Eval(inner))
			out.children = append(out.children, form.children[2].Eval(inner))
		}
		return out
	}
	Special["let"] = func(form *List, c Context) LispValue {
		if form.children[1].Type() != "list" {
			return Atom{t: "error", value: "First argument to 'let' must be a list of form '(identifier list)'"}
		}
		params := form.children[1].(List)
		inner := NewContext(c)
		for i := 0; i < len(params.children); i += 2 {
			if params.children[i].Type() != "identifier" {
				return Atom{t: "error", value: "Even parameters to 'let' must be identifiers"}
			}
			ident := params.children[i].Value().(string)
			var val LispValue = NIL
			if len(params.children) > i+1 {
				val = params.children[i+1].Eval(inner)
			}
			inner.Set(ident, val)
		}
		var last LispValue = NIL
		for _, n := range form.children[2:] {
			last = n.Eval(inner)
		}
		return last
	}
	Special["assert"] = func(form *List, c Context) LispValue {
		if Boolean(form.children[1].Eval(c)) {
			return TRUE
		}
		return Atom{t: "error", value: fmt.Sprintf("Assert failed '%s'", form.children[1].String())}
	}
	Special["macro"] = func(form *List, c Context) LispValue {
		if form.children[1].Type() != "identifier" {
			return Atom{t: "error", value: fmt.Sprintf("macro expected argument 0 of type 'identifier', got type '%s'", form.children[1].Type())}
		}
		name := form.children[1].Value().(string)
		if form.children[2].Type() != "list" {
			return Atom{t: "error", value: fmt.Sprintf("macro expected argument 1 of type 'list', got type '%s'", form.children[2].Type())}
		}
		params := form.children[2].(List)
		v := NewMacro(func(args *List, outer Context) LispValue {
			var last LispValue = NIL
			for _, r := range form.children[3:] {
				p := r.Copy()
				for j, a := range args.children[1:] {
					p = NestedReplace(p, params.children[j].Value().(string), a)
				}
				last = p.Eval(outer)
			}
			return last
		})
		c.Set(name, v)
		return v
	}
	Special["debug"] = func(form *List, c Context) LispValue {
		fmt.Println(c.(*BaseContext).scope)
		return NIL
	}
}
