package dfa

import (
	"fmt"
)

func sample() {
	// States
	Starting := State("starting")
	Running := State("running")
	Resending := State("resending")
	Finishing := State("finishing")
	Exiting := State("exiting")
	Terminating := State("terminating")
	// Letters
	Failure := Letter("failure")
	SendFailure := Letter("send-failure")
	SendSuccess := Letter("send-success")
	EverybodyStarted := Letter("everybody-started")
	EverybodyFinished := Letter("everybody-finished")
	ProducersFinished := Letter("producers-finished")
	Exit := Letter("exit")

	inputs := make(chan Letter)
	defer close(inputs)
	d := New(inputs)
	d.SetStartState(Starting)
	d.SetFinalStates(Exiting, Terminating)

	d.SetTransition(Starting, EverybodyStarted, Running)
	d.SetTransition(Starting, Failure, Exiting)
	d.SetTransition(Starting, Exit, Exiting)

	d.SetTransition(Running, SendFailure, Resending)
	d.SetTransition(Running, ProducersFinished, Finishing)
	d.SetTransition(Running, Failure, Exiting)
	d.SetTransition(Running, Exit, Exiting)

	d.SetTransition(Resending, SendSuccess, Running)
	d.SetTransition(Resending, SendFailure, Resending)
	d.SetTransition(Resending, Failure, Exiting)
	d.SetTransition(Resending, Exit, Exiting)

	d.SetTransition(Finishing, EverybodyFinished, Terminating)
	d.SetTransition(Finishing, Failure, Exiting)
	d.SetTransition(Finishing, Exit, Exiting)

	fmt.Println(d.GraphViz())

	d.SetTransitionLogger(func(s State) { fmt.Println(s) })

	go func() {
		inputs <- EverybodyStarted
		inputs <- Failure
		inputs <- EOF
	}()
	fmt.Println(d.Run())
}
