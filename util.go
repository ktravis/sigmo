package sigmo

import (
	"fmt"
	"log"
)

func Add(a Atom, b Atom) Atom {
	if (a.t == "int" || a.t == "float") && (b.t == "int" || b.t == "float") {
		sum := a.AsFloat() + b.AsFloat()
		if a.t == "int" && b.t == "int" {
			return Atom{t: "int", value: int(sum)}
		}
		return Atom{t: "float", value: sum}
	}
	return Atom{t: "error", value: "Non-numeric value being added"}
}

func Negate(a Atom) Atom {
	if a.t == "float" {
		return Atom{t: "float", value: -a.value.(float64)}
	} else if a.t == "int" {
		return Atom{t: "int", value: -a.value.(int)}
	}
	return Atom{t: "error", value: "Non-numeric value cannot be negated"}
}

func Multiply(a Atom, b Atom) Atom {
	if (a.t == "int" || a.t == "float") && (b.t == "int" || b.t == "float") {
		prod := a.AsFloat() * b.AsFloat()
		if a.t == "int" && b.t == "int" {
			return Atom{t: "int", value: int(prod)}
		}
		return Atom{t: "float", value: prod}
	}
	return Atom{t: "error", value: "Non-numeric value being multiplied"}
}

func Divide(a Atom, b Atom) Atom {
	if (a.t == "int" || a.t == "float") && (b.t == "int" || b.t == "float") {
		div := a.AsFloat() / b.AsFloat()
		if a.t == "int" && b.t == "int" {
			return Atom{t: "int", value: int(div)}
		}
		return Atom{t: "float", value: div}
	}
	return Atom{t: "error", value: "Non-numeric value being divided"}
}

func Compare(a Value, b Value) bool {
	if a.Type() != b.Type() {
		return false
	}
	if a.Type() == "nil" {
		return a == b
	}
	return a.Value() == b.Value()
}

func (n Atom) AsFloat() float64 {
	if n.t == "float" {
		return n.value.(float64)
	} else if n.t == "int" {
		return float64(n.value.(int))
	}
	log.Fatal("Non-numeric type cannot be converted to float.")
	return -1
}

func CompareNum(a Atom, b Atom) int {
	an := a.AsFloat()
	bn := b.AsFloat()
	if an == bn {
		return 0
	} else if an > bn {
		return 1
	} else {
		return -1
	}
}

func Boolean(n Value) bool {
	if n.Type() == "list" {
		return len(n.(*List).children) > 0
	}
	if n.Type() == "hash" {
		h := n.(*Hash)
		return len(h.vals)+len(h.sym_vals) > 0
	}
	a := n.(Atom)
	switch a.t {
	case "string":
		return len(a.value.(string)) > 0
	case "int":
		return a.value.(int) != 0
	case "float":
		return a.value.(float64) != 0.0
	case "bool":
		return a == TRUE
	case "function":
		return true
	default:
		return false
	}
}

func NestedReplace(n Value, subs *map[string]Value) []Value {
	switch n.Type() {
	case "identifier":
		if v, ok := (*subs)[n.Value().(string)]; ok {
			return []Value{v}
		}
	case "expansion":
		if v, ok := (*subs)[n.Value().(string)]; ok {
			if v.Type() != "list" {
				log.Fatal(fmt.Sprintf("Cannot expand value of type '%s'", v.Type()))
			}
			out := []Value{}
			out = append(out, v.(*List).children...)
			return out
		}
	case "list":
		o := n.(*List)
		l := &List{Quoted: o.Quoted}
		for _, c := range o.children {
			l.children = append(l.children, NestedReplace(c, subs)...)
		}
		return []Value{l}
	}
	return []Value{n}
}
