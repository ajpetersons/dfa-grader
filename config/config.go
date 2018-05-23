package config

import (
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

const (
	maxScoreKey = "maxScore"
	timeoutKey  = "timeout"

	langDiffKey = "langDiff."
	maxDepthKey = "maxDepth"
	minDepthKey = "minDepth"

	dfaDiffKey = "dfaSyntaxDiff."
)

type langDiff struct {
	MaxDepth int
	MinDepth int
	Timeout  time.Duration
}

type dfaDiff struct {
	MaxDepth int
	Timeout  time.Duration
}

var (
	// MaxScore is maximum possible score for DFA
	MaxScore float64
	// LangDiff has all parameters for running language diff calculation
	LangDiff langDiff
	// DFADiff has all parameters to find dfa syntax mistakes
	DFADiff dfaDiff
)

// Read prepares config file
func Read(filename string) error {
	viper.SetDefault(maxScoreKey, 100.0)
	viper.SetDefault(langDiffKey+timeoutKey, 3*time.Second)
	viper.SetDefault(langDiffKey+maxDepthKey, 14)
	viper.SetDefault(langDiffKey+minDepthKey, 4)
	viper.SetDefault(dfaDiffKey+maxDepthKey, 2)
	viper.SetDefault(dfaDiffKey+timeoutKey, 3*time.Second)

	if filename != "" {
		viper.SetConfigName(filepath.Base(filename))
		viper.AddConfigPath(filepath.Dir(filename))

		err := viper.ReadInConfig()
		if err != nil {
			return err
		}
	}

	MaxScore = viper.GetFloat64(maxScoreKey)
	LangDiff = langDiff{
		MaxDepth: viper.GetInt(langDiffKey + maxDepthKey),
		MinDepth: viper.GetInt(langDiffKey + minDepthKey),
		Timeout:  viper.GetDuration(langDiffKey + timeoutKey),
	}
	DFADiff = dfaDiff{
		MaxDepth: viper.GetInt(dfaDiffKey + maxDepthKey),
		Timeout:  viper.GetDuration(dfaDiffKey + timeoutKey),
	}

	return nil
}
