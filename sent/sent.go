package sent

import (
	"bufio"
	"io"
	"os"
	"sort"
)

type Token struct {
	ID    string
	Text  string
	Lemma string
}

type Dependency struct {
	Name           string
	HeadIndex      int
	DependentIndex int
}

type Sentence interface {
	Tokens() []Token
	PrimaryDependencies() []Dependency
	SecondaryDependencies() []Dependency
	AddDependency(name, headID, depID string) error
	RemoveDependency(dep *Dependency) error
	Output(io.Writer)
}

func SentencesFromFile(filename string, readFunc func(*bufio.Reader) (Sentence, error)) (sents []Sentence, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		sent, err := readFunc(r)
		if sent != nil {
			sents = append(sents, sent)
		} else if err != nil {
			// error while parsing sentence
			return sents, err
		} else {
			// end of file
			return sents, nil
		}
	}
}

func sortDependencies(deps []Dependency) {
	sort.Slice(deps, func(i, j int) bool {
		dist1 := deps[i].DependentIndex - deps[i].HeadIndex
		if dist1 < 0 {
			dist1 = -dist1
		}
		dist2 := deps[j].DependentIndex - deps[j].HeadIndex
		if dist2 < 0 {
			dist2 = -dist2
		}
		if dist1 < dist2 {
			// primary sort by distance
			return true
		} else if dist1 == dist2 {
			// secondary sort by dependent index
			return deps[i].DependentIndex < deps[j].DependentIndex
		}
		return false
	})
}
