package main

import (
	"flag"
	"fmt"
	"github.com/bobappleyard/readline"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	str "strings"
)

const (
	NORMAL = iota
	COMMENT
	STRING
)

var identRegexp = regexp.MustCompile(`^[\w\$!\+\-=<>\*\/](?:(?:\/|-)[\d\w]|[\d\w_\$])*[\?!\*]?(\.\.\.)?$`)
var symbolRegexp = regexp.MustCompile(`^:\w(?:(?:-)[\d\w]|[\d\w_\$])*$`)

func IsIdentifier(token string) bool {
	return identRegexp.MatchString(token)
}

func IsSymbol(token string) bool {
	return symbolRegexp.MatchString(token)
}

func Categorize(input string) LispValue {
	if str.HasSuffix(input, "...") {
		if IsIdentifier(input[:len(input)-3]) {
			return Atom{value: input[:len(input)-3], t: "expansion"}
		} else {
			return Atom{t: "error", value: fmt.Sprintf("Invalid token '%s'", input)}
		}
	} else if str.HasPrefix(input, "#") {
		if IsIdentifier(input[1:]) {
			return Atom{value: input[1:], t: "type"}
		} else {
			return Atom{t: "error", value: fmt.Sprintf("Invalid token '%s'", input)}
		}
	} else if i := str.Index(input, "#"); i > -1 {
		a := input[:i]
		t := input[i+1:]
		if IsIdentifier(a) && IsIdentifier(t) {
			return Atom{value: input, t: "typed id"}
		} else {
			return Atom{t: "error", value: fmt.Sprintf("Invalid token '%s'", input)}
		}
	}
	if input == "true" {
		return TRUE
	} else if input == "false" {
		return FALSE
	} else if input == "nil" {
		return NIL
	}
	i, err := strconv.ParseInt(input, 10, 64)
	if err == nil {
		return Atom{value: int(i), t: "int"}
	}
	f, err2 := strconv.ParseFloat(input, 64)
	if err2 == nil {
		return Atom{value: f, t: "float"}
	}
	if input[0] == '"' && input[len(input)-1] == '"' {
		return Atom{value: input[1 : len(input)-1], t: "string"}
	} else if IsIdentifier(input) {
		return Atom{value: input, t: "identifier"}
	} else if IsSymbol(input) {
		return Atom{value: input, t: "symbol"}
	} else {
		return Atom{t: "error", value: fmt.Sprintf("Invalid token '%s'", input)}
	}
}

func Tokenize(input string) []string {
	tokens := []string{}
	tok := []rune{}
	mode := NORMAL
	for _, c := range input {
		add_token := false
		add_char := false
		if mode == COMMENT {
			if c == '\n' {
				mode = NORMAL
			}
		} else if mode == STRING {
			if c == '"' && len(tok) > 0 && (tok[len(tok)-1] != '\\') {
				mode = NORMAL
				add_token = true
			} else if len(tok) > 0 && tok[len(tok)-1] == '\\' {
				switch c {
				case 'n':
					tok = tok[:len(tok)-1]
					c = '\n'
				case 't':
					tok = tok[:len(tok)-1]
					c = '\t'
				case 'r':
					tok = tok[:len(tok)-1]
					c = '\r'
				}
			}
			add_char = true
		} else {
			switch c {
			case ';':
				mode = COMMENT
			case '"':
				mode = STRING
				add_char = true
			case ' ':
				add_token = true
			case '\n':
				add_token = true
			case '\t':
				add_token = true
			case '(':
				if len(tok) > 0 && tok[len(tok)-1] != '\'' {
					tokens = append(tokens, string(tok))
					tok = []rune{}
				}
				add_char = true
				add_token = true
			case ')':
				if len(tok) > 0 {
					tokens = append(tokens, string(tok))
					tok = []rune{}
				}
				add_char = true
				add_token = true
			case '[':
				continue
			case ']':
				continue
			case '{':
				if len(tok) > 0 {
					tokens = append(tokens, string(tok))
					tok = []rune{}
				}
				add_char = true
				add_token = true
			case '}':
				if len(tok) > 0 {
					tokens = append(tokens, string(tok))
					tok = []rune{}
				}
				add_char = true
				add_token = true
			default:
				add_char = true
			}
		}
		if add_char {
			tok = append(tok, c)
		}
		if add_token && len(tok) > 0 {
			tokens = append(tokens, string(tok))
			tok = []rune{}
		}
	}
	if len(tok) > 0 {
		tokens = append(tokens, string(tok))
	}
	return tokens
}

func Parse(tokens []string) []LispValue {
	var output []LispValue
	var stack []LispContainer
	for _, token := range tokens {
		if token == "'(" {
			stack = append(stack, &List{Quoted: true})
		} else if token == "(" {
			stack = append(stack, &List{})
		} else if token == ")" {
			s := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			n := len(stack)
			if n > 0 {
				stack[n-1].Append(s)
			} else {
				output = append(output, s)
			}
		} else if token == "{" {
			stack = append(stack, &Hash{pairs: []LispValue{},
				sym_vals: make(map[string]LispValue),
				vals:     make(map[string]LispValue)})
		} else if token == "}" {
			if stack[len(stack)-1].Type() != "hash" {
				log.Fatal("Unexpected token '}' (no matching open bracket).")
			}
			s := stack[len(stack)-1].(*Hash)
			stack = stack[:len(stack)-1]
			n := len(stack)
			if n > 0 {
				stack[n-1].Append(s)
			} else {
				output = append(output, s)
			}
		} else {
			s := Categorize(token)
			n := len(stack)
			if n > 0 {
				stack[n-1].Append(s)
			} else {
				return []LispValue{s}
			}
		}
	}
	return output
}

func REPL(c *Context) *Context {
	tokens := []string{}
	main_prompt := "> "
	incomplete := ". "
	leftCount := 0
	rightCount := 0

	for {
		prompt := main_prompt
		if len(tokens) > 0 {
			prompt = incomplete
		}
		line, err := readline.String(prompt)
		if err != nil {
			fmt.Println("error: ", err)
			break
		}
		if line != "" {
			line = str.TrimRight(line, "\r\n ")
			readline.AddHistory(line)
			if line == "quit" {
				fmt.Println("bai :3")
				break
			}
			temp := Tokenize(line)
			for _, t := range temp {
				if t == "(" || t == "'(" {
					leftCount += 1
				} else if t == ")" {
					rightCount += 1
				}
			}
			tokens = append(tokens, temp...)

			if leftCount == rightCount {
				leftCount = 0
				rightCount = 0
				nodes := Parse(tokens)
				for _, n := range nodes {
					r := n.Eval(c)
					if r.Type() == "error" {
						fmt.Println("error:", r.Value().(string))
						break
					}
					if r != NIL {
						fmt.Println(r)
					}
				}
				tokens = []string{}
			}
		}
	}
	return c
}

func LoadFile(filename string, c *Context) LispValue {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	// need to figure out how to handle parse problems
	nodes := Parse(Tokenize(string(dat)))
	var last LispValue = NIL
	for _, n := range nodes {
		last = n.Eval(c)
		if last.Type() == "error" {
			log.Fatal(last.Value().(string))
		}
	}
	return last
}

func main() {
	command := flag.String("c", "", "Execute a command")
	interactive := flag.Bool("i", false, "Enter the repl, after processing other commands")
	flag.Parse()
	filename := flag.Arg(0)

	c := NewContext(nil)
	Setup(c)
	Bootstrap(c)
	if !(*command == "") {
		nodes := Parse(Tokenize(*command))
		for _, n := range nodes {
			r := n.Eval(c)
			if r != NIL {
				fmt.Println(r)
			}
		}
		if *interactive {
			REPL(c)
		}
	} else if filename != "" {
		LoadFile(filename, c)
		if *interactive {
			REPL(c)
		}
	} else {
		REPL(c)
	}

}
