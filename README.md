sigmo
=====

Golang LISP dialect

```bash
go build
./lsp                   # cli
./lsp test.lsp          # run a file
./lsp -c '(print "hi")' # run a single command
./lsp -i test.lsp       # run a file, drop into cli with context
```

An interpreted lisp, written in Go.

All of the basic lisp features you would expected, and a few extra:

- namespaces `(namespace test ...)`
- importing files `(import core/math)`
- `for` loop construct 
- errors and "guards" (think try/except) `(guard (error "help"))`
- real macros
- hashmap values `{ "a" 1 }`
- value expansions `(mylist...)`
- type hints for functions `(defn onlyints (a#int) (println 'a was an int'))`

See examples for more!
