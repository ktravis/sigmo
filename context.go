package sigmo

import (
	"fmt"
	"strings"
)

type Context interface {
	Namespace(string) Context
	Get(string) Value
	Set(string, Value) Value
	SetExisting(string, Value) Value
	CopyLocals() map[string]Value
}

type context struct {
	parent     Context
	scope      map[string]Value
	ns         string
	namespaces map[string]Context
}

func NewContext(parent Context) *context {
	c := &context{
		parent:     parent,
		namespaces: make(map[string]Context),
		scope:      make(map[string]Value),
	}
	if parent == nil {
		setBuiltins(c)
	}
	return c
}

func (c *context) CopyLocals() map[string]Value {
	m := make(map[string]Value)
	for k, v := range c.scope {
		m[k] = v.Copy()
	}
	return m
}

func (c *context) Get(identifier string) Value {
	if strings.Contains(identifier, "/") {
		paths := strings.Split(identifier, "/")
		ns := strings.Join(paths[:len(paths)-1], "/")
		identifier = paths[len(paths)-1]
		if c.ns != ns {
			n := c.Namespace(ns)
			if n == nil {
				return Atom{t: "error", value: fmt.Sprintf("Unknown namespace '%s'", ns)}
			}
			return n.Get(identifier)
		}
	}
	a, ok := c.scope[identifier]
	if ok {
		return a
	}
	if c.parent != nil {
		return c.parent.Get(identifier)
	}
	return Atom{t: "error", value: fmt.Sprintf("Unknown identifier '%s'", identifier)}
}

func (c *context) Set(identifier string, l Value) Value {
	if strings.Contains(identifier, "/") {
		paths := strings.Split(identifier, "/")
		ns := strings.Join(paths[:len(paths)-1], "/")
		o := identifier
		identifier = paths[len(paths)-1]
		if c.ns != ns {
			if c.parent != nil {
				return c.parent.Set(o, l)
			}
			var x Context = c
			for _, p := range paths[:len(paths)-1] {
				x = x.Namespace(p)
			}
			return x.Set(identifier, l)
		}
	}
	c.scope[identifier] = l
	return l
}

func (c *context) SetExisting(identifier string, l Value) Value {
	// TODO: what about ns?
	if _, ok := c.scope[identifier]; ok {
		c.scope[identifier] = l
		return l
	}
	if c.parent != nil {
		return c.parent.SetExisting(identifier, l)
	}
	return Atom{t: "error", value: fmt.Sprintf("Unknown identifier '%s'", identifier)}
}

func (c *context) Namespace(path string) Context {
	if c.parent != nil {
		return c.parent.Namespace(path)
	}
	if strings.Contains(path, "/") {
		paths := strings.SplitN(path, "/", 2)
		x, exists := c.namespaces[paths[0]]
		if !exists {
			// TODO: err?
			return nil
		}
		return x.Namespace(paths[1])
	}
	x, exists := c.namespaces[path]
	if !exists {
		newCtx := NewContext(c)
		c.namespaces[path] = newCtx
		newCtx.ns = strings.Join([]string{c.ns, path}, "/")
		return newCtx
	}
	return x
}
