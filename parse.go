package sigmo

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type ReadMode int

const (
	ReadNormal ReadMode = iota
	ReadComment
	ReadString
)

var identRegexp = regexp.MustCompile(`^[\w\$!\+\-=<>\*\/](?:(?:\/|-)[\d\w]|[\d\w_\$])*[\?!\*]?(\.\.\.)?$`)
var symbolRegexp = regexp.MustCompile(`^:\w(?:(?:-)[\d\w]|[\d\w_\$])*$`)

func isIdentifier(token string) bool {
	return identRegexp.MatchString(token)
}

func isSymbol(token string) bool {
	return symbolRegexp.MatchString(token)
}

func categorize(input string) (Value, error) {
	if strings.HasSuffix(input, "...") {
		if isIdentifier(input[:len(input)-3]) {
			return Atom{
				t:     "expansion",
				value: input[:len(input)-3],
			}, nil
		} else {
			return nil, fmt.Errorf("Invalid token '%s'", input)
		}
	} else if strings.HasPrefix(input, "#") {
		if isIdentifier(input[1:]) {
			return Atom{
				t:     "type",
				value: input[1:],
			}, nil
		} else {
			return nil, fmt.Errorf("Invalid token '%s'", input)
		}
	} else if i := strings.Index(input, "#"); i > -1 {
		a := input[:i]
		t := input[i+1:]

		if isIdentifier(a) && isIdentifier(t) {
			return Atom{
				t:     "typed id",
				value: input,
			}, nil
		} else {
			return nil, fmt.Errorf("Invalid token '%s'", input)
		}
	}

	if input == "true" {
		return TRUE, nil
	} else if input == "false" {
		return FALSE, nil
	} else if input == "nil" {
		return NIL, nil
	}

	if i, err := strconv.ParseInt(input, 10, 64); err == nil {
		return Atom{
			t:     "int",
			value: int(i),
		}, nil
	}

	if f, err2 := strconv.ParseFloat(input, 64); err2 == nil {
		return Atom{
			t:     "float",
			value: f,
		}, nil
	}

	if input[0] == '"' && input[len(input)-1] == '"' {
		return Atom{
			t:     "string",
			value: input[1 : len(input)-1],
		}, nil
	} else if isIdentifier(input) {
		return Atom{
			t:     "identifier",
			value: input,
		}, nil
	} else if isSymbol(input) {
		return Atom{
			t:     "symbol",
			value: input,
		}, nil
	}

	return nil, fmt.Errorf("Invalid token '%s'", input)
}

func Tokenize(input string) []string {
	tokens := []string{}
	tok := []rune{}
	mode := ReadNormal

	for _, c := range input {
		switch mode {
		case ReadComment:
			if c == '\n' {
				mode = ReadNormal
			}
		case ReadString:
			if c == '"' && len(tok) > 0 && (tok[len(tok)-1] != '\\') {
				mode = ReadNormal
				if len(tok) > 0 {
					tokens = append(tokens, string(tok))
					tok = []rune{}
				}
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
			tok = append(tok, c)
		case ReadNormal:
			switch c {
			// don't add char, do add token
			case '\n':
			case '\t':
			case ' ':
				break
			case ';':
				mode = ReadComment
				continue
			case '"':
				mode = ReadString
				tok = append(tok, c)
				continue
			case '(':
				if len(tok) > 0 && tok[len(tok)-1] != '\'' {
					tokens = append(tokens, string(tok))
					tok = []rune{}
				}
				tok = append(tok, c)
			case ')':
				if len(tok) > 0 {
					tokens = append(tokens, string(tok))
					tok = []rune{}
				}
				tok = append(tok, c)
			case '[':
				continue
			case ']':
				continue
			case '{':
				if len(tok) > 0 {
					tokens = append(tokens, string(tok))
					tok = []rune{}
				}
				tok = append(tok, c)
			case '}':
				if len(tok) > 0 {
					tokens = append(tokens, string(tok))
					tok = []rune{}
				}
				tok = append(tok, c)
			default:
				tok = append(tok, c)
				continue
			}
			if len(tok) > 0 {
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

func Parse(tokens []string) ([]Value, error) {
	var output []Value
	var stack []Container

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
			stack = append(stack, &Hash{
				pairs:    []Value{},
				sym_vals: make(map[string]Value),
				vals:     make(map[string]Value),
			})
		} else if token == "}" {
			if stack[len(stack)-1].Type() != "hash" {
				return nil, fmt.Errorf("Unexpected token '}' (no matching open bracket).")
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
			s, err := categorize(token)
			if err != nil {
				return nil, err
			}

			n := len(stack)
			if n > 0 {
				stack[n-1].Append(s)
			} else {
				return []Value{s}, nil
			}
		}
	}
	return output, nil
}
