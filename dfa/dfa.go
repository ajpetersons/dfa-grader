package dfa

// From github.com/lytics/dfa

import (
	"bytes"
	"errors"
	"fmt"
)

var errDone = errors.New("out of inputs")

// State describes single state in DFA
type State string

func (s State) String() string {
	return string(s)
}

// Letter describes allowed input symbol
type Letter string

func (l Letter) String() string {
	return string(l)
}

// EOF is used to mark end of input
var EOF Letter = "EOF"

// DFA describes a dfa with some helper fields
type DFA struct {
	q  map[State]bool           // States
	e  map[Letter]bool          // Alphabet
	d  map[domainElement]*State // Transition
	q0 State                    // Start State
	f  map[State]bool           // Final States

	input  chan Letter   // Inputs to the DFA
	stop   chan struct{} // Stops the DFA
	logger func(State)   // Logger for transitions
}

type domainElement struct {
	l Letter
	s State
}

// New creates empty DFA
func New(inputs chan Letter) *DFA {
	return &DFA{
		q:      make(map[State]bool),
		e:      make(map[Letter]bool),
		f:      make(map[State]bool),
		d:      make(map[domainElement]*State),
		stop:   make(chan struct{}),
		input:  inputs,
		logger: func(State) {},
	}
}

// SetTransition adds new transition to DFA
func (m *DFA) SetTransition(from State, input Letter, to State) error {
	if from == State("") || to == State("") {
		return errors.New("state cannot be defined as the empty string")
	}

	m.q[to] = true
	m.q[from] = true
	m.e[input] = true
	de := domainElement{l: input, s: from}
	if _, ok := m.d[de]; !ok {
		m.d[de] = &to
	}

	return nil
}

// SetLetter adds a new symbol to alphabet
func (m *DFA) SetLetter(l Letter) {
	m.e[l] = true
}

// SetState adds a new state to list of states
func (m *DFA) SetState(q State) {
	m.q[q] = true
}

// SetStartState sets q0, there can be only one.
func (m *DFA) SetStartState(q0 State) {
	m.q0 = q0
}

// SetFinalStates marks final states, there can be more than one. If DFA stops
// when in one of these states, input will be marked as accepted
func (m *DFA) SetFinalStates(f ...State) {
	for _, q := range f {
		m.f[q] = true
	}
}

func (m *DFA) SetTransitionLogger(logger func(State)) {
	m.logger = logger
}

// States returns list of states in the DFA.
func (m *DFA) States() []State {
	q := make([]State, 0, len(m.q))
	for s := range m.q {
		q = append(q, s)
	}
	return q
}

// Alphabet returns list of letters in the alphabet of the DFA.
func (m *DFA) Alphabet() []Letter {
	e := make([]Letter, 0, len(m.e))
	for l := range m.e {
		e = append(e, l)
	}
	return e
}

// Valid checks if start state exists and is within set of DFA's states. Also
// checks if all final states are within set of DFA's states
func (m *DFA) Valid() (bool, error) {
	if m.q0 == State("") {
		return false, errors.New("no start state defined")
	}
	if _, ok := m.q[m.q0]; !ok {
		return false,
			fmt.Errorf("start state '%v' is not in the set of states", m.q0)
	}
	for s := range m.f {
		if _, ok := m.q[s]; !ok {
			return false,
				fmt.Errorf("final state '%v' is not in the set of states", s)
		}
	}

	return true, nil
}

// Determinize adds missing transitions to automata
func (m *DFA) Determinize() error {
	binName := "bin"
	for m.q[State(binName)] {
		binName += "_bin"
	}
	bin := State(binName)
	m.SetState(bin)

	for s := range m.q {
		for l := range m.e {
			if _, ok := m.d[domainElement{l: l, s: s}]; !ok {
				err := m.SetTransition(s, l, bin)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Minimize removes obsolete transitions and minimizes the DFA
func (m *DFA) Minimize() {
	m.removeUnreachable()
	m.mergeNonDistinguishable()
}

func (m *DFA) removeUnreachable() {
	// let reachable_states := {q0};
	// let new_states := {q0};
	// do {
	//     temp := the empty set;
	//     for each q in new_states do
	//         for each c in Σ do
	//             temp := temp ∪ {p such that p = δ(q,c)};
	//         end;
	//     end;
	//     new_states := temp \ reachable_states;
	//     reachable_states := reachable_states ∪ new_states;
	// } while (new_states ≠ the empty set);
	// unreachable_states := Q \ reachable_states;
	reachable := make(map[State]bool)
	reachable[m.q0] = true
	newStates := make(map[State]bool)
	newStates[m.q0] = true
	for len(newStates) != 0 {
		reached := make(map[State]bool)
		for s := range newStates {
			for l := range m.e {
				reached[*m.d[domainElement{s: s, l: l}]] = true
			}
		}
		for s := range reachable {
			delete(reached, s)
		}
		newStates = reached
		for s := range newStates {
			reachable[s] = true
		}
	}
	unreachable := make(map[State]bool)
	for s := range m.q {
		if !reachable[s] {
			unreachable[s] = true
		}
	}

	for s := range unreachable {
		if m.f[s] {
			delete(m.f, s)
		}

		for l := range m.e {
			delete(m.d, domainElement{s: s, l: l})
		}
	}
}

func (m *DFA) mergeNonDistinguishable() {
	type doubleState struct {
		a, b State
	}
	distinguishable := make(map[doubleState]bool)
	for f := range m.f {
		for s := range m.q {
			if !m.f[s] {
				distinguishable[doubleState{a: s, b: f}] = true
				distinguishable[doubleState{a: f, b: s}] = true
			}
		}
	}

	for {
		shouldBreak := true
		for s1 := range m.q {
			for s2 := range m.q {
				if distinguishable[doubleState{a: s1, b: s2}] {
					continue
				}
				for l := range m.e {
					pair := doubleState{
						a: *m.d[domainElement{s: s1, l: l}],
						b: *m.d[domainElement{s: s2, l: l}],
					}
					if distinguishable[pair] {
						shouldBreak = false
						distinguishable[doubleState{a: s1, b: s2}] = true
						distinguishable[doubleState{a: s2, b: s1}] = true
					}
				}
			}
		}
		if shouldBreak {
			break
		}
	}

	for s1 := range m.q {
		for s2 := range m.q {
			if s1 == s2 {
				continue
			}
			if !distinguishable[doubleState{a: s1, b: s2}] {
				for k, v := range m.d {
					if *v == s2 {
						*v = s1
					}
					if k.s == s2 {
						delete(m.d, k)
						m.d[domainElement{s: s1, l: k.l}] = v
					}
				}
				delete(m.f, s2)
				if m.q0 == s2 {
					m.q0 = s1
				}
				delete(m.q, s2)
			}
		}
	}
}

func (m *DFA) doTransition(s State) (State, error) {
	var next *State
	l := <-m.input
	if l == EOF {
		return s, errDone
	}
	// Reject upfront if letter is not in alphabet.
	if !m.e[l] {
		return s, fmt.Errorf("letter '%v' is not in alphabet", l)
	}
	// Check the transition function.
	if next = m.d[domainElement{l: l, s: s}]; next != nil {
		m.logger(*next)
	} else {
		// Otherwise stop the DFA with a rejected state,
		// the DFA has rejected the input sequence.
		return s, fmt.Errorf(
			"no state transition for input '%v' from '%v'",
			l, s,
		)
	}
	return *next, nil
}

// Run the DFA, blocking until Stop is called or inputs run out.
// Returns the last state and true if the last state was a final state.
// Use EOF to indicate end of inout
func (m *DFA) Run() (State, bool, error) {
	valid, err := m.Valid()
	if !valid {
		return State(""), false, err
	}

	// Run the DFA.
	// The current state, starts at q0.
	s := m.q0
	m.logger(s)
	for {
		var stopnow bool
		select {
		case <-m.stop:
			stopnow = true
		default:
		}
		if stopnow {
			break
		}
		s, err = m.doTransition(s)
		if err == errDone {
			break
		}
		if err != nil {
			return State(""), false, err
		}
	}
	// check if the current state is accepted or rejected by the DFA.
	var accepted bool
	if m.f[s] {
		accepted = true
	}
	return s, accepted, nil
}

// Stop the DFA.
func (m *DFA) Stop() {
	close(m.stop)
}

// GraphViz representation string which can be copy-n-pasted into
// any online tool like http://graphs.grevian.org/graph to get
// a diagram of the DFA.
func (m *DFA) GraphViz() string {
	var buf bytes.Buffer
	buf.WriteString("digraph {\n")
	for do, to := range m.d {
		buf.WriteString(fmt.Sprintf(
			"    \"%s\" -> \"%s\"[label=\"%s\"];\n",
			do.s, to, do.l,
		))
	}
	buf.WriteString("}")
	return buf.String()
}
