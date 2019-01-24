package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jroimartin/gocui"
)

var (
	sents      []sentence
	disp       display
	dispSentID int
	tokenSep   string
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %v FILE\n", os.Args[0])
		os.Exit(1)
	}
	sents = sentencesFromFile(os.Args[1], readConllSentence)
	tokenSep = "   "

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("display", 'n', gocui.ModNone, func(_ *gocui.Gui, _ *gocui.View) error {
		navigateSentence(1)
		return nil
	}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("display", 'p', gocui.ModNone, func(_ *gocui.Gui, _ *gocui.View) error {
		navigateSentence(-1)
		return nil
	}); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	width, height := g.Size()

	v, err := g.SetView("display", 0, 0, width-1, height-1)
	if err != nil {
		if err == gocui.ErrUnknownView {
			disp = newDisplay(g, "display", tokenSep)
			disp.drawSentence(sents[dispSentID])
			g.SetCurrentView("display")
		} else {
			return err
		}
	}
	disp.layout(v)

	v, err = g.SetView("cmd", 0, height-3, width-1, height-1)
	if err != nil {
		if err == gocui.ErrUnknownView {
			// v.Frame = false
			// v.Editor = gocui.DefaultEditor
			v.Editable = true
			g.SetViewOnBottom("cmd")
		} else {
			return err
		}
	}
	return nil
}

func navigateSentence(offset int) {
	dispSentID += offset
	if dispSentID < 0 {
		dispSentID = 0
	}
	if dispSentID >= len(sents) {
		dispSentID = len(sents) - 1
	}
	disp.drawSentence(sents[dispSentID])
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
