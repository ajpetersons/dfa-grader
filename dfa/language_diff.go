package dfa

import (
	"dfa-grader/config"
	"sync"
)

func (m *DFA) getWordsUpToN(n int) []map[string]bool {
	var words []map[string]bool
	var activeStates []map[string]State

	for i := 0; i <= n; i++ {
		activeStates = append(activeStates, make(map[string]State))
		words = append(words, make(map[string]bool))
	}

	activeStates[0][""] = m.q0
	words[0][""] = m.f[m.q0]

	for i := 1; i <= n; i++ {
		for word, state := range activeStates[i-1] {
			for l := range m.e {
				nextState := *m.d[domainElement{l: l, s: state}]
				nextWord := word + string(l)
				// automata is deterministic, there is only one path to nextWord
				activeStates[i][nextWord] = nextState
				words[i][nextWord] = m.f[nextState]
			}
		}
	}

	return words
}

// GetLanguageDifference calculates score given metric to check how many words
// differ for the languages
// Automata MUST be determinized
// m2 is automata that is expected to be received
func GetLanguageDifference(m1, m2 *DFA) float64 {
	n := config.LangDiff.MaxDepth - len(m2.Alphabet())
	if len(m2.Alphabet()) == 5 {
		// worst case
		n--
	}
	if n < config.LangDiff.MaxDepth {
		n = config.LangDiff.MaxDepth
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)
	var words1, words2 []map[string]bool
	go func() {
		words1 = m1.getWordsUpToN(n)
		wg.Done()
	}()
	go func() {
		words2 = m2.getWordsUpToN(n)
		wg.Done()
	}()
	wg.Wait()

	nDiffs := make(chan float64)
	defer close(nDiffs)

	for i := 0; i <= n; i++ {
		go func(n int) {
			different := make(map[string]bool)
			// l2 is size of language(m2)
			var l2 int

			// for loops implement xor of languages
			for w, res := range words1[n] {
				if res != words2[n][w] {
					different[w] = true
				}
			}
			for w, res := range words2[n] {
				if res != words1[n][w] {
					different[w] = true
				}
				if res {
					l2++
				}
			}

			if l2 == 0 {
				l2 = 1
			}

			nDiffs <- float64(len(different)) / float64(l2)
		}(i)

	}

	var summaryDiff float64

	var received int
	// use n+1 because we test words of length from 0 to n
	for received < n+1 {
		v := <-nDiffs
		received++
		summaryDiff += v
	}

	return summaryDiff / float64(n+1)
}
