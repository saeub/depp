package sent

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type proconSentence struct {
	conll  *conllSentence
	procon []Dependency
}

func ReadProconSentence(reader *bufio.Reader) (Sentence, error) {
	relRegex, _ := regexp.Compile(`^([pc])(\d+),(\d+)$`)
	effRegex, _ := regexp.Compile(`^([pn](?:eff|ac))(\d+)$`)

	c, err := ReadConllSentence(reader)
	conll, _ := c.(*conllSentence)
	if conll == nil {
		// end of file, no more sentences to return
		return nil, nil
	}
	procon := []Dependency{}
	if err != nil {
		// non-CoNLL line
		line := err.Error()
		log.Println(line)
		for len(line) != 0 {
			rel := relRegex.FindStringSubmatch(line)
			eff := effRegex.FindStringSubmatch(line)
			if len(rel) > 0 {
				headID, _ := strconv.Atoi(rel[2])
				dependentID, _ := strconv.Atoi(rel[3])
				procon = append(procon, Dependency{
					Name:           rel[1],
					HeadIndex:      headID - 1,
					DependentIndex: dependentID - 1,
				})
			} else if len(eff) > 0 {
				headID, _ := strconv.Atoi(eff[2])
				procon = append(procon, Dependency{
					Name:           eff[1],
					HeadIndex:      headID - 1,
					DependentIndex: headID - 1,
				})
			} else {
				// invalid line
				return &proconSentence{conll, procon}, errors.New(line)
			}
			l, err := reader.ReadString('\n')
			line = strings.TrimSpace(l)
			log.Println(line, err)
			if err != nil {
				// end of file
				break
			}
		}
	}
	return &proconSentence{conll, procon}, nil
}

func (sent *proconSentence) Tokens() []Token {
	return sent.conll.Tokens()
}

func (sent *proconSentence) PrimaryDependencies() []Dependency {
	return sent.procon
}

func (sent *proconSentence) SecondaryDependencies() []Dependency {
	return sent.conll.PrimaryDependencies()
}

func (sent *proconSentence) AddDependency(name, headID, depID string) error {
	switch name {
	case "p", "c":
		headID, _ := strconv.Atoi(headID)
		dependentID, _ := strconv.Atoi(depID)
		sent.procon = append(sent.procon, Dependency{
			Name:           name,
			HeadIndex:      headID - 1,
			DependentIndex: dependentID - 1,
		})
	case "peff", "neff", "pac", "nac":
		headID, _ := strconv.Atoi(headID)
		sent.procon = append(sent.procon, Dependency{
			Name:           name,
			HeadIndex:      headID - 1,
			DependentIndex: headID - 1,
		})
	}
	return nil
}

func (sent *proconSentence) RemoveDependency(dep *Dependency) error {
	for i, d := range sent.procon {
		if &d == dep {
			sent.procon = append(sent.procon[:i], sent.procon[i+1:]...)
		}
	}
	return nil
}

func (sent *proconSentence) Output(writer io.Writer) {
	fmt.Fprint(writer, "not inplemented")
}
