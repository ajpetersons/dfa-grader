package dfa

import (
	"fmt"
	"sync"
)

var maxDepth int
var progress []*sync.WaitGroup

func getEditCount(m1, m2 *DFA, depth int, solution chan<- int) {
	// FIXME: terrible performance. for positive scenarios with small number of
	// necessary edits we could improve by doing non-recursive function with
	// linear pace relative to current depth
	defer progress[depth].Done()

	var err error
	if depth > maxDepth {
		return
	}

	// check if m1 == m2
	eq, err := compare(m1, m2)
	if err != nil {
		fmt.Println("Could not minimize automata, aborting calculation")
		return
	}
	if eq {
		fmt.Println(depth)
		return
	}

	// try different start states
	for _, s := range m1.States() {
		m1Copy := m1.Copy()
		m1Copy.SetStartState(s)
		progress[depth+1].Add(1)
		go getEditCount(m1Copy, m2, depth+1, solution)
	}

	// try swapping one state final/non-final
	for _, s := range m1.States() {
		m1Copy := m1.Copy()
		finalStates := m1Copy.FinalStates()
		var wasFinal bool
		for idx, f := range finalStates {
			if f == s {
				finalStates = append(finalStates[:idx], finalStates[idx+1:]...)
				wasFinal = true
			}
		}
		if !wasFinal {
			finalStates = append(finalStates, s)
		}
		m1Copy.SetFinalStates(finalStates...)
		progress[depth+1].Add(1)
		go getEditCount(m1Copy, m2, depth+1, solution)
	}

	// try adding a new state
	mCopied := m1.Copy()
	s := mCopied.GetNewState()
	for _, l := range mCopied.Alphabet() {
		mCopied.SetTransition(s, l, s)
	}
	progress[depth+1].Add(1)
	go getEditCount(mCopied, m2, depth+1, solution)

	// try switching transition
	for _, from := range m1.States() {
		for _, to := range m1.States() {
			for _, l := range m1.Alphabet() {
				// greedy
				m1Copy := m1.Copy()
				m1Copy.SetTransition(from, l, to)
				progress[depth+1].Add(1)
				go getEditCount(m1Copy, m2, depth+1, solution)
			}
		}
	}
}

// GetDFASyntaxDifference calculates score by measuring amount of edits
// necessary to transform one dfa into the other
// m2 is automata that is expected to be received
// function returns number of edits that was necessary
func GetDFASyntaxDifference(m1, m2 *DFA) int {
	maxDepthCalculated := len(m2.States()) * len(m2.Alphabet())
	maxDepthCalculated = 2

	maxDepth = maxDepthCalculated
	solutions := make(chan int)
	defer close(solutions)

	m2Min := m2.Copy()
	err := m2Min.Determinize()
	if err != nil {
		return maxDepthCalculated
	}
	m2Min.Minimize()

	// +2 to make sure exit condition can be satisfied
	for i := 0; i < maxDepthCalculated+2; i++ {
		progress = append(progress, &sync.WaitGroup{})
	}

	progress[0].Add(1)
	go getEditCount(m1, m2Min, 0, solutions)

	go func() {
		for {
			res, more := <-solutions
			if !more {
				return
			}
			if res < maxDepth {
				maxDepth = res
			}
		}
	}()

	for i := 0; i <= maxDepthCalculated; i++ {
		progress[i].Wait()
		fmt.Printf("Done all edits of size %d\n", i)
	}

	return maxDepth
}
