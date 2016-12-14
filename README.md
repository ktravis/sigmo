sigmo
=====

An interpreted lisp, written in Go.

All of the basic lisp features you would expected, and a few extra:

- namespaces `(namespace test ...)`
- importing files `(import core/math)`
- `for` loop construct 
- errors and "guards" (think try/except) `(guard (error "help"))`
- "real" macros
- hashmap values `{ "a" 1 }`
- value expansions `(mylist...)`
- type hints for functions `(defn onlyints (a#int) (println 'a was an int'))`

See [examples](./tree/master/examples/) for more!

install
#######

On linux: (requires package `libreadline-dev`)

```bash
go get github.com/ktravis/sigmo
```

build
#####

```bash
go build ./cmd/sigmo
```

run
###

```bash
./sigmo                   # cli
./sigmo test.mo           # run a file
./sigmo -c '(print "hi")' # run a single command
./sigmo -i test.mo        # run a file, drop into cli with context
```
