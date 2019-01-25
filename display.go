package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	sent "./sent"
	"github.com/jroimartin/gocui"
)

type display struct {
	buffer         []string
	tokenSeparator string
	resetScroll    bool // reset scroll offset on next layout()
}

func newDisplay(g *gocui.Gui, viewname string, tokenSep string) (d *display) {
	if err := g.SetKeybinding(viewname, gocui.KeyArrowLeft, gocui.ModNone,
		func(_ *gocui.Gui, v *gocui.View) error {
			scrollDisplay(d, v, -5, 0)
			return nil
		}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(viewname, 'h', gocui.ModNone,
		func(_ *gocui.Gui, v *gocui.View) error {
			scrollDisplay(d, v, -5, 0)
			return nil
		}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(viewname, gocui.KeyArrowRight, gocui.ModNone,
		func(_ *gocui.Gui, v *gocui.View) error {
			scrollDisplay(d, v, 5, 0)
			return nil
		}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(viewname, 'l', gocui.ModNone,
		func(_ *gocui.Gui, v *gocui.View) error {
			scrollDisplay(d, v, 5, 0)
			return nil
		}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(viewname, gocui.KeyArrowUp, gocui.ModNone,
		func(_ *gocui.Gui, v *gocui.View) error {
			scrollDisplay(d, v, 0, -1)
			return nil
		}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(viewname, 'k', gocui.ModNone,
		func(_ *gocui.Gui, v *gocui.View) error {
			scrollDisplay(d, v, 0, -1)
			return nil
		}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(viewname, gocui.KeyArrowDown, gocui.ModNone,
		func(_ *gocui.Gui, v *gocui.View) error {
			scrollDisplay(d, v, 0, 1)
			return nil
		}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(viewname, 'j', gocui.ModNone,
		func(_ *gocui.Gui, v *gocui.View) error {
			scrollDisplay(d, v, 0, 1)
			return nil
		}); err != nil {
		log.Panicln(err)
	}
	return &display{
		tokenSeparator: tokenSep,
	}
}

func (d *display) drawMessage(msg string) {
	d.buffer = []string{msg}
}

func (d *display) drawSentence(sent sent.Sentence) {
	d.buffer = make([]string, 0)
	var sentLine strings.Builder
	var idLine strings.Builder
	toks := sent.Tokens()
	idPositions := make([]int, len(toks))
	for i, t := range toks {
		tokenLen := len([]rune(t.Text))
		idString := strconv.Itoa(i)
		idLen := len(idString)
		spacing := max(tokenLen, idLen)
		idOffset := (spacing - idLen) / 2

		sentLine.WriteString(t.Text)
		sentLine.WriteString(strings.Repeat(" ", spacing-tokenLen))
		idLine.WriteString(strings.Repeat(" ", idOffset))
		idPositions[i] = len([]rune(idLine.String()))
		idLine.WriteString(idString)
		idLine.WriteString(strings.Repeat(" ", spacing-idLen-idOffset))

		sentLine.WriteString(d.tokenSeparator)
		idLine.WriteString(d.tokenSeparator)
	}
	depbufAbove := drawDependencies(sent.DependenciesAbove(), idPositions, true)

	for i := len(depbufAbove) - 1; i >= 0; i-- {
		d.buffer = append(d.buffer, depbufAbove[i])
	}
	d.buffer = append(d.buffer, sentLine.String(), idLine.String())
	depbufBelow := drawDependencies(sent.DependenciesBelow(), idPositions, false)
	for i := 0; i < len(depbufBelow); i++ {
		d.buffer = append(d.buffer, depbufBelow[i])
	}
	d.resetScroll = true
}

func drawDependencies(deps []sent.Dependency, idPos []int, above bool) []string {
	buf := make([][]rune, 0)
	for _, d := range deps {
		head := idPos[d.HeadIndex]
		dependent := idPos[d.DependentIndex]
		left := min(head, dependent)
		right := max(head, dependent)
		var leftMark, rightMark rune
		if d.HeadIndex < d.DependentIndex {
			// head left of dependent
			if above {
				leftMark = '┌'
				rightMark = '╖'
			} else {
				leftMark = '└'
				rightMark = '╜'
			}
		} else if d.HeadIndex > d.DependentIndex {
			// head right of dependent
			if above {
				leftMark = '╓'
				rightMark = '┐'
			} else {
				leftMark = '╙'
				rightMark = '┘'
			}
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
		labelRunes := []rune(d.Name)
		labelLen := len(labelRunes)
		labelOffset := (right - left - labelLen) / 2
		labelY := deplineY
		if head == dependent {
			// dodge buffer edge
			left = max(left, -labelOffset)
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
	strbuf := make([]string, len(buf))
	for i, l := range buf {
		strbuf[i] = string(l)
	}
	return strbuf
}

func scrollDisplay(d *display, v *gocui.View, x int, y int) {
	viewWidth, viewHeight := v.Size()
	bufWidth := longestLen(d.buffer)
	bufHeight := len(d.buffer)
	oldX, oldY := v.Origin()
	maxX := max(0, bufWidth-viewWidth-3)
	maxY := max(0, bufHeight-viewHeight)
	v.SetOrigin(limit(oldX+x, 0, maxX), limit(oldY+y, 0, maxY))
}

func unscrollDisplay(d *display, v *gocui.View) {
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
		unscrollDisplay(d, v)
		d.resetScroll = false
	}
	return nil
}
