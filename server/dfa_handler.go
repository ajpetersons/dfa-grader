package server

import (
	"dfa-grader/config"
	"dfa-grader/dfa"
	"dfa-grader/grader"
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
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		resp := response{
			Status:  "fail",
			Message: "Request data too large",
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(&resp)
		return
	}

	// validate data
	var data struct {
		Attempt automata `json:"attempt"`
		Target  automata `json:"target"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		resp := response{
			Status:  "fail",
			Message: "Unable to process request data",
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(&resp)
		return
	}

	start := time.Now()
	fmt.Println("Received grading request")

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

	eq, err := dfa.Compare(dfaAttemptMin, dfaTargetMin)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := response{
			Status:  "fail",
			Message: "Could not minimize DFA",
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(&resp)
		return
	}
	if eq {
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
		langDiffScore := grader.GetLanguageDifference(dfaAttempt, dfaTarget)

		scaledLangDiffScore = config.MaxScore * langDiffScore

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		dfaSyntaxDiffScore := grader.GetDFASyntaxDifference(dfaAttempt, dfaTarget)

		scaledDFASyntaxDiffScore = config.MaxScore * dfaSyntaxDiffScore

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

	for _, s := range a.States {
		m.SetState(dfa.State(s))
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
