package main

import (
	"log"
	"regexp"

	termbox "github.com/nsf/termbox-go"
)

type command struct {
	prompt   string
	regex    *regexp.Regexp
	callback func(match []string)
	input    []rune
}

type commandStatus int

const (
	cmdInputting commandStatus = iota
	cmdEntered
	cmdCancelled
)

func newCommand(prompt, expression string, callback func(match []string)) *command {
	regex, err := regexp.Compile(expression)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &command{
		prompt:   prompt,
		regex:    regex,
		callback: callback,
		input:    []rune{},
	}
}

func (c *command) match() []string {
	return c.regex.FindStringSubmatch(string(c.input))
}

func (c *command) handleKeyPress(key termbox.Key, ch rune) {
	log.Println(key, ch)
	if key != 0 {
		switch key {
		case termbox.KeyEnter:
			c.callback(c.regex.FindStringSubmatch(string(c.input)))
		case termbox.KeyEsc:
			c.callback(nil)
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
	log.Println("render")
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
