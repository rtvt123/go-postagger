package main

import (
	"time"

	. "github.com/advancedlogic/go-postagger/engine"
)

func main() {
	start := time.Now().UnixNano()
	hmmParser := NewHMMParser("simple.pos")
	hmmParser.Load()
	hmm := NewHMM(hmmParser)
	wordSequence := hmmParser.FWordSequence()
	ws := make([]string, wordSequence.Len())
	cont := 0
	for w := wordSequence.Front(); w != nil; w = w.Next() {
		ws[cont] = w.Value.(string)
		cont++
	}
	hmm.FViterbi(ws)
	stop := time.Now().UnixNano()
	println("test delta", (stop-start)/1000, "mics")
}
