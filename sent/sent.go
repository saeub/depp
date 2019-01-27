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
	DependenciesAbove() []Dependency
	DependenciesBelow() []Dependency
	AddDependency(string, string, string) error
	Output(io.Writer)
}

func SentencesFromFile(filename string, readFunc func(*bufio.Reader) Sentence) (sents []Sentence) {
	f, _ := os.Open(filename)
	defer f.Close()
	r := bufio.NewReader(f)
	for sent := readFunc(r); sent != nil; sent = readFunc(r) {
		sents = append(sents, sent)
	}
	return sents
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
