# DFA grader

A tool written in GO that grades DFA solutions

## QuickStart
- Install Go
- Install godep
- Fetch dependencies (`dep ensure`)
- Run the server (`go run main.go --serve`)
- For additional help use `go run main.go -h`


## Data
Server accepts such data:
```
POST /grade
{
    "attempt": DFA,     // student attempt
    "target": DFA       // expected automaton
}

DFA: {
    "transitions": array of TRANSITION, // transitions of DFA
    "start_state": string,              // start state of DFA
    "final_states": array of string,    // accepting states of DFA
    "alphabet": array of string         // alphabet of DFA
}

TRANSITION: {
    "from": string,     // from state
    "to": string,       // to state
    "symbol": string    // with symbol
}
```
Note: DFA states are implied from transitions

Response data:
```
{
    "status": "ok / fail",      // short status message
    "message": string,          // human readable status description
    "error": string,            // error description, if any
    "total_score": float,       // achieved score, if successful
    "max_score": float,         // max score
    "lang_diff_score": float,   // achieved score in language difference method
    "dfa_diff_score": float     // achieved score in dfa synatx difference method
}
```

## Footnote
Tool was developed during bachelor's thesis in University of Latvia 2018