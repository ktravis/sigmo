package main

import (
	"bufio"
	"fmt"
	"os"
	str "strings"
)

var Special = make(map[string]func(*List, *Context) LispValue)

var TRUE = Atom{t: "bool", value: true}
var FALSE = Atom{t: "bool", value: false}
var NIL = Atom{t: "nil"}

type Context struct {
	scope      map[string]LispValue
	parent     *Context
	ns         string
	namespaces map[string]*Context
}

func (c *Context) Get(identifier string) LispValue {
	if str.Contains(identifier, "/") {
		paths := str.Split(identifier, "/")
		if c.ns != str.Join(paths[:len(paths)-1], "/") {
			return c.GetFromNamespace(paths[:len(paths)-1], paths[len(paths)-1])
		}
		identifier = paths[len(paths)-1]
	}
	a, ok := c.scope[identifier]
	if ok {
		return a
	}
	p := c.parent
	if p != nil {
		return p.Get(identifier)
	}
	v := LispValue(Atom{t: "error", value: fmt.Sprintf("Unknown identifier '%s'", identifier)})
	return v
}

func (c *Context) Set(identifier string, l LispValue) LispValue {
	if str.Contains(identifier, "/") {
		paths := str.Split(identifier, "/")
		if c.ns != str.Join(paths[:len(paths)-1], "/") {
			return c.SetInNamespace(paths[:len(paths)-1], paths[len(paths)-1], l)
		}
		identifier = paths[len(paths)-1]
	}
	c.scope[identifier] = l
	return l
}

func (c *Context) SetExisting(identifier string, l LispValue) LispValue {
	_, ok := c.scope[identifier]
	if ok {
		c.scope[identifier] = l
		return l
	}
	p := c.parent
	if p != nil {
		return p.SetExisting(identifier, l)
	}
	v := LispValue(Atom{t: "error", value: fmt.Sprintf("Unknown identifier '%s'", identifier)})
	return v
}

func (c *Context) GetFromNamespace(paths []string, identifier string) LispValue {
	if c.parent == nil {
		context := c
		var ok bool
		for _, n := range paths {
			if context, ok = context.namespaces[n]; !ok {
				return LispValue(Atom{t: "error", value: fmt.Sprintf("Unknown namespace '%s'", str.Join(paths, "/"))})
			}
		}
		return context.Get(identifier)
	}
	return c.parent.GetFromNamespace(paths, identifier)
}

func (c *Context) SetInNamespace(paths []string, identifier string, l LispValue) LispValue {
	root := c
	for root.parent != nil {
		root = root.parent
	}
	for i, p := range paths {
		x, ok := root.namespaces[p]
		if !ok {
			root.namespaces[p] = NewContext(root)
			x = root.namespaces[p]
			x.ns = str.Join(paths[:i+1], "/")
		}
		root = x
	}
	root.scope[identifier] = l
	return l
}

func GetNamespace(start *Context, path string) *Context {
	paths := str.Split(path, "/")
	root := start
	for root.parent != nil {
		root = root.parent
	}
	curr := root
	for i, p := range paths[:len(paths)-1] {
		x, ok := curr.namespaces[p]
		if !ok {
			curr.namespaces[p] = NewContext(curr)
			x = curr.namespaces[p]
			x.ns = str.Join(paths[:i], "/")
		}
		curr = x
	}
	x, exists := curr.namespaces[paths[len(paths)-1]]
	if !exists {
		x = NewContext(curr)
		curr.namespaces[paths[len(paths)-1]] = x
		x.ns = str.Join(paths, "/")
	}
	return x
}

func NewContext(parent *Context) *Context {
	c := Context{
		scope:      make(map[string]LispValue),
		parent:     parent,
		namespaces: make(map[string]*Context),
	}
	return &c
}

type LispValue interface {
	String() string
	Eval(*Context) LispValue
	Value() interface{}
	Type() string
	Copy() LispValue
}

type LispContainer interface {
	LispValue
	Append(LispValue)
	//Children() []LispValue
}

type List struct {
	children []LispValue
	Quoted   bool
}

func (l *List) String() string {
	elms := []string{}
	for _, c := range l.children {
		elms = append(elms, c.String())
	}
	if l.Quoted {
		return fmt.Sprintf("'(%s)", str.Join(elms, " "))
	}
	return fmt.Sprintf("(%s)", str.Join(elms, " "))
}

func (l *List) Eval(c *Context) LispValue {
	if l.Quoted {
		return l
	}
	output := List{}
	if len(l.children) > 0 {
		first := l.children[0]
		if first.Type() == "identifier" {
			f, special := Special[first.Value().(string)]
			if special {
				return f(l, c)
			}
			m := first.Eval(c)
			if m.Type() == "macro" {
				return m.(LispMacro).Call(l, c)
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
			return n.(LispFunction).Call(&l, c)
		}
	}
	return &output
}

func (l *List) Value() interface{} {
	return l.children
}

func (l *List) Copy() LispValue {
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

func (l *List) Append(v LispValue) {
	l.children = append(l.children, v)
}

type Hash struct {
	pairs    []LispValue
	vals     map[string]LispValue
	sym_vals map[string]LispValue
}

func MakeHash(pairs []LispValue, c *Context) LispValue {
	h := Hash{vals: make(map[string]LispValue),
		sym_vals: make(map[string]LispValue)}
	var last LispValue = NIL
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
	return fmt.Sprintf("{%s}", str.Join(elms, " "))
}

func (h *Hash) Eval(c *Context) LispValue {
	if len(h.pairs) > 0 {
		return MakeHash(h.pairs, c)
	}
	return h
}

func (h *Hash) Value() interface{} {
	return h.vals
}

func (h *Hash) Copy() LispValue {
	n := Hash{vals: make(map[string]LispValue),
		sym_vals: make(map[string]LispValue)}
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

func (h *Hash) Append(l LispValue) {
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

func (a Atom) Eval(c *Context) LispValue {
	if a.t == "identifier" || a.t == "expansion" {
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

func (f LispFunction) Copy() LispValue {
	return LispFunction{argc: f.argc, argtypes: f.argtypes, value: f.value}
}

func (f LispFunction) Type() string {
	return "function"
}

func (f LispFunction) Call(args *List, c *Context) LispValue {
	// do some argc checking
	return f.value(args, c)
}

// "*"

func NewFunction(name string, types string, fn func(*List, *Context) LispValue) LispFunction {
	split := str.Split(types, ",")
	wrapped := func(args *List, c *Context) LispValue {
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
					if !str.Contains(split[i-1], r.Type()) {
						return Atom{t: "error", value: fmt.Sprintf("Function '%s' cannot have '%s' as argtype, expected '%s'.", name, r.Type(), split[i-1])}
					}
				}
				expanded = true
				break
			default:
				if !str.Contains(t, a) {
					return Atom{t: "error", value: fmt.Sprintf("Function '%s' cannot have '%s' as argtype, expected '%s'.", name, a, t)}
				}

			}
		}
		if !expanded && len(args.children) > len(split) {
			return &Atom{t: "error", value: fmt.Sprintf("Function '%s' expected %d args, but got %d.", name, len(split), len(args.children))}
		}
		return fn(args, c)
	}
	return LispFunction{value: wrapped}
}

type LispMacro struct {
	argc     int
	argtypes []string
	value    func(*List, *Context) LispValue
}

func (f LispMacro) String() string {
	return "{macro}"
}

func (f LispMacro) Eval(c *Context) LispValue {
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

func (f LispMacro) Call(args *List, c *Context) LispValue {
	return f.value(args, c)
}

func NewMacro(name string, types string, fn func(*List, *Context) LispValue) LispMacro {
	split := str.Split(types, ",")
	wrapped := func(args *List, c *Context) LispValue {
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
					if !str.Contains(split[i-1], r.Type()) {
						return &Atom{t: "error", value: fmt.Sprintf("Macro '%s' cannot have '%s' as argtype, expected '%s'.", name, r.Type(), split[i-1])}
					}
				}
				expanded = true
				break
			default:
				if !str.Contains(t, a) {
					return &Atom{t: "error", value: fmt.Sprintf("Macro '%s' cannot have '%s' as argtype, expected '%s'.", name, a, t)}
				}

			}
		}
		if !expanded && len(args.children) > len(split) {
			return &Atom{t: "error", value: fmt.Sprintf("Macro '%s' expected %d args, but got %d.", name, len(split), len(args.children))}
		}
		return fn(args, c)
	}
	return LispMacro{value: wrapped}
}

func ParseArgs(argnames *List, argvals *List, c *Context) LispValue {
	for i, a := range argnames.children {
		if i >= len(argvals.children) && a.Type() != "expansion" {
			return Atom{t: "error", value: "Not enough arguments to function"}
		}
		switch a.Type() {
		case "identifier":
			c.Set(a.Value().(string), argvals.children[i])
		case "expansion":
			c.Set(a.Value().(string), &List{children: argvals.children[i:]})
			return nil
		case "typed id":
			x := a.Value().(string)
			j := str.Index(x, "#")
			t := x[j+1:]
			x = x[:j]
			if argvals.children[i].Type() != t {
				if conv := c.Get(t); conv.Type() == "function" {
					temp := &List{children: []LispValue{argvals.children[i]}}
					c.Set(x, conv.(LispFunction).Call(temp, c))
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

func Setup(c *Context) {
	Special["lambda"] = func(form *List, c *Context) LispValue {
		// TODO not enough args to func
		copied := make(map[string]LispValue)
		for k, v := range c.scope {
			copied[k] = v
		}
		return NewFunction("anonymous", "**", func(args *List, outer *Context) LispValue {
			inner := NewContext(outer)
			inner.scope = copied
			argnames := form.children[1].(*List)
			if err := ParseArgs(argnames, args, inner); err != nil {
				return err
			}
			return form.children[2].Eval(inner)
		})
	}
	Special["def"] = func(form *List, c *Context) LispValue {
		if len(form.children) != 3 {
			return Atom{t: "error", value: "Wrong number of arguments to 'def'"}
		}
		if form.children[1].Type() != "identifier" {
			return Atom{t: "error", value: fmt.Sprintf("def expected argument 0 of type 'identifier', got type '%s'", form.children[1].Type())}
		}
		v := form.children[2].Eval(c)
		if v.Type() == "error" {
			return v
		}
		c.Set(form.children[1].Value().(string), v)
		return v
	}
	Special["do"] = func(form *List, c *Context) LispValue {
		var last LispValue = NIL
		for _, n := range form.children[1:] {
			last = n.Eval(c)
			if last.Type() == "error" {
				return last
			}
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
			for _, n := range form.children[2:] {
				last = n.Eval(c)
				if last.Type() == "error" {
					return last
				}
			}
		}
		return last
	}
	Special["for"] = func(form *List, c *Context) LispValue {
		out := List{}
		if form.children[1].Type() != "list" {
			return Atom{t: "error", value: "First argument to 'for' must be a list of form '(identifier list)'"}
		}
		params := form.children[1].(*List)
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
			x := n.Eval(inner)
			if x.Type() == "error" {
				return x
			}
			inner.Set(ident, x)
			x = form.children[2].Eval(inner)
			if x.Type() == "error" {
				return x
			}
			out.children = append(out.children, x)
		}
		return &out
	}
	Special["let"] = func(form *List, c *Context) LispValue {
		if form.children[1].Type() != "list" {
			return Atom{t: "error", value: "First argument to 'let' must be a list of form '(identifier list)'"}
		}
		params := form.children[1].(*List)
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
			if last.Type() == "error" {
				return last
			}
		}
		return last
	}
	Special["assert"] = func(form *List, c *Context) LispValue {
		if Boolean(form.children[1].Eval(c)) {
			return TRUE
		}
		return Atom{t: "error", value: fmt.Sprintf("Assert failed '%s'", form.children[1].String())}
	}
	Special["input"] = func(form *List, c *Context) LispValue {
		reader := bufio.NewReader(os.Stdin)
		in, err := reader.ReadString('\n')
		if err != nil {
			return Atom{t: "error", value: err}
		}
		return Atom{t: "string", value: in}
	}
	Special["macro"] = func(form *List, c *Context) LispValue {
		if form.children[1].Type() != "identifier" {
			return Atom{t: "error", value: fmt.Sprintf("macro expected argument 0 of type 'identifier', got type '%s'", form.children[1].Type())}
		}
		name := form.children[1].Value().(string)
		if form.children[2].Type() != "list" {
			return Atom{t: "error", value: fmt.Sprintf("macro expected argument 1 of type 'list', got type '%s'", form.children[2].Type())}
		}
		v := NewMacro(name, "**", func(args *List, outer *Context) LispValue {
			argnames := form.children[2].(*List)
			subs := make(map[string]LispValue)
			for i, a := range argnames.children {
				if i >= len(args.children)-1 {
					return Atom{t: "error", value: fmt.Sprintf("Not enough arguments to macro '%s'. Expected %d, got %d.", name, len(argnames.children), len(args.children)-1)}
				}
				if a.Type() == "identifier" {
					subs[a.Value().(string)] = args.children[i+1]
				} else if a.Type() == "expansion" {
					subs[a.Value().(string)] = &List{children: args.children[1+i:]}
					break
				}
			}
			// error needs to bubble here somewhere
			var last LispValue = NIL

			for _, r := range form.children[3:] {
				p := r.Copy()
				for _, b := range NestedReplace(p, &subs) {
					last = b.Eval(outer)
					if last.Type() == "error" {
						return last
					}
				}
			}
			return last
		})
		c.Set(name, v)
		return v
	}
	Special["debug"] = func(form *List, c *Context) LispValue {
		fmt.Println("ns", c.ns)
		fmt.Println("scope", c.scope)
		fmt.Println("namespaces", c.namespaces)
		fmt.Println("parent", c.parent)
		return NIL
	}
	Special["set!"] = func(form *List, c *Context) LispValue {
		if form.children[1].Type() != "identifier" {
			return Atom{t: "error", value: fmt.Sprintf("set! expected argument 0 of type 'identifier', got type '%s'", form.children[1].Type())}
		}
		x := form.children[2].Eval(c)
		if x.Type() == "error" {
			return x
		}
		return c.SetExisting(form.children[1].Value().(string), x)
	}
	Special["namespace"] = func(form *List, c *Context) LispValue {
		if form.children[1].Type() != "identifier" {
			return Atom{t: "error", value: fmt.Sprintf("namespace! expected argument 0 of type 'identifier', got type '%s'", form.children[1].Type())}
		}
		ns := GetNamespace(c, form.children[1].Value().(string))
		var last LispValue = NIL
		for _, x := range form.children[2:] {
			last = x.Eval(ns)
			if last.Type() == "error" {
				return last
			}
		}
		return last
	}
	// TODO: add an "as" ie (import core/math m) or (import core/math *)
	Special["import"] = func(form *List, c *Context) LispValue {
		if form.children[1].Type() == "identifier" {
			return LoadFile(fmt.Sprintf("%s/%s.lsp", os.Getenv("LSP_ROOT"), form.children[1].Value().(string)), c)
		}
		if form.children[1].Type() == "string" {
			return LoadFile(form.children[1].Value().(string), c)
		}
		return Atom{t: "error", value: fmt.Sprintf("import expected argument 0 of type 'identifier' (namespace) or 'string', got type '%s'", form.children[1].Type())}
	}
	Special["guard"] = func(form *List, c *Context) LispValue {
		res := form.children[1].Eval(c)
		if res.Type() == "error" {
			wrapped := Atom{t: "string", value: res.Value().(string)}
			if len(form.children) > 2 {
				handler := form.children[2].Eval(c)
				if handler.Type() == "function" {
					args := &List{children: []LispValue{wrapped}}
					return handler.(LispFunction).Call(args, c)
				} else if handler.Type() == "macro" {
					args := &List{children: []LispValue{NIL, wrapped}}
					return handler.(LispMacro).Call(args, c)
				} else {
					return Atom{t: "error", value: fmt.Sprintf("guard expected argument 1 of type 'function', got type '%s'", handler.Type())}
				}
			}
			return NIL
		} else {
			return res
		}
	}
	Special["cond"] = func(form *List, c *Context) LispValue {
		for _, pair := range form.children[1:] {
			if pair.Type() != "list" {
				return Atom{t: "error", value: fmt.Sprintf("Statements in body of 'cond' must be of type 'list', not '%s'", pair.Type())}
			}
			l := pair.(*List)
			if len(l.children) != 2 {
				return Atom{t: "error", value: "Statements in body of 'cond' should have length of two (bool body)"}
			}
			if Boolean(l.children[0].Eval(c)) {
				return l.children[1].Eval(c)
			}
		}
		return NIL
	}
	// TODO: (cond (bool op) (bool op) ...)
}
