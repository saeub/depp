package main

import (
	"io/ioutil"
	"log"
	"regexp"

	"github.com/jroimartin/gocui"
)

func newCommand(g *gocui.Gui, cmdViewname, dispViewname, prompt, expression string, callback func([]string)) {
	regex, err := regexp.Compile(expression)
	if err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(cmdViewname, gocui.KeyEnter, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			cmd, _ := ioutil.ReadAll(v)
			match := regex.FindStringSubmatch(string(cmd))
			if match == nil {
				// TODO handle invalid command
			} else {
				v.Clear()
				v.SetCursor(0, 0)
				g.SetCurrentView(dispViewname)
				g.SetViewOnBottom(cmdViewname)
				callback(match)
			}
			return nil
		}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(cmdViewname, gocui.KeyEsc, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			v.Clear()
			v.SetCursor(0, 0)
			g.SetCurrentView(dispViewname)
			g.SetViewOnBottom(cmdViewname)
			callback(nil)
			return nil
		}); err != nil {
		log.Panicln(err)
	}

	v, _ := g.View(cmdViewname)
	v.Title = prompt
	g.SetCurrentView(cmdViewname)
	g.SetViewOnTop(cmdViewname)
}
