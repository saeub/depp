package sent

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type conllSentence struct {
	rows [][]string
}

func ReadConllSentence(reader *bufio.Reader) (Sentence, error) {
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
			// invalid line
			return &conllSentence{rows}, errors.New(line)
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		// end of file, no more sentences to return
		return nil, nil
	}
	return &conllSentence{rows}, nil
}

func (sent *conllSentence) Tokens() []Token {
	toks := make([]Token, len(sent.rows))
	// TODO check for inconsistent IDs?
	for i, f := range sent.rows {
		toks[i] = Token{f[0], f[1], f[2]}
	}
	return toks
}

func (sent *conllSentence) PrimaryDependencies() []Dependency {
	deps := make([]Dependency, len(sent.rows))
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
		deps[i] = Dependency{f[7], headIndex, dependentIndex}
	}
	// display shorter dependencies closer to the sentence
	sortDependencies(deps)
	return deps
}

func (sent *conllSentence) SecondaryDependencies() []Dependency {
	return nil
}

func (sent *conllSentence) AddDependency(name, headID, depID string) error {
	// TODO check for invalid head ID
	if depID == "" {
		// root dependency
		depID = headID
		headID = "0"
	}
	for _, r := range sent.rows {
		if r[0] == depID {
			r[7] = name
			r[6] = headID
			return nil
		}
	}
	return fmt.Errorf("no token with ID %s", depID)
}

func (sent *conllSentence) RemoveDependency(dep *Dependency) error {
	depID := strconv.Itoa(dep.DependentIndex + 1)
	for _, r := range sent.rows {
		if r[0] == depID {
			r[7] = ""
			r[6] = "0"
			return nil
		}
	}
	return fmt.Errorf("no token with ID %s", depID)
}

func (sent *conllSentence) Output(writer io.Writer) {
	for _, f := range sent.rows {
		fmt.Fprint(writer, strings.Join(f, "\t"))
	}
}
