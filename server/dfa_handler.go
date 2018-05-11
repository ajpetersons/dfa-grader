package server

import (
	"dfa-grader/dfa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type DFAHandler struct{}

func NewDFAHandler() *DFAHandler {
	return &DFAHandler{}
}

func (h *DFAHandler) Register(r *mux.Router) {
	r.HandleFunc("/grade", h.HandleDFATest).Methods(http.MethodPost)
}

func (h *DFAHandler) HandleDFATest(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, 1024*1024*10))
	if err != nil {
		http.Error(w, "request data too large",
			http.StatusRequestEntityTooLarge)
		return
	}

	// validate data
	var data struct {
		Attempt automata `json:"attempt"`
		Target  automata `json:"target"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(
			w, fmt.Sprintf("unable to process request data: %s", err.Error()),
			http.StatusUnprocessableEntity,
		)
		return
	}

	start := time.Now()

	dfaAttempt := createDFA(data.Attempt)
	dfaTarget := createDFA(data.Target)
	err = dfaAttempt.Determinize()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := response{
			Message: "Could not parse attempted solution dfa",
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(&resp)
		return
	}
	err = dfaTarget.Determinize()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := response{
			Message: "Could not parse target dfa",
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(&resp)
		return
	}

	dfaAttemptMin := dfaAttempt.Copy()
	dfaTargetMin := dfaTarget.Copy()

	dfaAttemptMin.Minimize()
	dfaTargetMin.Minimize()

	if dfaAttemptMin.Equiv(dfaTargetMin) {
		w.WriteHeader(http.StatusOK)
		resp := response{
			Message:    "Graded automata",
			MaxScore:   100.0,
			TotalScore: 100.0,
		}
		json.NewEncoder(w).Encode(&resp)
		return
	}

	// TODO: to goroutines
	langDiffScore := dfa.GetLanguageDifference(dfaAttempt, dfaTarget)
	dfaSyntaxDiffScore := dfa.GetDFASyntaxDifference(dfaAttempt, dfaTarget)

	var scaledDFASyntaxDiffScore float64
	if dfaSyntaxDiffScore <= dfa.DFASyntaxDiffWorstPosScore {
		scaledDFASyntaxDiffScore = 1 - (float64(dfaSyntaxDiffScore) / float64(
			len(dfaAttemptMin.States())*len(dfaAttemptMin.Alphabet()),
		))
		scaledDFASyntaxDiffScore *= 100.0
	}
	totalScore := math.Max(langDiffScore, scaledDFASyntaxDiffScore)
	fmt.Println("Total time to compute grade:", time.Now().Sub(start))

	resp := response{
		Message:       "Graded automata",
		MaxScore:      100.0,
		TotalScore:    totalScore,
		LangDiffScore: 100.0 * langDiffScore,
		DFADiffScore:  scaledDFASyntaxDiffScore,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&resp)
}

func createDFA(a automata) *dfa.DFA {
	m := dfa.New()

	for _, l := range a.Alphabet {
		m.SetLetter(dfa.Letter(l))
	}

	m.SetStartState(dfa.State(a.StartState))

	finals := []dfa.State{}
	for _, f := range a.FinalStates {
		finals = append(finals, dfa.State(f))
	}
	m.SetFinalStates(finals...)

	for _, t := range a.Transitions {
		m.SetTransition(
			dfa.State(t.From),
			dfa.Letter(t.Symbol),
			dfa.State(t.To),
		)
	}

	return m
}
