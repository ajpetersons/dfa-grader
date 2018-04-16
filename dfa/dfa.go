package dfa

// From github.com/lytics/dfa

import (
	"bytes"
	"errors"
	"fmt"
)

var doneErr = errors.New("out of inputs")

type State string

func (s State) String() string {
	return string(s)
}

type Letter string

func (l Letter) String() string {
	return string(l)
}

var EOF Letter = "EOF"

type DFA struct {
	q      map[State]bool           // States
	e      map[Letter]bool          // Alphabet
	d      map[domainElement]*State // Transition
	q0     State                    // Start State
	f      map[State]bool           // Final States
	input  chan Letter              // Inputs to the DFA
	stop   chan struct{}            // Stops the DFA
	logger func(State)              // Logger for transitions
}

type domainElement struct {
	l Letter
	s State
}

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

// SetStartState, there can be only one.
func (m *DFA) SetStartState(q0 State) {
	m.q0 = q0
}

// SetFinalStates, there can be more than one. If DFA stops when in one of these
// states, input will be marked as accepted
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
	for s, _ := range m.q {
		q = append(q, s)
	}
	return q
}

// Alphabet returns list of letters in the alphabet of the DFA.
func (m *DFA) Alphabet() []Letter {
	e := make([]Letter, 0, len(m.e))
	for l, _ := range m.e {
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
	for s, _ := range m.f {
		if _, ok := m.q[s]; !ok {
			return false,
				fmt.Errorf("final state '%v' is not in the set of states", s)
		}
	}

	return true, nil
}

func (m *DFA) doTransition(s State) (State, error) {
	var next *State
	l := <-m.input
	if l == EOF {
		return s, doneErr
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
		if err == doneErr {
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
