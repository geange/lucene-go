package main

import (
	"fmt"
	"github.com/geange/lucene-go/core"
)

func main() {

	automaton := core.NewAutomaton()

	state := automaton.CreateState()

	automaton.SetAccept(state, true)

	fmt.Println(automaton.IsAccept(state))
}
