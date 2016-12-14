package sigmo

import (
	"fmt"
	"strings"
)

var (
	TRUE  = Atom{t: "bool", value: true}
	FALSE = Atom{t: "bool", value: false}
	NIL   = Atom{t: "nil"}
)

type Value interface {
	String() string
	Eval(Context) Value
	Value() interface{}
	Type() string
	Copy() Value
}

type Container interface {
	Value
	Append(Value)
	//Children() []Value
}

type List struct {
	children []Value
	Quoted   bool
}

func (l *List) String() string {
	elms := []string{}
	for _, c := range l.children {
		elms = append(elms, c.String())
	}
	if l.Quoted {
		return fmt.Sprintf("'(%s)", strings.Join(elms, " "))
	}
	return fmt.Sprintf("(%s)", strings.Join(elms, " "))
}

func (l *List) Eval(c Context) Value {
	if l.Quoted {
		return l
	}
	output := List{}
	if len(l.children) > 0 {
		first := l.children[0]
		if first.Type() == "identifier" {
			f, special := specialForms[first.Value().(string)]
			if special {
				return f(l, c)
			}
			m := first.Eval(c)
			if m.Type() == "macro" {
				return m.(Macro).Call(l, c)
			}
		}
		for _, a := range l.children {
			if a.Type() == "expansion" {
				ls := a.Eval(c)
				if ls.Type() != "list" {
					return &Atom{t: "error", value: fmt.Sprintf("Cannot expand value of type '%s'", ls.Type())}
				}
				output.children = append(output.children, ls.(*List).children...)
				continue
			}
			e := a.Eval(c)
			if e.Type() == "error" {
				return e
			}
			output.children = append(output.children, e)
		}
		n := output.children[0]
		if n.Type() == "function" {
			l := List{children: output.children[1:]}
			return n.(Function).Call(&l, c)
		}
	}
	return &output
}

func (l *List) Value() interface{} {
	return l.children
}

func (l *List) Copy() Value {
	n := List{Quoted: l.Quoted}
	for _, el := range l.children {
		n.children = append(n.children, el.Copy())
	}
	return &n
}

func (l *List) Type() string {
	return "list"
}

func (l *List) Length() Atom {
	return Atom{t: "int", value: len(l.children)}
}

func (l *List) Append(v Value) {
	l.children = append(l.children, v)
}

type Hash struct {
	pairs    []Value
	vals     map[string]Value
	sym_vals map[string]Value
}

func MakeHash(pairs []Value, c Context) Value {
	h := Hash{vals: make(map[string]Value),
		sym_vals: make(map[string]Value)}
	var last Value = NIL
	for i, v := range pairs {
		if i%2 == 0 {
			last = v.Eval(c)
			if !(last.Type() == "string" || last.Type() == "symbol") {
				return Atom{t: "error", value: fmt.Sprintf("Invalid key type '%s' for hash.", last.Type())}
			}
		} else {
			if last.Type() == "string" {
				h.vals[last.Value().(string)] = v.Eval(c)
			} else if last.Type() == "symbol" {
				h.sym_vals[last.Value().(string)] = v.Eval(c)
			}
			last = NIL
		}
	}
	if last != NIL {
		if last.Type() == "string" {
			h.vals[last.Value().(string)] = NIL
		} else if last.Type() == "symbol" {
			h.sym_vals[last.Value().(string)] = NIL
		}
	}
	return &h
}

func (h *Hash) String() string {
	elms := []string{}
	for k, v := range h.vals {
		elms = append(elms, k, v.String())
	}
	for k, v := range h.sym_vals {
		elms = append(elms, k, v.String())
	}
	return fmt.Sprintf("{%s}", strings.Join(elms, " "))
}

func (h *Hash) Eval(c Context) Value {
	if len(h.pairs) > 0 {
		return MakeHash(h.pairs, c)
	}
	return h
}

func (h *Hash) Value() interface{} {
	return h.vals
}

func (h *Hash) Copy() Value {
	n := Hash{vals: make(map[string]Value),
		sym_vals: make(map[string]Value)}
	for k, v := range h.vals {
		n.vals[k] = v.Copy()
	}
	for k, v := range h.sym_vals {
		n.sym_vals[k] = v.Copy()
	}
	return &n
}

func (h *Hash) Type() string {
	return "hash"
}

func (h *Hash) Length() Atom {
	return Atom{t: "int", value: len(h.vals) + len(h.sym_vals)}
}

func (h *Hash) Append(l Value) {
	h.pairs = append(h.pairs, l)
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
	case "symbol":
		return fmt.Sprintf("%s", a.value)
	case "string":
		return fmt.Sprintf("\"%s\"", a.value)
	case "type":
		return fmt.Sprintf("#%s", a.value)
	}
	return a.t
}

func (a Atom) Eval(c Context) Value {
	if a.t == "identifier" || a.t == "expansion" {
		return c.Get(a.value.(string))
	}
	return a
}

func (a Atom) Value() interface{} {
	return a.value
}

func (a Atom) Copy() Value {
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

type Function func(*List, Context) Value

func (f Function) String() string {
	return "{fn}"
}

func (f Function) Eval(c Context) Value {
	return f
}

func (f Function) Value() interface{} {
	return NIL
}

func (f Function) Copy() Value {
	return f
}

func (f Function) Type() string {
	return "function"
}

func (f Function) Call(args *List, c Context) Value {
	return f(args, c)
}

// "*"

func NewFunction(name string, types string, fn func(*List, Context) Value) Function {
	split := strings.Split(types, ",")
	if len(split) == 1 && split[0] == "**" {
		return Function(fn)
	}
	wrapped := func(args *List, c Context) Value {
		expanded := false
		for i, t := range split {
			if i >= len(args.children) {
				if t == "**" {
					break
				}
				return Atom{t: "error", value: fmt.Sprintf("Function '%s' expected %d args, only got %d.", name, len(split), len(args.children))}
			}
			a := args.children[i].Type()
			switch t {
			case "**":
				expanded = true
				break
			case "*":
				continue
			case "+":
				if i < 1 {
					return Atom{t: "error", value: fmt.Sprintf("Function '%s' cannot have '+' as first argtype parameter.", name)}
				}
				for _, r := range args.children[i:] {
					if !strings.Contains(split[i-1], r.Type()) {
						return Atom{t: "error", value: fmt.Sprintf("Function '%s' cannot have '%s' as argtype, expected '%s'.", name, r.Type(), split[i-1])}
					}
				}
				expanded = true
				break
			default:
				if !strings.Contains(t, a) {
					return Atom{t: "error", value: fmt.Sprintf("Function '%s' cannot have '%s' as argtype, expected '%s'.", name, a, t)}
				}

			}
		}
		if !expanded && len(args.children) > len(split) {
			return &Atom{t: "error", value: fmt.Sprintf("Function '%s' expected %d args, but got %d.", name, len(split), len(args.children))}
		}
		return fn(args, c)
	}
	return Function(wrapped)
}

type Macro func(*List, Context) Value

func (m Macro) String() string {
	return "{macro}"
}

func (m Macro) Eval(c Context) Value {
	return m
}

func (m Macro) Value() interface{} {
	return NIL
}

func (m Macro) Copy() Value {
	return m
}

func (m Macro) Type() string {
	return "macro"
}

func (m Macro) Call(args *List, c Context) Value {
	return m(args, c)
}

func NewMacro(name string, types string, fn func(*List, Context) Value) Macro {
	split := strings.Split(types, ",")
	if len(split) == 1 && split[0] == "**" {
		return Macro(fn)
	}
	wrapped := func(args *List, c Context) Value {
		expanded := false
		for i, t := range split {
			if i >= len(args.children) {
				return &Atom{t: "error", value: fmt.Sprintf("Macro '%s' expected %d args, only got %d.", name, len(split), len(args.children))}
			}
			a := args.children[i].Type()
			switch t {
			case "**":
				expanded = true
				break
			case "*":
				continue
			case "+":
				if i < 1 {
					return &Atom{t: "error", value: fmt.Sprintf("Macro '%s' cannot have '+' as first argtype parameter.", name)}
				}
				for _, r := range args.children[i:] {
					if !strings.Contains(split[i-1], r.Type()) {
						return &Atom{t: "error", value: fmt.Sprintf("Macro '%s' cannot have '%s' as argtype, expected '%s'.", name, r.Type(), split[i-1])}
					}
				}
				expanded = true
				break
			default:
				if !strings.Contains(t, a) {
					return &Atom{t: "error", value: fmt.Sprintf("Macro '%s' cannot have '%s' as argtype, expected '%s'.", name, a, t)}
				}

			}
		}
		if !expanded && len(args.children) > len(split) {
			return &Atom{t: "error", value: fmt.Sprintf("Macro '%s' expected %d args, but got %d.", name, len(split), len(args.children))}
		}
		return fn(args, c)
	}
	return Macro(wrapped)
}

func ParseArgs(argnames *List, argvals *List, c Context) Value {
	for i, a := range argnames.children {
		if i >= len(argvals.children) && a.Type() != "expansion" {
			return Atom{t: "error", value: "Not enough arguments to function"}
		}
		switch a.Type() {
		case "identifier":
			c.Set(a.Value().(string), argvals.children[i].Copy())
		case "expansion":
			children := []Value{}
			for _, x := range argvals.children[i:] {
				children = append(children, x.Copy())
			}
			c.Set(a.Value().(string), &List{children: children})
			return nil
		case "typed id":
			x := a.Value().(string)
			j := strings.Index(x, "#")
			t := x[j+1:]
			x = x[:j]
			if argvals.children[i].Type() != t {
				if conv := c.Get(t); conv.Type() == "function" {
					temp := &List{children: []Value{argvals.children[i]}}
					c.Set(x, conv.(Function).Call(temp, c))
					continue
				}
				return Atom{t: "error", value: fmt.Sprintf("Expected argument '%s' of type '%s', got '%s'", x, t, argvals.children[i].Type())}
			}
			c.Set(x, argvals.children[i])
		default:
			return Atom{t: "error", value: fmt.Sprintf("Cannot use type '%s' in function argument list", a.Type())}
		}
	}
	return nil
}
