Sample1 ilustrates scenario with two syntax mistakes. Automata structure is exactly the same apart from final states. It should be noted, that target automaton can be minimized. If start state of this automaton is moved from state 0 to state 5, then in fact the accepted language does not change.
Sample2 ilustrates language difference. In this case automata structure differs thus there are many modifications needed to make them equal. In this case maximum score will be awarded by language difference method.

In both cases there is only one edit necessary to fix the automata, so score awarded by DFA syntac diff method is 90/100
In Sample1 language differs more, so score is relatively lower in this case