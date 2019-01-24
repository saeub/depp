package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/jroimartin/gocui"
)

type display struct {
	buffer         []string
	tokenSeparator string
	resetScroll    bool // reset scroll offset on next layout()
}

func newDisplay(g *gocui.Gui, viewname string, tokenSep string) display {
	if err := g.SetKeybinding(viewname, gocui.KeyArrowLeft, gocui.ModNone,
		func(_ *gocui.Gui, v *gocui.View) error {
			scrollView(v, -5, 0)
			return nil
		}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(viewname, 'h', gocui.ModNone,
		func(_ *gocui.Gui, v *gocui.View) error {
			scrollView(v, -5, 0)
			return nil
		}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(viewname, gocui.KeyArrowRight, gocui.ModNone,
		func(_ *gocui.Gui, v *gocui.View) error {
			scrollView(v, 5, 0)
			return nil
		}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(viewname, 'l', gocui.ModNone,
		func(_ *gocui.Gui, v *gocui.View) error {
			scrollView(v, 5, 0)
			return nil
		}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(viewname, gocui.KeyArrowUp, gocui.ModNone,
		func(_ *gocui.Gui, v *gocui.View) error {
			scrollView(v, 0, -1)
			return nil
		}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(viewname, 'k', gocui.ModNone,
		func(_ *gocui.Gui, v *gocui.View) error {
			scrollView(v, 0, -1)
			return nil
		}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(viewname, gocui.KeyArrowDown, gocui.ModNone,
		func(_ *gocui.Gui, v *gocui.View) error {
			scrollView(v, 0, 1)
			return nil
		}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(viewname, 'j', gocui.ModNone,
		func(_ *gocui.Gui, v *gocui.View) error {
			scrollView(v, 0, 1)
			return nil
		}); err != nil {
		log.Panicln(err)
	}
	return display{
		tokenSeparator: tokenSep,
	}
}

func (d *display) drawMessage(msg string) {
	d.buffer = []string{msg}
}

func (d *display) drawSentence(sent sentence) {
	d.buffer = make([]string, 0)
	var sentLine strings.Builder
	var idLine strings.Builder
	idPositions := make([]int, len(sent.tokens()))
	for i, t := range sent.tokens() {
		tokenLen := len([]rune(t.text))
		idString := strconv.Itoa(i)
		idLen := len(idString)
		spacing := max(tokenLen, idLen)
		idOffset := (spacing - idLen) / 2

		sentLine.WriteString(t.text)
		sentLine.WriteString(strings.Repeat(" ", spacing-tokenLen))
		idLine.WriteString(strings.Repeat(" ", idOffset))
		idPositions[i] = len([]rune(idLine.String()))
		idLine.WriteString(idString)
		idLine.WriteString(strings.Repeat(" ", spacing-idLen-idOffset))

		sentLine.WriteString(d.tokenSeparator)
		idLine.WriteString(d.tokenSeparator)
	}
	d.buffer = append(d.buffer, sentLine.String(), idLine.String())
	depbuf := drawDependencies(sent.dependenciesBelow(), idPositions, false)
	// TODO dependencies above
	for _, l := range depbuf {
		d.buffer = append(d.buffer, l)
	}
	d.resetScroll = true
}

func drawDependencies(deps []dependency, idPos []int, above bool) []string {
	buf := make([][]rune, 0)
	for _, d := range deps {
		head := idPos[d.headIndex]
		dependent := idPos[d.dependentIndex]
		left := min(head, dependent)
		right := max(head, dependent)
		var leftMark, rightMark rune
		if d.headIndex < d.dependentIndex {
			// head left of dependent
			leftMark = '└'
			rightMark = '╜'
		} else if d.headIndex > d.dependentIndex {
			// head right of dependent
			leftMark = '╙'
			rightMark = '┘'
		} else {
			// head == dependent
			leftMark = ' '
			rightMark = '║'
		}

		// find free space
		var deplineY int
		free := false
		for ly, l := range buf {
			for lx := left; lx <= right; lx++ {
				if l[lx] != ' ' {
					free = false
					break
				}
				free = true
			}
			if free {
				deplineY = ly
				break
			} else {
				if l[head] == ' ' {
					l[head] = '│'
				}
				l[dependent] = '║'
			}
		}

		// no free space :(
		if !free {
			// add empty line
			buf = append(buf, []rune(strings.Repeat(" ", idPos[len(idPos)-1]+1)))
			deplineY = len(buf) - 1
		}

		// draw horizontal line and label
		buf[deplineY][left] = leftMark
		buf[deplineY][right] = rightMark
		labelRunes := []rune(d.name)
		labelLen := len(labelRunes)
		labelOffset := (right - left - labelLen) / 2
		labelY := deplineY
		if head == dependent {
			// dodge buffer edge
			left = max(left, 0)
			left = min(left, len(buf[labelY])+labelOffset)
			for lx := left + labelOffset; lx < left+labelOffset+labelLen; lx++ {
				buf[labelY][lx] = labelRunes[lx-left-labelOffset]
			}
		} else {
			// left-align label if too long
			labelOffset = max(labelOffset, 1)
			for lx := left + 1; lx < right; lx++ {
				if lx >= left+labelOffset && lx < left+labelOffset+labelLen {
					buf[labelY][lx] = labelRunes[lx-left-labelOffset]
				} else {
					buf[labelY][lx] = '─'
				}
			}
		}
	}
	strbuf := make([]string, len(deps))
	for i, l := range buf {
		strbuf[i] = string(l)
	}
	return strbuf
}

func scrollView(v *gocui.View, x int, y int) {
	oldX, oldY := v.Origin()
	v.SetOrigin(oldX+x, oldY+y)
}

func resetScrollView(v *gocui.View) {
	v.SetOrigin(0, 0)
}

func (d *display) layout(v *gocui.View) error {
	v.Clear()
	if d.buffer == nil {
		fmt.Fprintln(v, "nothing to display")
		return nil
	}
	for _, l := range d.buffer {
		fmt.Fprintln(v, l)
	}
	if d.resetScroll {
		resetScrollView(v)
		d.resetScroll = false
	}
	return nil
}
