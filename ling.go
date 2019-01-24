package main

import "io"

type token struct {
	id    string
	text  string
	lemma string
}

type dependency struct {
	name           string
	headIndex      int
	dependentIndex int
}

type sentence interface {
	tokens() []token
	dependenciesAbove() []dependency
	dependenciesBelow() []dependency
	output(io.Writer)
}
