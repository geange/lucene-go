package main

import (
	"fmt"
	"github.com/geange/lucene-go/core/util/automaton"
)

func main() {

	automaton := automaton.NewAutomaton()

	state := automaton.CreateState()

	automaton.SetAccept(state, true)

	fmt.Println(automaton.IsAccept(state))
}
