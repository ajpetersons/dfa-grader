package dfa

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
// m2 is automata that is expected to be received
func GetLanguageDifference(m1, m2 *DFA) float64 {
	n := 2 * len(m2.States())
	if n < 10 {
		n = 10
	}

	words1 := m1.getWordsUpToN(n)
	words2 := m2.getWordsUpToN(n)

	var different []map[string]bool
	var summaryDiff float64

	for i := 0; i <= n; i++ {
		different = append(different, make(map[string]bool))
		// l2 is size of language(m2)
		var l2 int

		// for loops implement xor of languages
		for w, res := range words1[i] {
			if res != words2[i][w] {
				different[i][w] = true
			}
		}
		for w, res := range words2[i] {
			if res != words1[i][w] {
				different[i][w] = true
			}
			if res {
				l2++
			}
		}

		if l2 == 0 {
			l2 = 1
		}

		summaryDiff += float64(len(different[i])) / float64(l2)
	}

	return summaryDiff / float64(n)
}
