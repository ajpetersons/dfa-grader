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
	"github.com/pkg/errors"
)

// dfaHandler may hold any specific variables needed for this handler
type dfaHandler struct{}

func newDFAHandler() *dfaHandler {
	return &dfaHandler{}
}

// register adds endpoints to this handler
func (h *dfaHandler) register(r *mux.Router) {
	r.HandleFunc("/grade", h.handleDFATest).Methods(http.MethodPost)
}

func (h *dfaHandler) handleDFATest(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, 1024*1024*10))
	if err != nil {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		resp := response{
			Status:  "fail",
			Message: "Request data too large",
			Error:   err.Error(),
		}
		encodeResponse(w, &resp)
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
		encodeResponse(w, &resp)
		return
	}

	start := time.Now()
	fmt.Println("Received grading request")

	dfaAttempt, err := createDFA(data.Attempt)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		resp := response{
			Status:  "fail",
			Message: "Unable to create attempted DFA",
			Error:   err.Error(),
		}
		encodeResponse(w, &resp)
		return
	}
	dfaTarget, err := createDFA(data.Target)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		resp := response{
			Status:  "fail",
			Message: "Unable to create target DFA",
			Error:   err.Error(),
		}
		encodeResponse(w, &resp)
		return
	}
	err = dfaAttempt.Determinize()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := response{
			Status:  "fail",
			Message: "Could not parse attempted solution dfa",
			Error:   err.Error(),
		}
		encodeResponse(w, &resp)
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
		encodeResponse(w, &resp)
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
		encodeResponse(w, &resp)
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
		encodeResponse(w, &resp)
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
	fmt.Println("Total time to compute grade:", time.Since(start))

	resp := response{
		Status:        "ok",
		Message:       "Graded automata",
		MaxScore:      config.MaxScore,
		TotalScore:    totalScore,
		LangDiffScore: scaledLangDiffScore,
		DFADiffScore:  scaledDFASyntaxDiffScore,
	}
	w.WriteHeader(http.StatusOK)
	encodeResponse(w, &resp)
}

func encodeResponse(w http.ResponseWriter, data interface{}) {
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// nolint: gocyclo
func createDFA(a automata) (*dfa.DFA, error) {
	m := dfa.New()

	if len(a.Alphabet) == 0 {
		return nil, errors.New("alphabet should not be empty")
	}
	for _, l := range a.Alphabet {
		m.SetLetter(dfa.Letter(l))
	}

	if len(a.States) == 0 {
		return nil, errors.New("automata should have at least one state")
	}
	for _, s := range a.States {
		m.SetState(dfa.State(s))
	}

	if a.StartState == "" {
		return nil, errors.New("start state should not be empty")
	}
	if !m.HasState(dfa.State(a.StartState)) {
		return nil, errors.New("start state not in list of states")
	}
	m.SetStartState(dfa.State(a.StartState))

	finals := []dfa.State{}
	for _, f := range a.FinalStates {
		if !m.HasState(dfa.State(f)) {
			return nil, errors.Errorf(
				"final state '%s' not in list of states", f,
			)
		}
		finals = append(finals, dfa.State(f))
	}
	m.SetFinalStates(finals...)

	for _, t := range a.Transitions {
		if !m.HasState(dfa.State(t.From)) {
			return nil, errors.Errorf(
				"transition state '%s' not in list of states", t.From,
			)
		}
		if !m.HasState(dfa.State(t.To)) {
			return nil, errors.Errorf(
				"transition state '%s' not in list of states", t.To,
			)
		}
		if !m.HasLetter(dfa.Letter(t.Symbol)) {
			return nil, errors.Errorf(
				"transition symbol '%s' not in list of states", t.Symbol,
			)
		}
		m.SetTransition( // nolint: errcheck
			dfa.State(t.From),
			dfa.Letter(t.Symbol),
			dfa.State(t.To),
		)
	}

	return m, nil
}
