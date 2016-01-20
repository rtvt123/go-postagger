package main

import (
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
)

var (
	train = kingpin.Flag("train", "Enable train model").Short('t').Bool()
	model = kingpin.Flag("model", "HMM Trained Model").Short('m').Required().String()
	text  = kingpin.Flag("text", "Text to process").Short('t').Required().String()
)

func main() {
	kingpin.Parse()
	if *train {
		println("START")
		start := time.Now().UnixNano()
		ptrain := NewHMMParser("training/training.set")
		println("Training set loaded")
		ptrain.fParseTrainer()
		println("Training set trained")
		ptrain.Save()
		stop := time.Now().UnixNano()
		println("training delta", (stop-start)/1000, "mics")
	} else {
		println("Loading model " + *model)
		hmmParser := NewHMMParser(*model)
		hmmParser.Load()
		hmm := NewHMM(hmmParser)
		println("Model " + *model + " loaded")

		start := time.Now().UnixNano()
		ws := strings.Split("<s> "+*text, " ")
		hmm.fViterbi(ws)
		println(hmm.String())
		stop := time.Now().UnixNano()
		println("test delta", (stop-start)/1000, "mics")
	}
}
