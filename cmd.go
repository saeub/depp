package main

import (
	"regexp"

	termbox "github.com/nsf/termbox-go"
)

type command struct {
	prompt   string
	regex    *regexp.Regexp
	callback func(input string, match []string)
	input    []rune
}

type commandStatus int

const (
	cmdInputting commandStatus = iota
	cmdEntered
	cmdCancelled
)

func newCommand(prompt, expression string, callback func(input string, match []string)) (*command, error) {
	regex, err := regexp.Compile(expression)
	if err != nil {
		return nil, err
	}
	return &command{
		prompt:   prompt,
		regex:    regex,
		callback: callback,
		input:    []rune{},
	}, nil
}

func (c *command) match() []string {
	return c.regex.FindStringSubmatch(string(c.input))
}

func (c *command) handleKeyPress(key termbox.Key, ch rune) {
	if key != 0 {
		switch key {
		case termbox.KeyEnter:
			inputStr := string(c.input)
			c.callback(inputStr, c.regex.FindStringSubmatch(inputStr))
		case termbox.KeyEsc:
			c.callback(string(c.input), nil)
		case termbox.KeyBackspace, termbox.KeyBackspace2:
			if len(c.input) > 0 {
				c.input = c.input[:len(c.input)-1]
			}
		}
	} else if ch != 0 {
		c.input = append(c.input, ch)
	}
}

func (c *command) render() {
	tWidth, tHeight := termbox.Size()
	y := tHeight - 1
	prompt := []rune(c.prompt)
	input := []rune(c.input)
	for x := 0; x < tWidth; x++ {
		if x < len(prompt) {
			termbox.SetCell(x, y, prompt[x], termbox.ColorDefault, termbox.ColorDefault)
		} else if x < len(prompt)+len(input) {
			termbox.SetCell(x, y, input[x-len(prompt)], termbox.ColorDefault, termbox.ColorDefault)
		} else {
			termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}
