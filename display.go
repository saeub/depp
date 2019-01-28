package main

import (
	"sort"

	sent "./sent"
	termbox "github.com/nsf/termbox-go"
)

type display struct {
	tokenSpacing     int
	offsetX, offsetY int
	drawables        []drawable
}

type drawable interface {
	draw(offsetX, offsetY int)
	occupiesRect(x, y, width, height int) bool
	zIndex() int
}

type label struct {
	x, y   int
	text   string
	fg, bg termbox.Attribute
}

type bracket struct {
	xStart, yStart int
	xEnd, yEnd     int
	yMiddle        int
	below          bool
	style          bracketStyle
	fg, bg         termbox.Attribute
}

type bracketStyle struct {
	legStart    rune
	legEnd      rune
	middle      rune
	cornerStart []rune
	cornerEnd   []rune
}

var (
	bracketAbove = bracketStyle{
		legStart:    '║',
		legEnd:      '│',
		middle:      '─',
		cornerStart: []rune{'╓', '╖'},
		cornerEnd:   []rune{'┌', '┐'},
	}
	bracketBelow = bracketStyle{
		legStart:    '║',
		legEnd:      '│',
		middle:      '─',
		cornerStart: []rune{'╙', '╜'},
		cornerEnd:   []rune{'└', '┘'},
	}
)

func (l label) draw(offsetX, offsetY int) {
	for i, r := range []rune(l.text) {
		termbox.SetCell(offsetX+l.x+i, offsetY+l.y, r, l.fg, l.bg)
	}
}

func (l label) occupiesRect(x, y, width, height int) bool {
	lLen := len([]rune(l.text))
	return overlap(l.x, l.x+lLen-1, x, x+width-1) && overlap(l.y, l.y, y, y+height-1)
}

func (l label) zIndex() int {
	return 2
}

func (b bracket) draw(offsetX, offsetY int) {
	var dirY int
	if b.below {
		dirY = 1
	} else {
		dirY = -1
	}
	var dirX int
	if b.xStart < b.xEnd {
		dirX = 1
	} else {
		dirX = -1
	}

	// draw legs
	for y := b.yStart; y != b.yMiddle; y += dirY {
		termbox.SetCell(offsetX+b.xStart, offsetY+y, b.style.legStart, termbox.ColorDefault, termbox.ColorDefault)
	}
	for y := b.yEnd; y != b.yMiddle; y += dirY {
		termbox.SetCell(offsetX+b.xEnd, offsetY+y, b.style.legEnd, termbox.ColorDefault, termbox.ColorDefault)
	}

	// draw middle part
	for x := b.xStart; x != b.xEnd; x += dirX {
		termbox.SetCell(offsetX+x, offsetY+b.yMiddle, b.style.middle, termbox.ColorDefault, termbox.ColorDefault)
	}

	// draw corners
	var start int
	if b.xStart < b.xEnd {
		start = 0
	} else {
		start = 1
	}
	termbox.SetCell(offsetX+b.xStart, offsetY+b.yMiddle, b.style.cornerStart[start], termbox.ColorDefault, termbox.ColorDefault)
	termbox.SetCell(offsetX+b.xEnd, offsetY+b.yMiddle, b.style.cornerEnd[start^1], termbox.ColorDefault, termbox.ColorDefault)
}

func (b bracket) occupiesRect(x, y, width, height int) bool {
	return overlap(b.xStart, b.xEnd, x, x+width-1) && overlap(b.yMiddle, b.yMiddle, y, y+height-1)
}

func (b bracket) zIndex() int {
	return 1
}

func newDisplay(tokenSpacing int) *display {
	return &display{
		tokenSpacing: tokenSpacing,
		offsetX:      0,
		offsetY:      0,
		drawables:    []drawable{},
	}
}

func (d *display) addDrawable(item drawable) {
	d.drawables = append(d.drawables, item)
}

func (d *display) putSentence(sent sent.Sentence) {
	// clear drawables
	d.drawables = []drawable{}

	// draw tokens
	toks := sent.Tokens()
	tokPos := make([]int, len(toks))
	x := 0
	for i, tok := range toks {
		tokLen := len([]rune(tok.Text))
		idLen := len([]rune(tok.ID))
		totalLen := max(tokLen, idLen)
		idOffset := (totalLen - idLen) / 2

		d.addDrawable(label{
			x:    x,
			y:    -1,
			text: tok.Text,
			fg:   termbox.ColorDefault,
			bg:   termbox.ColorDefault,
		})
		d.addDrawable(label{
			x:    x + idOffset,
			y:    0,
			text: tok.ID,
			fg:   termbox.ColorDefault,
			bg:   termbox.ColorDefault,
		})

		tokPos[i] = x + (totalLen-1)/2
		x += totalLen + d.tokenSpacing
	}

	// draw dependencies
	for _, dep := range sent.DependenciesAbove() {
		disp.putDependency(dep, tokPos, -2, false)
	}
	for _, dep := range sent.DependenciesBelow() {
		disp.putDependency(dep, tokPos, 1, true)
	}

	d.sortDrawables()
}

// TODO less redundancy
func (d *display) putDependency(dep sent.Dependency, tokPos []int, yStart int, below bool) {
	xStart := tokPos[dep.HeadIndex]
	xEnd := tokPos[dep.DependentIndex]
	yMiddle := 0
	var dirY int
	if below {
		dirY = 1
	} else {
		dirY = -1
	}

yLoop:
	for y := yStart; ; y += dirY {
		for _, drwbl := range d.drawables {
			if drwbl.occupiesRect(xStart, y, xEnd-xStart+1, 1) {
				// not a free spot
				continue yLoop
			}
		}
		// free spot found
		yMiddle = y
		break
	}

	d.addDrawable(bracket{
		xStart:  xStart,
		yStart:  yStart,
		xEnd:    xEnd,
		yEnd:    yStart,
		yMiddle: yMiddle,
		below:   below,
		style:   bracketBelow,
		fg:      termbox.ColorDefault,
		bg:      termbox.ColorDefault,
	})
	lLen := len([]rune(dep.Name))
	d.addDrawable(label{
		x:    (xStart + xEnd - lLen + 1) / 2,
		y:    yMiddle,
		text: dep.Name,
		fg:   termbox.ColorBlack,
		bg:   termbox.ColorWhite,
	})
}

func (d *display) sortDrawables() {
	sort.Slice(d.drawables, func(i, j int) bool {
		return d.drawables[i].zIndex() < d.drawables[j].zIndex()
	})
}

func (d *display) scroll(x, y int) {
	d.offsetX -= x
	d.offsetY -= y
}

func (d *display) resetScroll() {
	_, height := termbox.Size()
	d.offsetX = 0
	d.offsetY = height / 2
}

func (d *display) render() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for _, drwbl := range d.drawables {
		drwbl.draw(d.offsetX, d.offsetY)
	}
	termbox.Flush()
}
