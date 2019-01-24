package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

func sentencesFromFile(filename string, readFunc func(*bufio.Reader) sentence) (sents []sentence) {
	f, _ := os.Open(filename)
	r := bufio.NewReader(f)
	for sent := readFunc(r); sent != nil; sent = readFunc(r) {
		sents = append(sents, sent)
	}
	return sents
}

type conllSentence struct {
	rows [][]string
}

func readConllSentence(reader *bufio.Reader) sentence {
	rows := make([][]string, 0)
	for true {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			if len(rows) == 0 {
				if err != nil {
					// end of file
					break
				}
				// ignore leading whitespace
				continue
			}
			// end of sentence
			break
		}
		row := strings.Split(line, "\t")
		if len(row) < 10 {
			log.Panicln("invalid line:", line)
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		// end of file, no more sentences to return
		return nil
	}
	return &conllSentence{rows}
}

func (sent *conllSentence) tokens() []token {
	toks := make([]token, len(sent.rows))
	// TODO check for inconsistent IDs?
	for i, f := range sent.rows {
		toks[i] = token{f[0], f[1], f[2]}
	}
	return toks
}

func (sent *conllSentence) dependenciesAbove() []dependency {
	return nil
}

func (sent *conllSentence) dependenciesBelow() []dependency {
	deps := make([]dependency, len(sent.rows))
	for i, f := range sent.rows {
		headID, _ := strconv.Atoi(f[6])
		headIndex := headID - 1 // convert CoNLL ID to slice index
		if headIndex < 0 || headIndex >= len(sent.rows) {
			// invalid ID; set current token as head
			headIndex = i
		}
		dependentID, _ := strconv.Atoi(f[0])
		dependentIndex := dependentID - 1 // convert CoNLL ID to slice index
		if dependentIndex < 0 || dependentIndex > len(sent.rows) {
			// invalid ID; set current token as dependent
			dependentIndex = i
		}
		deps[i] = dependency{f[7], headIndex, dependentIndex}
	}
	// display shorter dependencies closer to the sentence
	sort.Slice(deps, func(i, j int) bool {
		return abs(deps[i].dependentIndex-deps[i].headIndex) < abs(deps[j].dependentIndex-deps[j].headIndex)
	})
	return deps
}

func (sent *conllSentence) output(writer io.Writer) {
	for _, f := range sent.rows {
		fmt.Fprint(writer, strings.Join(f, "\t"))
	}
}
