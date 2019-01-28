package main

import (
	"fmt"
	"log"
	"os"

	termbox "github.com/nsf/termbox-go"

	sent "./sent"
)

var (
	loadedSents []sent.Sentence
	disp        *display
	dispSentID  int
	cmd         *command
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %v FILE\n", os.Args[0])
		os.Exit(1)
	}

	var err error
	loadedSents, err = sent.SentencesFromFile(os.Args[1], sent.ReadConllSentence)
	if err != nil {
		log.Fatalln(err)
	}

	if len(loadedSents) == 0 {
		log.Fatalln("no sentences or wrong format")
	}

	err = mainloop()
	if err != nil {
		log.Fatalln(err)
	}
}

func mainloop() error {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)

	disp = newDisplay(3)
	disp.putSentence(loadedSents[0])
	disp.resetScroll()

	var exitErr error
	for exitErr == nil {
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		disp.render()
		if cmd != nil {
			cmd.render()
		}
		termbox.Flush()
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyCtrlC {
				exitErr = fmt.Errorf("keyboard interrupt")
			} else {
				handleKeyPress(ev.Key, ev.Ch)
			}
		case termbox.EventMouse:
			handleClick(ev.Key, ev.MouseX, ev.MouseY)
		}
	}
	return exitErr
}

func handleKeyPress(key termbox.Key, ch rune) {
	if cmd != nil {
		cmd.handleKeyPress(key, ch)
	} else {
		if key != 0 {
			switch key {

			}
		} else if ch != 0 {
			switch ch {
			case 'h':
				disp.scroll(-5, 0)
			case 'j':
				disp.scroll(0, 1)
			case 'k':
				disp.scroll(0, -1)
			case 'l':
				disp.scroll(5, 0)
			case 'p':
				navigateSentence(-1)
			case 'n':
				navigateSentence(1)
			case 'a':
				cmd = newCommand("add: ", `^([^\d\s]+)(\d+)(?:,(\d+))?$`, func(match []string) {
					if match != nil {
						loadedSents[dispSentID].AddDependency(match[1], match[2], match[3])
						disp.putSentence(loadedSents[dispSentID])
					} else {
						// TODO handle invalid command
					}
					cmd = nil
				})
			}
		}
	}
}

func handleClick(key termbox.Key, x, y int) {
	if key == termbox.MouseLeft {
		disp.selectAt(x, y)
	}
}

func navigateSentence(offset int) {
	dispSentID += offset
	if dispSentID < 0 {
		dispSentID = 0
	}
	if dispSentID >= len(loadedSents) {
		dispSentID = len(loadedSents) - 1
	}
	disp.putSentence(loadedSents[dispSentID])
	disp.resetScroll()
}
