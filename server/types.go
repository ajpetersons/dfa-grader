package server

type transition struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Symbol string `json:"symbol"`
}

type automata struct {
	Transitions []transition `json:"transitions"`
	StartState  string       `json:"start_state"`
	FinalStates []string     `json:"final_states"`
	Alphabet    []string     `json:"alphabet"`
}

type response struct {
	Message       string  `json:"message"`
	Error         string  `json:"error,omitempty"`
	TotalScore    float64 `json:"total_score,omitempty"`
	MaxScore      float64 `json:"max_score,omitempty"`
	LangDiffScore float64 `json:"lang_diff_score,omitempty"`
	DFADiffScore  float64 `json:"dfa_diff_score,omitempty"`
}
