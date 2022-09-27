package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Attribute int

type Color struct {
	params  []Attribute
	noColor *bool
}

func (c *Color) Add(value ...Attribute) *Color {
	c.params = append(c.params, value...)
	return c
}

func (c *Color) SprintFunc() func(a ...interface{}) string {
	return func(a ...interface{}) string {
		return c.wrap(fmt.Sprint(a...))
	}
}

func (c *Color) SprintfFunc() func(format string, a ...interface{}) string {
	return func(format string, a ...interface{}) string {
		return c.wrap(fmt.Sprintf(format, a...))
	}
}

func (c *Color) wrap(s string) string {
	if c.isNoColorSet() {
		return s
	}

	return c.format() + s + c.unformat()
}

func (c *Color) sequence() string {
	format := make([]string, len(c.params))
	for i, v := range c.params {
		format[i] = strconv.Itoa(int(v))
	}

	return strings.Join(format, ";")
}


func (c *Color) format() string {
	return fmt.Sprintf("%s[%sm", escape, c.sequence())
}

func (c *Color) unformat() string {
	return fmt.Sprintf("%s[%dm", escape, Reset)
}

func (c *Color) isNoColorSet() bool {
	// check first if we have user set action
	if c.noColor != nil {
		return *c.noColor
	}

	// if not return the global option, which is disabled by default
	return NoColor
}

const escape = "\x1b"

const (
	Reset Attribute = iota
)

const (
	FgRed Attribute = iota + 31
	FgGreen
	FgYellow
	FgBlue
	FgMagenta
	FgCyan
)

var (
	colorList = []func(string, ...interface{}) string{
		BlueString,
		CyanString,
		GreenString,
		MagentaString,
		YellowString,
	}
	NoColor = noColorExists() || os.Getenv("TERM") == "dumb"
	colorCounter int
	colorsCache   = make(map[Attribute]*Color)
	colorsCacheMu sync.Mutex
)

func randColor(s string) string {
	colorCounter++
	if colorCounter == len(colorList) {
		colorCounter = 0
	}
	return colorList[colorCounter](s)
}
func errColor(s string) string {
	return RedString(s)
}

func getCachedColor(p Attribute) *Color {
	colorsCacheMu.Lock()
	defer colorsCacheMu.Unlock()

	c, ok := colorsCache[p]
	if !ok {
		c = New(p)
		colorsCache[p] = c
	}

	return c
}

func noColorExists() bool {
	_, exists := os.LookupEnv("NO_COLOR")
	return exists
}

func boolPtr(v bool) *bool {
	return &v
}

func New(value ...Attribute) *Color {
	c := &Color{
		params: make([]Attribute, 0),
	}

	if noColorExists() {
		c.noColor = boolPtr(true)
	}

	c.Add(value...)
	return c
}

func colorString(format string, p Attribute, a ...interface{}) string {
	c := getCachedColor(p)

	if len(a) == 0 {
		return c.SprintFunc()(format)
	}

	return c.SprintfFunc()(format, a...)
}

func RedString(format string, a ...interface{}) string { return colorString(format, FgRed, a...) }

func BlueString(format string, a ...interface{}) string { return colorString(format, FgBlue, a...) }

func CyanString(format string, a ...interface{}) string { return colorString(format, FgCyan, a...) }

func GreenString(format string, a ...interface{}) string { return colorString(format, FgGreen, a...) }

func MagentaString(format string, a ...interface{}) string { return colorString(format, FgMagenta, a...) }

func YellowString(format string, a ...interface{}) string { return colorString(format, FgYellow, a...) }
