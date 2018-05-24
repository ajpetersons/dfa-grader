package grader

import (
	"dfa-grader/config"
	"dfa-grader/dfa"
	"fmt"
	"time"
)

func getWordsUpToN(
	m *dfa.DFA,
	n int,
	returns chan<- map[string]bool,
	kill <-chan struct{},
) {
	prevStates := make(map[string]dfa.State)

	prevStates[""] = m.StartState()
	returns <- map[string]bool{
		"": m.IsFinal(m.StartState()),
	}

	for i := 1; i <= n; i++ {
		nextStates := make(map[string]dfa.State)
		words := make(map[string]bool)

		for word, state := range prevStates {
			for _, l := range m.Alphabet() {
				nextState, err := m.TransitionTarget(state, l)
				if err != nil {
					fmt.Printf("ERROR: %s\n", err.Error())
					continue
				}
				nextWord := word + string(l)
				// automata is deterministic, there is only one path to nextWord
				nextStates[nextWord] = nextState
				words[nextWord] = m.IsFinal(nextState)
			}
		}

		select {
		case <-kill:
			return
		case returns <- words:
		}

		prevStates = make(map[string]dfa.State)
		for k, v := range nextStates {
			prevStates[k] = v
		}
	}
}

// GetLanguageDifference calculates score given metric to check how many words
// differ for the languages
// Automata MUST be determinized
// m2 is automata that is expected to be received
func GetLanguageDifference(m1, m2 *dfa.DFA) float64 {
	n := config.LangDiff.MaxDepth - len(m2.Alphabet())
	if len(m2.Alphabet()) == 5 {
		// worst case
		n--
	}
	if n < config.LangDiff.MaxDepth {
		n = config.LangDiff.MaxDepth
	}

	kill := make(chan struct{})
	go func() {
		time.Sleep(config.LangDiff.Timeout)
		close(kill)
	}()

	words1 := make(chan map[string]bool)
	words2 := make(chan map[string]bool)

	go getWordsUpToN(m1, n, words1, kill)
	go getWordsUpToN(m2, n, words2, kill)

	nDiffs := make(chan float64)

	go func() {
		for i := 0; i <= n; i++ {
			different := make(map[string]bool)
			// l2 is size of language(m2)
			var l2 int
			var w1, w2 map[string]bool

			select {
			case w1 = <-words1:
			case <-kill:
				return
			}
			select {
			case w2 = <-words2:
			case <-kill:
				return
			}

			// for loops implement xor of languages
			for w, res := range w1 {
				if res != w2[w] {
					different[w] = true
				}
			}
			for w, res := range w2 {
				if res != w1[w] {
					different[w] = true
				}
				if res {
					l2++
				}
			}

			if l2 == 0 {
				l2 = 1
			}

			score := float64(len(different)) / float64(l2)
			fmt.Printf("Lang diff: length %d, score %f\n", i, score)
			nDiffs <- score
		}
	}()

	var summaryDiff float64

	var received int
	// use n+1 because we test words of length from 0 to n
	for received < n+1 {
		var end bool
		select {
		case v := <-nDiffs:
			received++
			summaryDiff += v
		case <-kill:
			end = true
			fmt.Println("Lang diff: killed by timeout")
		}
		if end {
			break
		}
	}

	if received == 0 {
		return 0.0
	}

	return summaryDiff / float64(received)
}
