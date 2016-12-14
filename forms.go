package sigmo

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
)

type Form func(*List, Context) Value

var specialForms map[string]Form

func lambdaForm(form *List, c Context) Value {
	// TODO: not enough args to func
	copied := c.CopyLocals()
	return NewFunction("anonymous", "**", func(args *List, outer Context) Value {
		inner := NewContext(outer)
		for k, v := range copied {
			inner.Set(k, v.Copy())
		}
		// TODO: check this!
		argnames := form.children[1].(*List)
		if err := ParseArgs(argnames, args, inner); err != nil {
			return err
		}
		return form.children[2].Eval(inner)
	})
}

func defForm(form *List, c Context) Value {
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

func doForm(form *List, c Context) Value {
	var last Value = NIL
	for _, n := range form.children[1:] {
		last = n.Eval(c)
		if last.Type() == "error" {
			return last
		}
	}
	return last
}

func ifForm(form *List, c Context) Value {
	if Boolean(form.children[1].Eval(c)) {
		return form.children[2].Eval(c)
	}
	if len(form.children) > 3 {
		return form.children[3].Eval(c)
	}
	return NIL // should this be false?
}

func whileForm(form *List, c Context) Value {
	var last Value = NIL
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

func forForm(form *List, c Context) Value {
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

func letForm(form *List, c Context) Value {
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
		var val Value = NIL
		if len(params.children) > i+1 {
			val = params.children[i+1].Eval(inner)
		}
		inner.Set(ident, val)
	}
	var last Value = NIL
	for _, n := range form.children[2:] {
		last = n.Eval(inner)
		if last.Type() == "error" {
			return last
		}
	}
	return last
}

// TODO: this and input don't need to be special forms
func assertForm(form *List, c Context) Value {
	if Boolean(form.children[1].Eval(c)) {
		return TRUE
	}
	return Atom{t: "error", value: fmt.Sprintf("Assert failed '%s'", form.children[1].String())}
}

func inputForm(form *List, c Context) Value {
	reader := bufio.NewReader(os.Stdin)
	in, err := reader.ReadString('\n')
	if err != nil {
		return Atom{t: "error", value: err}
	}
	return Atom{t: "string", value: in}
}

func macroForm(form *List, c Context) Value {
	if form.children[1].Type() != "identifier" {
		return Atom{t: "error", value: fmt.Sprintf("macro expected argument 0 of type 'identifier', got type '%s'", form.children[1].Type())}
	}
	name := form.children[1].Value().(string)
	if form.children[2].Type() != "list" {
		return Atom{t: "error", value: fmt.Sprintf("macro expected argument 1 of type 'list', got type '%s'", form.children[2].Type())}
	}
	v := NewMacro(name, "**", func(args *List, outer Context) Value {
		argnames := form.children[2].(*List)
		subs := make(map[string]Value)
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
		var last Value = NIL

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

//func debugForm(form *List, c Context) Value {
//fmt.Println("ns", c.ns)
//fmt.Println("scope", c.scope)
//fmt.Println("namespaces", c.namespaces)
//fmt.Println("parent", c.parent)
//return NIL
//}

func setBangForm(form *List, c Context) Value {
	if form.children[1].Type() != "identifier" {
		return Atom{t: "error", value: fmt.Sprintf("set! expected argument 0 of type 'identifier', got type '%s'", form.children[1].Type())}
	}
	x := form.children[2].Eval(c)
	if x.Type() == "error" {
		return x
	}
	return c.SetExisting(form.children[1].Value().(string), x)
}

func namespaceForm(form *List, c Context) Value {
	if form.children[1].Type() != "identifier" {
		return Atom{t: "error", value: fmt.Sprintf("namespace! expected argument 0 of type 'identifier', got type '%s'", form.children[1].Type())}
	}
	ns := c.Namespace(form.children[1].Value().(string))
	var last Value = NIL
	for _, x := range form.children[2:] {
		last = x.Eval(ns)
		if last.Type() == "error" {
			return last
		}
	}
	return last
}

// TODO: add an "as" ie (import core/math m) or (import core/math *)
func importForm(form *List, c Context) Value {
	var fname string
	switch form.children[1].Type() {
	case "identifier":
		fname = fmt.Sprintf("%s/%s.mo", os.Getenv("SIGMO_ROOT"), form.children[1].Value().(string))
	case "string":
		fname = form.children[1].Value().(string)
	default:
		return Atom{t: "error", value: fmt.Sprintf("import expected argument 0 of type 'identifier' (namespace) or 'string', got type '%s'", form.children[1].Type())}
	}
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return Atom{t: "error", value: fmt.Sprintf("error during import of '%s': %v", fname, err)}
	}

	tok := Tokenize(string(data))
	nodes, err := Parse(tok)
	if err != nil {
		return Atom{t: "error", value: fmt.Sprintf("error during import of '%s': %v", fname, err)}
	}

	var last Value
	for _, n := range nodes {
		v := n.Eval(c)
		if v.Type() == "error" {
			return v
		}
		last = v
	}
	return last
}

func guardForm(form *List, c Context) Value {
	res := form.children[1].Eval(c)
	if res.Type() == "error" {
		wrapped := Atom{t: "string", value: res.Value().(string)}
		if len(form.children) > 2 {
			handler := form.children[2].Eval(c)
			if handler.Type() == "function" {
				args := &List{children: []Value{wrapped}}
				return handler.(Function).Call(args, c)
			} else if handler.Type() == "macro" {
				args := &List{children: []Value{NIL, wrapped}}
				return handler.(Macro).Call(args, c)
			} else {
				return Atom{t: "error", value: fmt.Sprintf("guard expected argument 1 of type 'function', got type '%s'", handler.Type())}
			}
		}
		return NIL
	} else {
		return res
	}
}

func condForm(form *List, c Context) Value {
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

func init() {
	specialForms = map[string]Form{
		"lambda": lambdaForm,
		"def":    defForm,
		"do":     doForm,
		"if":     ifForm,
		"while":  whileForm,
		"for":    forForm,
		"let":    letForm,
		"assert": assertForm,
		"input":  inputForm,
		"macro":  macroForm,
		//"debug":     debugForm,
		"set!":      setBangForm,
		"namespace": namespaceForm,
		"import":    importForm,
		"guard":     guardForm,
		"cond":      condForm,
	}
}

// TODO: (cond (bool op) (bool op) ...)
