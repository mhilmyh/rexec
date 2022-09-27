package main

import "fmt"

type writer struct {
	prefix string
	pipe   chan string
}

func newWriter(prefix string) *writer {
	w := &writer{
		prefix: prefix,
		pipe:   make(chan string),
	}
	go w.run()
	return w
}

func (c *writer) run() {
	for {
		fmt.Print(<-c.pipe)
	}
}

func (c *writer) Write(b []byte) (int, error) {
	c.pipe <- c.prefix + string(b)
	return len(b), nil
}