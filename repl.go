package sigmo

import (
	"fmt"
	"strings"

	"github.com/bobappleyard/readline"
)

var (
	mainPrompt       = "> "
	incompletePrompt = ". "
)

func REPL(c Context) (Context, error) {
	tokens := []string{}
	leftCount := 0
	rightCount := 0

	for {
		prompt := mainPrompt
		if len(tokens) > 0 {
			prompt = incompletePrompt
		}

		line, err := readline.String(prompt)
		if err != nil {
			return nil, err
		}

		if line != "" {
			line = strings.TrimRight(line, "\r\n ")
			readline.AddHistory(line)

			if line == "quit" {
				return c, nil
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
				nodes, err := Parse(tokens)
				if err != nil {
					fmt.Println("error: ", err)
				} else {
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
				}

				leftCount = 0
				rightCount = 0
				tokens = []string{}
			}
		}
	}
}
