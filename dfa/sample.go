package dfa

import (
	"fmt"
)

func Sample() {
	//minimize
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
	//determinize+minimize
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
	// equiv
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

func Sample4() {
	//get words to n
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

	words := d.getWordsUpToN(10)
	for idx, set := range words {
		fmt.Println(idx, ":")
		for w, accepted := range set {
			if accepted {
				fmt.Println(w)
			}
		}
	}
}

func Sample5() {
	// lang-diff
	// States
	s0 := State("0")
	s1 := State("1")
	s2 := State("2")
	s3 := State("3")
	s4 := State("4")
	s5 := State("5")
	s6 := State("6")
	// Letters
	l0 := Letter("0")
	l1 := Letter("1")

	d := New()
	d.SetStartState(s0)
	d.SetFinalStates(s4, s5)

	d.SetTransition(s0, l0, s1)
	d.SetTransition(s0, l1, s0)

	d.SetTransition(s1, l0, s1)
	d.SetTransition(s1, l1, s2)

	d.SetTransition(s2, l0, s3)
	d.SetTransition(s2, l1, s2)

	d.SetTransition(s3, l0, s3)
	d.SetTransition(s3, l1, s4)

	d.SetTransition(s4, l0, s5)
	d.SetTransition(s4, l1, s4)

	d.SetTransition(s5, l0, s5)
	d.SetTransition(s5, l1, s6)

	d.SetTransition(s6, l0, s6)
	d.SetTransition(s6, l1, s6)

	fmt.Println(d.GraphViz())

	dd := New()
	dd.SetStartState(s0)
	dd.SetFinalStates(s4)

	dd.SetTransition(s0, l0, s1)
	dd.SetTransition(s0, l1, s0)

	dd.SetTransition(s1, l0, s1)
	dd.SetTransition(s1, l1, s2)

	dd.SetTransition(s2, l0, s3)
	dd.SetTransition(s2, l1, s2)

	dd.SetTransition(s3, l0, s3)
	dd.SetTransition(s3, l1, s4)

	dd.SetTransition(s4, l0, s4)
	dd.SetTransition(s4, l1, s5)

	dd.SetTransition(s5, l0, s5)
	dd.SetTransition(s5, l1, s5)

	fmt.Println(dd.GraphViz())

	dr := GetLanguageDifference(dd, d)
	fmt.Println(dr)

	fmt.Println((1 + 0.5*dr) * (1 + 0.5*dr))
}
