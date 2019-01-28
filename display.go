package main

import (
	"sort"

	sent "./sent"
	termbox "github.com/nsf/termbox-go"
)

type display struct {
	tokenSpacing     int
	offsetX, offsetY int
	drawables        []*drawable
	selectedDrawable *drawable
}

type drawable interface {
	draw(offsetX, offsetY int, selected bool)
	occupiesRect(x, y, width, height int) bool
	zIndex() int
	selectable() bool
	Data() interface{}
}

type label struct {
	x, y   int
	text   string
	fg, bg termbox.Attribute
	data   interface{}
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

func (l label) draw(offsetX, offsetY int, selected bool) {
	fg := l.fg
	bg := l.bg
	if selected {
		fg |= termbox.AttrBold
		bg |= termbox.AttrReverse
	}
	for i, r := range []rune(l.text) {
		termbox.SetCell(offsetX+l.x+i, offsetY+l.y, r, fg, bg)
	}
}

func (l label) occupiesRect(x, y, width, height int) bool {
	lLen := len([]rune(l.text))
	return overlap(l.x, l.x+lLen-1, x, x+width-1) && overlap(l.y, l.y, y, y+height-1)
}

func (l label) zIndex() int {
	return 2
}

func (l label) selectable() bool {
	return true
}

func (l label) Data() interface{} {
	return l.data
}

func (b bracket) draw(offsetX, offsetY int, selected bool) {
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

	fg := b.fg
	bg := b.bg
	if selected {
		fg |= termbox.AttrBold
		bg |= termbox.AttrReverse
	}

	// draw legs
	for y := b.yStart; y != b.yMiddle; y += dirY {
		termbox.SetCell(offsetX+b.xStart, offsetY+y, b.style.legStart, fg, bg)
	}
	for y := b.yEnd; y != b.yMiddle; y += dirY {
		termbox.SetCell(offsetX+b.xEnd, offsetY+y, b.style.legEnd, fg, bg)
	}

	// draw middle part
	for x := b.xStart; x != b.xEnd; x += dirX {
		termbox.SetCell(offsetX+x, offsetY+b.yMiddle, b.style.middle, fg, bg)
	}

	// draw corners
	var start int
	if b.xStart < b.xEnd {
		start = 0
	} else {
		start = 1
	}
	termbox.SetCell(offsetX+b.xStart, offsetY+b.yMiddle, b.style.cornerStart[start], fg, bg)
	termbox.SetCell(offsetX+b.xEnd, offsetY+b.yMiddle, b.style.cornerEnd[start^1], fg, bg)
}

func (b bracket) occupiesRect(x, y, width, height int) bool {
	return overlap(b.xStart, b.xEnd, x, x+width-1) && overlap(b.yMiddle, b.yMiddle, y, y+height-1)
}

func (b bracket) zIndex() int {
	return 1
}

func (b bracket) selectable() bool {
	return false
}

func (b bracket) Data() interface{} {
	return nil
}

func newDisplay(tokenSpacing int) *display {
	return &display{
		tokenSpacing: tokenSpacing,
		offsetX:      0,
		offsetY:      0,
		drawables:    []*drawable{},
	}
}

func (d *display) addDrawable(item drawable) {
	d.drawables = append(d.drawables, &item)
}

func (d *display) removeDrawable(item *drawable) {
	index := -1
	for i, drwbl := range d.drawables {
		if drwbl == item {
			index = i
			break
		}
	}
	d.drawables = append(d.drawables[:index], d.drawables[index+1:]...)
}

func (d *display) selectAt(x, y int) {
	drwbls := []*drawable{}
	for _, drwbl := range d.drawables {
		if (*drwbl).selectable() && (*drwbl).occupiesRect(x-d.offsetX-1, y-d.offsetY-1, 1, 1) {
			drwbls = append(drwbls, drwbl)
		}
	}
	if len(drwbls) > 0 {
		d.selectedDrawable = drwbls[0]
	} else {
		d.selectedDrawable = nil
	}
}

func (d *display) putSentence(sent sent.Sentence) {
	// clear drawables
	d.drawables = []*drawable{}

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
			data: tok,
		})
		d.addDrawable(label{
			x:    x + idOffset,
			y:    0,
			text: tok.ID,
			fg:   termbox.ColorDefault,
			bg:   termbox.ColorDefault,
			data: tok,
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
			if (*drwbl).occupiesRect(xStart, y, xEnd-xStart+1, 1) {
				// not a free spot
				continue yLoop
			}
		}
		// free spot found
		yMiddle = y
		break
	}

	d.addDrawable(&bracket{
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
	d.addDrawable(&label{
		x:    (xStart + xEnd - lLen + 1) / 2,
		y:    yMiddle,
		text: dep.Name,
		fg:   termbox.ColorDefault,
		bg:   termbox.ColorDefault,
		data: dep,
	})
}

func (d *display) sortDrawables() {
	sort.Slice(d.drawables, func(i, j int) bool {
		return (*d.drawables[i]).zIndex() < (*d.drawables[j]).zIndex()
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
	for _, drwbl := range d.drawables {
		(*drwbl).draw(d.offsetX, d.offsetY, drwbl == d.selectedDrawable)
	}
}
