package dfa

import (
	"dfa-grader/config"
	"fmt"
	"strings"
	"sync"
)

var solutionMu = &sync.Mutex{}
var progress []*sync.WaitGroup // FIXME: cant use globals, they will interfere between simultaneous tasks

func checkEq(m1, m2 *DFA, depth int, solution *int) bool {
	// check if m1 == m2
	eq, err := compare(m1, m2)
	if err != nil {
		fmt.Println("Could not minimize automata, aborting calculation")
		return false
	}
	if eq {
		solutionMu.Lock()
		if *solution > depth {
			*solution = depth
		}
		solutionMu.Unlock()
	}
	return eq
}

func getEditCount(
	m1, m2 *DFA,
	depth int, solution *int,
	state, start, final, transition bool,
	lastEdit interface{},
) {
	defer progress[depth].Done()

	if depth > config.DFADiff.MaxDepth || depth > *solution {
		return
	}

	// check if m1 == m2
	if checkEq(m1, m2, depth, solution) {
		return
	}

	if state {
		// add new state
		// assume that syntax mistakes do not exceed single missing state
		m1Copy := m1.Copy()
		s := m1Copy.GetNewState()
		for _, l := range m1Copy.Alphabet() {
			m1Copy.SetTransition(s, l, s)
		}
		progress[depth+1].Add(1)
		go getEditCount(
			m1Copy, m2,
			depth+1, solution,
			true, true, true, true,
			lastEdit,
		)
	}

	// try different start states
	if start {
		for _, s := range m1.States() {
			if strings.Compare(string(lastEdit.(State)), string(s)) == 1 {
				continue
			}

			m1Copy := m1.Copy()
			m1Copy.SetStartState(s)
			progress[depth+1].Add(1)
			go getEditCount(
				m1Copy, m2,
				depth+1, solution,
				false, true, true, true,
				s,
			)
		}

		lastEdit = State("")
	}

	// try swapping one state final/non-final
	if final {
		for _, s := range m1.States() {
			if strings.Compare(string(lastEdit.(State)), string(s)) == 1 {
				continue
			}

			m1Copy := m1.Copy()
			finalStates := m1Copy.FinalStates()
			var wasFinal bool
			for idx, f := range finalStates {
				// need loop if we want to remove from final states
				if f == s {
					finalStates = append(finalStates[:idx], finalStates[idx+1:]...)
					wasFinal = true
					break
				}
			}
			if !wasFinal {
				finalStates = append(finalStates, s)
			}
			m1Copy.SetFinalStates(finalStates...)
			progress[depth+1].Add(1)
			go getEditCount(
				m1Copy, m2,
				depth+1, solution,
				false, false, true, true,
				s,
			)
		}

		lastEdit = domainElement{s: "", l: ""}
	}

	// try switching transition
	if transition {
		for _, from := range m1.States() {
			for _, to := range m1.States() {
				for _, l := range m1.Alphabet() {
					// greedy
					de := lastEdit.(domainElement)
					if strings.Compare(string(de.s), string(from)) == 1 {
						continue
					}
					if strings.Compare(string(de.s), string(from)) == 0 &&
						strings.Compare(string(de.l), string(l)) == 1 {
						continue
					}
					m1Copy := m1.Copy()
					m1Copy.SetTransition(from, l, to)
					progress[depth+1].Add(1)
					go getEditCount(
						m1Copy, m2,
						depth+1, solution,
						false, false, false, true,
						domainElement{s: from, l: l},
					)
				}
			}
		}
	}
}

// GetDFASyntaxDifference calculates score by measuring amount of edits
// necessary to transform one dfa into the other
// m2 is automata that is expected to be received
// function returns number of edits that was necessary
func GetDFASyntaxDifference(m1, m2 *DFA) int {
	solution := new(int)
	*solution = len(m1.States()) * len(m1.Alphabet())

	if *solution < config.DFADiff.MaxDepth {
		*solution = config.DFADiff.MaxDepth + 10
	}

	m2Min := m2.Copy()
	err := m2Min.Determinize()
	if err != nil {
		return *solution
	}
	m2Min.Minimize()

	// add new state
	// assume that syntax mistakes do not exceed single missing state
	s := m1.GetNewState()
	for _, l := range m1.Alphabet() {
		m1.SetTransition(s, l, s)
	}

	// +2 to make sure exit condition can be satisfied
	for i := 0; i < config.DFADiff.MaxDepth+2; i++ {
		progress = append(progress, &sync.WaitGroup{})
	}

	progress[0].Add(1)
	go getEditCount(m1, m2Min, 0, solution, true, true, true, true, State(""))

	for i := 0; i < config.DFADiff.MaxDepth+1; i++ {
		progress[i].Wait()
		fmt.Printf("Done all edits of size %d\n", i)
	}

	return *solution
}
