package main

import (
	"flag"
	"fmt"
	"github.com/bobappleyard/readline"
	"io/ioutil"
	"strconv"
	str "strings"
)

const (
	NORMAL = iota
	COMMENT
	STRING
)

// need an IsIdentifier(string) bool

func Tokenize(input string) []string {
	tokens := []string{}
	tok := []rune{}
	mode := NORMAL
	for _, c := range input {
		if mode == COMMENT {
			if c == '\n' {
				mode = NORMAL
			}
		} else {
			add_token := false
			add_char := false
			switch c {
			case ';':
				mode = COMMENT
			case '"':
				if mode == STRING && (tok[len(tok)-1] != '\\') {
					mode = NORMAL
					add_token = true
				} else {
					mode = STRING
				}
				add_char = true
			case ' ':
				if mode == STRING {
					add_char = true
				} else {
					add_token = true
				}
			case '\n':
				if mode == STRING {
					add_char = true
				} else {
					add_token = true
				}
			case '\t':
				if mode == STRING {
					add_char = true
				} else {
					add_token = true
				}
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
				if mode != STRING {
					continue
				} else {
					add_char = true
				}
			case ']':
				if mode != STRING {
					continue
				} else {
					add_char = true
				}
			case '}':
				if mode != STRING {
					continue
				} else {
					add_char = true
				}
			case '{':
				if mode != STRING {
					continue
				} else {
					add_char = true
				}
			default:
				add_char = true
			}
			if add_char {
				tok = append(tok, c)
			}
			if add_token && len(tok) > 0 {
				tokens = append(tokens, string(tok))
				tok = []rune{}
			}
		}
	}
	if len(tok) > 0 {
		tokens = append(tokens, string(tok))
	}
	return tokens
}

func Parse(tokens []string) []LispValue {
	var output []LispValue
	var stack []List
	for _, token := range tokens {
		if token == "(" {
			stack = append(stack, List{})
		} else if token == ")" {
			s := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if len(stack) > 0 {
				stack[len(stack)-1].children = append(stack[len(stack)-1].children, s)
			} else {
				output = append(output, s)
			}
		} else {
			if len(stack) > 0 {
				stack[len(stack)-1].children = append(stack[len(stack)-1].children, Categorize(token))
			} else {
				return []LispValue{Categorize(token)}
			}
		}
	}
	return output
}

func Categorize(input string) LispValue {
	if input == "true" {
		return TRUE
	} else if input == "false" {
		return FALSE
	} else if input == "nil" {
		return NIL
	}
	i, err := strconv.ParseInt(input, 10, 64)
	if err == nil {
		return Atom{value: i, t: "int"}
	}
	f, err2 := strconv.ParseFloat(input, 64)
	if err2 == nil {
		return Atom{value: f, t: "float"}
	}
	if input[0] == '"' && input[len(input)-1] == '"' {
		return Atom{value: input[1 : len(input)-1], t: "string"}
	} else {
		return Atom{value: input, t: "identifier"}
	}
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
				if t == "(" {
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
				//fmt.Println(Print(root))
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
		dat, err := ioutil.ReadFile(filename)
		if err != nil {
			panic(err)
		}
		// need to figure out how to handle parse problems
		nodes := Parse(Tokenize(string(dat)))
		for _, n := range nodes {
			r := n.Eval(c)
			if r.Type() == "error" {
				panic(r.Value().(string))
			}
		}
		if *interactive {
			REPL(c)
		}
	} else {
		REPL(c)
	}

}
