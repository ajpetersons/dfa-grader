package server

import (
	"dfa-grader/config"
	"dfa-grader/dfa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"sync"
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
			Status:  "fail",
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
			Status:  "fail",
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
			Status:     "ok",
			Message:    "Graded automata",
			MaxScore:   config.MaxScore,
			TotalScore: config.MaxScore,
		}
		json.NewEncoder(w).Encode(&resp)
		return
	}

	var scaledLangDiffScore, scaledDFASyntaxDiffScore float64
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		langDiffScore := dfa.GetLanguageDifference(dfaAttempt, dfaTarget)
		scaledLangDiffScore = config.MaxScore * langDiffScore

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		dfaSyntaxDiffScore := dfa.GetDFASyntaxDifference(dfaAttempt, dfaTarget)
		if dfaSyntaxDiffScore <= config.DFADiff.MaxDepth {
			scaled := 1 - float64(dfaSyntaxDiffScore)/float64(
				len(dfaAttemptMin.States())*len(dfaAttemptMin.Alphabet()),
			)
			scaledDFASyntaxDiffScore = scaled * config.MaxScore
		}

		wg.Done()
	}()
	wg.Wait()

	totalScore := math.Max(scaledLangDiffScore, scaledDFASyntaxDiffScore)
	fmt.Println("Total time to compute grade:", time.Now().Sub(start))

	resp := response{
		Status:        "ok",
		Message:       "Graded automata",
		MaxScore:      config.MaxScore,
		TotalScore:    totalScore,
		LangDiffScore: scaledLangDiffScore,
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
