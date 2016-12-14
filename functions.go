package sigmo

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

var builtins = map[string]Function{
	"%":           NewFunction("%", "int,int", modFunction),
	"+":           NewFunction("+", "int|float,+", plusFunction),
	"-":           NewFunction("-", "int|float,+", minusFunction),
	"*":           NewFunction("*", "int|float,+", mulFunction),
	"/":           NewFunction("/", "int|float,int|float", divFunction),
	"print":       NewFunction("print", "**", printFunction),
	"println":     NewFunction("println", "**", printlnFunction),
	"cat":         NewFunction("cat", "string,+", catFunction),
	"head":        NewFunction("head", "list", headFunction),
	"tail":        NewFunction("tail", "list", tailFunction),
	"cons":        NewFunction("cons", "*,*", consFunction),
	"rev":         NewFunction("rev", "list|string", revFunction),
	"len":         NewFunction("len", "list|string|hash", lenFunction),
	"eq":          NewFunction("eq", "*,*", eqFunction),
	"neq":         NewFunction("neq", "*,*", neqFunction),
	"and":         NewFunction("and", "bool,+", andFunction),
	"or":          NewFunction("or", "bool,+", orFunction),
	"xor":         NewFunction("xor", "bool,+", xorFunction),
	"not":         NewFunction("not", "bool", notFunction),
	"lt":          NewFunction("lt", "int|float,int|float", ltFunction),
	"lte":         NewFunction("lte", "int|float,int|float", lteFunction),
	"gt":          NewFunction("gt", "int|float,int|float", gtFunction),
	"gte":         NewFunction("gte", "int|float,int|float", gteFunction),
	"exec":        NewFunction("exec", "list", execFunction),
	"eval":        NewFunction("eval", "string", evalFunction),
	"trim":        NewFunction("trim", "string,string", trimFunction),
	"join":        NewFunction("join", "list,string", joinFunction),
	"split":       NewFunction("split", "string,string", splitFunction),
	"split-n":     NewFunction("split-n", "string,string,int", splitNFunction),
	"parse-int":   NewFunction("parse-int", "string", parseIntFunction),
	"parse-float": NewFunction("parse-float", "string", parseFloatFunction),
	"get":         NewFunction("get", "list,int", getFunction),
	"hget":        NewFunction("hget", "hash,string|symbol", hgetFunction),
	"hset!":       NewFunction("hset!", "hash,string|symbol,*", hsetBangFunction),
	"hcontains":   NewFunction("hcontains", "hash,string|symbol", hcontainsFunction),
	"type":        NewFunction("type", "*", typeFunction),
	"int":         NewFunction("int", "*", intFunction),
	"float":       NewFunction("float", "*", floatFunction),
	"string":      NewFunction("string", "*", stringFunction),
	"bool":        NewFunction("bool", "*", boolFunction),
	"floor":       NewFunction("floor", "float", floorFunction),
	"ceil":        NewFunction("ceil", "float", ceilFunction),
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

func setBuiltins(c Context) {
	for name, fn := range builtins {
		c.Set(name, fn)
	}
	for name, a := range aliases {
		c.Set(name, c.Get(a))
	}
}

func execFunction(input *List, c Context) Value {
	x := input.children[0].Copy().(*List)
	x.Quoted = false
	return x.Eval(c)
}

func evalFunction(input *List, c Context) Value {
	var last Value = NIL
	x := input.children[0].Value().(string)
	nodes, err := Parse(Tokenize(x))
	if err != nil {
		return Atom{t: "error", value: err}
	}
	for _, n := range nodes {
		last = n.Eval(c)
		if last.Type() == "error" {
			return last
		}
	}
	return last
}

// math
func modFunction(input *List, c Context) Value {
	return Atom{t: "int", value: input.children[0].Value().(int64) % input.children[1].Value().(int64)}
}

func plusFunction(input *List, c Context) Value {
	sum := input.children[0].(Atom)
	for _, n := range input.children[1:] {
		sum = Add(sum, n.(Atom))
	}
	return sum
}

func minusFunction(input *List, c Context) Value {
	return Add(input.children[0].(Atom), Negate(input.children[1].(Atom)))
}

func mulFunction(input *List, c Context) Value {
	sum := input.children[0].(Atom)
	for _, n := range input.children[1:] {
		sum = Multiply(sum, n.(Atom))
	}
	return sum
}

func divFunction(input *List, c Context) Value {
	return Divide(input.children[0].(Atom), input.children[1].(Atom))
}

func floorFunction(input *List, c Context) Value {
	return Atom{t: "float", value: math.Floor(input.children[0].Value().(float64))}
}

func ceilFunction(input *List, c Context) Value {
	return Atom{t: "float", value: math.Ceil(input.children[0].Value().(float64))}
}

// i/o
func printFunction(input *List, c Context) Value {
	output := []string{}
	for _, n := range input.children {
		if n.Type() == "string" {
			output = append(output, n.Value().(string))
		} else {
			output = append(output, n.String())
		}
	}
	fmt.Printf(strings.Join(output, " "))
	return NIL
}

func printlnFunction(input *List, c Context) Value {
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
		fmt.Println(strings.Join(output, " "))
	}
	return NIL
}

// list
func headFunction(input *List, c Context) Value {
	return input.children[0].(*List).children[0]
}

func tailFunction(input *List, c Context) Value {
	return &List{children: input.children[0].(*List).children[1:]}
}

func consFunction(input *List, c Context) Value {
	var left, right *List
	if input.children[0].Type() != "list" {
		left = &List{children: []Value{input.children[0]}}
	} else {
		left = input.children[0].(*List)
	}
	if input.children[1].Type() != "list" {
		right = &List{children: []Value{input.children[1]}}
	} else {
		right = input.children[1].(*List)
	}
	left.children = append(left.children, right.children...)
	return left
}

func revFunction(input *List, c Context) Value {
	l := input.children[0].(*List)
	out := []Value{}
	if len(l.children) > 0 {
		for i := len(l.children) - 1; i >= 0; i -= 1 {
			out = append(out, l.children[i])
		}
	}
	return &List{children: out}
}

func lenFunction(input *List, c Context) Value {
	if input.children[0].Type() == "list" {
		return input.children[0].(*List).Length()
	} else if input.children[0].Type() == "hash" {
		return input.children[0].(*Hash).Length()
	} else {
		return input.children[0].(Atom).Length()
	}
}

func getFunction(input *List, c Context) Value {
	l := input.children[0].(*List)
	i := input.children[1].Value().(int)
	for i < 0 {
		i += len(l.children)
	}
	if i < len(l.children) {
		return l.children[i]
	}
	return Atom{t: "error", value: fmt.Sprintf("Index '%d' out of list bounds.", i)}
}

// logical
func andFunction(input *List, c Context) Value {
	for _, n := range input.children {
		if !Boolean(n) {
			return FALSE
		}
	}
	return TRUE
}

func orFunction(input *List, c Context) Value {
	for _, n := range input.children {
		if Boolean(n) {
			return TRUE
		}
	}
	return FALSE
}

func xorFunction(input *List, c Context) Value {
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
}

// comparison
func eqFunction(input *List, c Context) Value {
	if !Compare(input.children[0], input.children[1]) {
		return FALSE
	}
	return TRUE
}

func neqFunction(input *List, c Context) Value {
	if Compare(input.children[0], input.children[1]) {
		return FALSE
	}
	return TRUE
}

func notFunction(input *List, c Context) Value {
	return Atom{t: "bool", value: !Boolean(input.children[0])}
}

func ltFunction(input *List, c Context) Value {
	return Atom{t: "bool", value: CompareNum(input.children[0].(Atom), input.children[1].(Atom)) == -1}
}

func lteFunction(input *List, c Context) Value {
	return Atom{t: "bool", value: CompareNum(input.children[0].(Atom), input.children[1].(Atom)) <= 0}
}

func gtFunction(input *List, c Context) Value {
	return Atom{t: "bool", value: CompareNum(input.children[0].(Atom), input.children[1].(Atom)) == 1}
}

func gteFunction(input *List, c Context) Value {
	return Atom{t: "bool", value: CompareNum(input.children[0].(Atom), input.children[1].(Atom)) >= 0}
}

// strings
func catFunction(input *List, c Context) Value {
	text := ""
	for _, s := range input.children {
		text += strings.Trim(s.String(), "\"")
	}
	return Atom{t: "string", value: text}
}

func trimFunction(input *List, c Context) Value {
	return Atom{t: "string", value: strings.Trim(input.children[0].Value().(string), input.children[1].Value().(string))}
}

func joinFunction(input *List, c Context) Value {
	elms := []string{}
	for _, c := range input.children[0].(*List).children {
		if c.Type() == "string" {
			elms = append(elms, c.Value().(string))
		} else {
			elms = append(elms, c.String())
		}
	}
	return Atom{t: "string", value: strings.Join(elms, input.children[1].Value().(string))}
}

func splitFunction(input *List, c Context) Value {
	out := &List{}
	for _, s := range strings.Split(input.children[0].Value().(string), input.children[1].Value().(string)) {
		out.children = append(out.children, Atom{t: "string", value: s})
	}
	return out
}

func splitNFunction(input *List, c Context) Value {
	out := &List{}
	for _, s := range strings.SplitN(input.children[0].Value().(string), input.children[1].Value().(string), input.children[2].Value().(int)) {
		out.children = append(out.children, Atom{t: "string", value: s})
	}
	return out
}

func parseIntFunction(input *List, c Context) Value {
	s := input.children[0].Value().(string)
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return Atom{t: "error", value: fmt.Sprintf("Could not convert string '%s' to an integer.", s)}
	}
	return Atom{t: "int", value: int(i)}
}

func parseFloatFunction(input *List, c Context) Value {
	s := input.children[0].Value().(string)
	f, err := strconv.ParseFloat(s, 10)
	if err != nil {
		return Atom{t: "error", value: fmt.Sprintf("Could not convert string '%s' to a float.", s)}
	}
	return Atom{t: "float", value: f}
}

// hash
func hgetFunction(input *List, c Context) Value {
	h := input.children[0].(*Hash)
	s := input.children[1].Value().(string)
	if input.children[1].Type() == "string" {
		return h.vals[s]
	} else {
		return h.sym_vals[s]
	}
}

func hsetBangFunction(input *List, c Context) Value {
	h := input.children[0].(*Hash)
	s := input.children[1].Value().(string)
	v := input.children[2]
	if input.children[1].Type() == "string" {
		h.vals[s] = v
	} else {
		h.sym_vals[s] = v
	}
	return h
}

func hcontainsFunction(input *List, c Context) Value {
	h := input.children[0].(*Hash)
	s := input.children[1].Value().(string)
	if input.children[1].Type() == "string" {
		_, ok := h.vals[s]
		return Atom{t: "bool", value: ok}
	} else {
		_, ok := h.sym_vals[s]
		return Atom{t: "bool", value: ok}
	}
}

// type
func typeFunction(input *List, c Context) Value {
	return Atom{t: "type", value: input.children[0].Type()}
}

func intFunction(input *List, c Context) Value {
	if input.children[0].Type() == "int" {
		return input.children[0]
	}
	return Atom{t: "int", value: int(input.children[0].Value().(float64))}
}

func floatFunction(input *List, c Context) Value {
	if input.children[0].Type() == "float" {
		return input.children[0]
	}
	return Atom{t: "float", value: float64(input.children[0].Value().(int))}
}

func stringFunction(input *List, c Context) Value {
	if input.children[0].Type() == "string" {
		return input.children[0]
	}
	return Atom{t: "string", value: input.children[0].String()}
}

func boolFunction(input *List, c Context) Value {
	return Atom{t: "bool", value: Boolean(input.children[0])}
}
