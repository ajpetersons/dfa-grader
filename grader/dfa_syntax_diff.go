package grader

import (
	"dfa-grader/config"
	"dfa-grader/dfa"
	"fmt"
	"strings"
	"sync"
	"time"
)

type dfaSyntaxSolver struct {
	progress  []*sync.WaitGroup
	mu        *sync.Mutex
	solution  *int
	timeouted chan struct{}
}

func newDFASyntaxSolver(depth int, m *dfa.DFA) *dfaSyntaxSolver {
	wgs := []*sync.WaitGroup{}
	for i := 0; i <= depth+1; i++ {
		wgs = append(wgs, &sync.WaitGroup{})
	}
	worst := new(int)
	*worst = len(m.States()) * len(m.Alphabet())
	if *worst <= config.DFADiff.MaxDepth {
		*worst = config.DFADiff.MaxDepth + 10
	}
	return &dfaSyntaxSolver{
		mu:        &sync.Mutex{},
		progress:  wgs,
		solution:  worst,
		timeouted: make(chan struct{}),
	}
}

func (solver *dfaSyntaxSolver) checkEq(
	m1, m2 *dfa.DFA,
	depth int,
) bool {
	// check if m1 == m2
	eq, err := dfa.Compare(m1, m2)
	if err != nil {
		fmt.Println("Could not minimize automata, aborting calculation")
		return false
	}
	if eq {
		solver.mu.Lock()
		if *solver.solution > depth {
			*solver.solution = depth
		}
		solver.mu.Unlock()
	}
	return eq
}

type domainElement struct {
	l dfa.Letter
	s dfa.State
}

func (solver *dfaSyntaxSolver) getEditCount(
	m1, m2 *dfa.DFA,
	depth int,
	state, start, final, transition bool,
	lastEdit interface{},
) {
	defer solver.progress[depth].Done()
	select {
	case <-solver.timeouted:
		return
	default:
	}

	if depth > config.DFADiff.MaxDepth || depth > *solver.solution {
		return
	}

	// check if m1 == m2
	if solver.checkEq(m1, m2, depth) {
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
		solver.progress[depth+1].Add(1)
		go solver.getEditCount(
			m1Copy, m2,
			depth+1,
			true, true, true, true,
			lastEdit,
		)
	}

	// try different start states
	if start {
		for _, s := range m1.States() {
			if strings.Compare(string(lastEdit.(dfa.State)), string(s)) == 1 {
				continue
			}

			m1Copy := m1.Copy()
			m1Copy.SetStartState(s)
			solver.progress[depth+1].Add(1)
			go solver.getEditCount(
				m1Copy, m2,
				depth+1,
				false, true, true, true,
				s,
			)
		}

		lastEdit = dfa.State("")
	}

	// try swapping one state final/non-final
	if final {
		for _, s := range m1.States() {
			if strings.Compare(string(lastEdit.(dfa.State)), string(s)) == 1 {
				continue
			}

			m1Copy := m1.Copy()
			finalStates := m1Copy.FinalStates()
			var wasFinal bool
			for idx, f := range finalStates {
				// need loop if we want to remove from final states
				if f == s {
					finalStates = append(
						finalStates[:idx],
						finalStates[idx+1:]...,
					)
					wasFinal = true
					break
				}
			}
			if !wasFinal {
				finalStates = append(finalStates, s)
			}
			m1Copy.SetFinalStates(finalStates...)
			solver.progress[depth+1].Add(1)
			go solver.getEditCount(
				m1Copy, m2,
				depth+1,
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
					solver.progress[depth+1].Add(1)
					go solver.getEditCount(
						m1Copy, m2,
						depth+1,
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
// function returns result in scale from 0 to 1
func GetDFASyntaxDifference(m1, m2 *dfa.DFA) float64 {
	solver := newDFASyntaxSolver(config.DFADiff.MaxDepth, m1)

	noResultScore := *solver.solution

	go func() {
		time.Sleep(config.DFADiff.Timeout)
		fmt.Println("Syntax diff: Timeouted")
		close(solver.timeouted)
	}()

	m2Min := m2.Copy()
	err := m2Min.Determinize()
	if err != nil {
		return 0.0
	}
	m2Min.Minimize()

	// add new state
	// assume that syntax mistakes do not exceed single missing state
	s := m1.GetNewState()
	for _, l := range m1.Alphabet() {
		m1.SetTransition(s, l, s)
	}

	solver.progress[0].Add(1)
	go solver.getEditCount(
		m1, m2Min,
		0,
		true, true, true, true,
		dfa.State(""),
	)

	haveResult := make(chan struct{}, 1)
	go func() {
		for i := 0; i < config.DFADiff.MaxDepth+1; i++ {
			solver.progress[i].Wait()
			fmt.Printf("Syntax Diff: Done all edits of size %d\n", i)
		}
		haveResult <- struct{}{}
	}()

	// wait until result or timeout
	select {
	case <-solver.timeouted:
	case <-haveResult:
	}

	solver.mu.Lock()
	defer solver.mu.Unlock()
	result := 1 - float64(*solver.solution)/float64(
		len(m2Min.States())*len(m2Min.Alphabet()),
	)
	if result < 0.0 || *solver.solution == noResultScore {
		result = 0.0
	}

	return result
}
