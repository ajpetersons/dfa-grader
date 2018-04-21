package dfa

import (
	"fmt"
)

func Sample() {
	// States
	se := State("e")
	s0 := State("0")
	s1 := State("1")
	s00 := State("00")
	s01 := State("01")
	s10 := State("10")
	s11 := State("11")
	// Letters
	l0 := Letter("0")
	l1 := Letter("1")

	d := New()
	d.SetStartState(se)
	d.SetFinalStates(s11)

	d.SetTransition(se, l0, s0)
	d.SetTransition(se, l1, s1)

	d.SetTransition(s0, l0, s00)
	d.SetTransition(s0, l1, s01)

	d.SetTransition(s1, l0, s10)
	d.SetTransition(s1, l1, s11)

	d.SetTransition(s00, l0, s00)
	d.SetTransition(s00, l1, s01)

	d.SetTransition(s01, l0, s10)
	d.SetTransition(s01, l1, s11)

	d.SetTransition(s10, l0, s00)
	d.SetTransition(s10, l1, s01)

	d.SetTransition(s11, l0, s10)
	d.SetTransition(s11, l1, s11)

	fmt.Println(d.GraphViz())

	d.Minimize()

	fmt.Println(d.GraphViz())
}

func Sample2() {
	// States
	sa := State("a")
	sb := State("b")
	sc := State("c")
	sd := State("d")
	se := State("e")
	// sf := State("f")
	// Letters
	l0 := Letter("0")
	l1 := Letter("1")

	d := New()
	d.SetStartState(sa)
	d.SetFinalStates(sc, sd, se)

	d.SetTransition(sa, l0, sb)
	d.SetTransition(sa, l1, sc)

	d.SetTransition(sb, l0, sa)
	d.SetTransition(sb, l1, sd)

	d.SetTransition(sc, l0, se)
	// d.SetTransition(sc, l1, sf)

	d.SetTransition(sd, l0, se)
	// d.SetTransition(sd, l1, sf)

	d.SetTransition(se, l0, se)
	// d.SetTransition(se, l1, sf)

	// d.SetTransition(sf, l0, sf)
	// d.SetTransition(sf, l1, sf)

	fmt.Println(d.GraphViz())

	err := d.Determinize()
	if err != nil {
		panic(err)
	}
	d.Minimize()

	fmt.Println(d.GraphViz())
}

func Sample3() {
	// States
	sa := State("a")
	sb := State("b")
	sc := State("c")
	sd := State("d")
	se := State("e")
	// sf := State("f")
	// Letters
	l0 := Letter("0")
	l1 := Letter("1")

	d := New()
	d.SetStartState(sa)
	d.SetFinalStates(sc, sd, se)

	d.SetTransition(sa, l0, sb)
	d.SetTransition(sa, l1, sc)

	d.SetTransition(sb, l0, sa)
	d.SetTransition(sb, l1, sd)

	d.SetTransition(sc, l0, se)
	// d.SetTransition(sc, l1, sf)

	d.SetTransition(sd, l0, se)
	// d.SetTransition(sd, l1, sf)

	d.SetTransition(se, l0, se)
	// d.SetTransition(se, l1, sf)

	// d.SetTransition(sf, l0, sf)
	// d.SetTransition(sf, l1, sf)

	err := d.Determinize()
	if err != nil {
		panic(err)
	}
	d.Minimize()
	fmt.Println(d.GraphViz())

	dd := New()
	dd.SetStartState(sa)
	dd.SetFinalStates(sc, sd, se)

	dd.SetTransition(sa, l0, sb)
	dd.SetTransition(sa, l1, sc)

	dd.SetTransition(sb, l0, sa)
	dd.SetTransition(sb, l1, sb)

	dd.SetTransition(sc, l0, se)
	// dd.SetTransition(sc, l1, sf)

	dd.SetTransition(sd, l0, se)
	// dd.SetTransition(sd, l1, sf)

	dd.SetTransition(se, l0, se)
	// dd.SetTransition(se, l1, sf)

	// dd.SetTransition(sf, l0, sf)
	// dd.SetTransition(sf, l1, sf)

	err = dd.Determinize()
	if err != nil {
		panic(err)
	}
	fmt.Println(dd.GraphViz())
	dd.Minimize()
	fmt.Println(dd.GraphViz())

	fmt.Println(d.Equiv(dd))

	fmt.Println(d.GraphViz())
}
