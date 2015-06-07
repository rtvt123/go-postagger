package engine

import (
	"regexp"
	"strings"
)

var re1 = regexp.MustCompile("{[0-9]}*.{[0-9]}*")
var re2 = regexp.MustCompile("({[!-/:-@[-`{-~]}+|{[0-9]}+)+")

type HMM struct {
	tagCounts                  map[string]int64
	wordCounts                 map[string]map[string]int64
	tagBigramCounts            map[string]map[string]int64
	tagForWordCounts           map[string]map[string]int64
	goodTuringTagBigramCounts  map[string]map[string]float64
	goodTuringTagUnigramCounts map[string]float64
	numberOfBigramsWithCount   map[int64]int64
	goodTuringCountsAvailable  bool
	numTrainingBigrams         int64
	mostFreqTag                string
	writer                     string

	ADDONE     bool
	GOODTURING bool
}

func NewHMM(p HMMParser) *HMM {
	instance := new(HMM)
	instance.tagCounts = p.TagCounts
	instance.wordCounts = p.WordCounts
	instance.tagBigramCounts = p.TagBigramCounts
	instance.tagForWordCounts = p.TagForWordCounts
	instance.mostFreqTag = p.MostFreqTag

	instance.goodTuringTagBigramCounts = make(map[string]map[string]float64)
	instance.goodTuringTagUnigramCounts = make(map[string]float64)
	instance.numberOfBigramsWithCount = make(map[int64]int64)
	instance.numTrainingBigrams = p.NumTrainingBigrams

	instance.ADDONE = true
	instance.GOODTURING = false

	return instance
}

func (this *HMM) f1Counts(m map[string]int64, key string) int64 {
	return m[key]
}

func (this *HMM) f2Counts(m map[string]map[string]int64, key1 string, key2 string) int64 {
	if _, exists := m[key1]; exists {
		return this.f1Counts(m[key1], key2)
	} else {
		return 0
	}
}

func (this *HMM) fd1Counts(m map[string]float64, key string) float64 {
	return m[key]
}

func (this *HMM) fd2Counts(m map[string]map[string]float64, key1 string, key2 string) float64 {
	if _, exists := m[key1]; exists {
		return m[key1][key2]
	} else {
		return 0.0
	}
}

func (this *HMM) fNumberOfBigramsWithCount(count int64) int64 {
	return this.numberOfBigramsWithCount[count]
}

func (this *HMM) fMakeGoodTuringCounts() {
	for _, im := range this.tagBigramCounts {
		for _, count := range im {
			this.numberOfBigramsWithCount[count]++
		}
	}

	for tag1, im := range this.tagBigramCounts {
		igtm := make(map[string]float64)
		this.goodTuringTagBigramCounts[tag1] = igtm

		unigramCount := 0.0
		for tag2, count := range im {
			newCount := (float64(count) + 1.0) * (float64(this.fNumberOfBigramsWithCount(count + 1))) / float64(this.fNumberOfBigramsWithCount(count))
			igtm[tag2] = newCount
			unigramCount += newCount
		}
		this.goodTuringTagUnigramCounts[tag1] = unigramCount
	}
	this.goodTuringCountsAvailable = true
}

func (this *HMM) fCalcLikelihood(tag string, word string) float64 {
	if this.ADDONE {
		vocabSize := len(this.tagForWordCounts)
		return float64(this.f2Counts(this.wordCounts, tag, word)+1) / float64(this.f1Counts(this.tagCounts, tag)+int64(vocabSize))
	} else if this.GOODTURING {
		return float64(this.f2Counts(this.wordCounts, tag, word)) / float64(this.fd1Counts(this.goodTuringTagUnigramCounts, tag))
	} else {
		return float64(this.f2Counts(this.wordCounts, tag, word)) / float64(this.f1Counts(this.tagCounts, tag))
	}
}

func (this *HMM) fCalcPriorProb(tag1 string, tag2 string) float64 {
	if this.ADDONE {
		vocabSize := len(this.tagCounts)
		return float64(this.f2Counts(this.tagBigramCounts, tag1, tag2)+1) / float64(this.f1Counts(this.tagCounts, tag1)+int64(vocabSize))
	} else if this.GOODTURING {
		if !this.goodTuringCountsAvailable {
			this.fMakeGoodTuringCounts()
		}
		gtcount := this.fd2Counts(this.goodTuringTagBigramCounts, tag1, tag2)

		if gtcount > 0.0 {
			return gtcount / float64(this.fd1Counts(this.goodTuringTagUnigramCounts, tag1))
		}

		return float64(this.fNumberOfBigramsWithCount(1)) / float64(this.numTrainingBigrams)
	} else {
		return float64(this.f2Counts(this.tagBigramCounts, tag1, tag2)) / float64(this.f1Counts(this.tagCounts, tag1))
	}
}

func (this *HMM) FViterbi(words []string) {
	sentenceStart := true
	var prevMap map[string]*Node
	for i, word := range words {
		sm := make(map[string]*Node)
		if sentenceStart {
			n := NewFullNode(word, "<s>", nil, 1.0)
			sm[word] = n
			sentenceStart = false
		} else {
			if tagcounts, exists := this.tagForWordCounts[word]; exists {
				for tag, _ := range tagcounts {
					sm[tag] = this.fCalcNode(word, tag, prevMap)
				}
			} else if strings.Title(word) == word {
				sm["NNP"] = this.fCalcNode(word, "NNP", prevMap)
			} else if re1.MatchString(word) || re2.MatchString(word) {
				sm["CD"] = this.fCalcNode(word, "CD", prevMap)
			} else if strings.Contains(word, "-") || strings.HasSuffix(word, "able") {
				sm["JJ"] = this.fCalcNode(word, "JJ", prevMap)
			} else if strings.HasPrefix(word, "ing") {
				sm["VBG"] = this.fCalcNode(word, "VBG", prevMap)
			} else if strings.HasPrefix(word, "ly") {
				sm["RB"] = this.fCalcNode(word, "RB", prevMap)
			} else if strings.HasPrefix(word, "ed") {
				sm["VBN"] = this.fCalcNode(word, "VBN", prevMap)
			} else if strings.HasPrefix(word, "s") {
				sm["NNS"] = this.fCalcNode(word, "NNS", prevMap)
			} else {
				//sm[this.mostFreqTag] = this.fCalcNode(word, this.mostFreqTag, prevMap)
				//newNode := this.fCalcUnseenWordNode(word, prevMap)
				//sm[newNode.tag] = newNode

				for tag, _ := range this.tagCounts {
					sm[tag] = this.fCalcNode(word, tag, prevMap)
				}

			}

			if i == len(words)-1 || words[i] == "<s>" {
				this.fBacktrace(sm)
				sentenceStart = true
			}
		}

		prevMap = sm
	}
}

func (this *HMM) fCalcNode(word string, tag string, prevMap map[string]*Node) *Node {
	n := NewSimpleNode(word, tag)
	maxProb := 0.0
	for prevTag, prevNode := range prevMap {
		prevProb := prevNode.prob
		prevProb *= this.fCalcPriorProb(prevTag, tag)
		if prevProb >= maxProb {
			maxProb = prevProb
			n.parent = prevNode
		}
	}

	n.prob = maxProb * this.fCalcLikelihood(tag, word)
	return n
}

func (this *HMM) fBacktrace(m map[string]*Node) {
	n := NewSimpleNode("NOMAX", "NOMAX")
	for _, currentNode := range m {
		if currentNode.prob >= n.prob {
			n = currentNode
		}
	}

	stack := new(Stack)
	for n != nil {
		stack.Push(n)
		n = n.parent
	}

	for stack.Len() != 0 {
		n = stack.Pop().(*Node)
		this.writer += n.tag + " " + n.word + " "
	}

	println(this.writer)
}

func (this *HMM) fCalcUnseenWordNode(word string, prevMap map[string]*Node) *Node {
	maxProb := 0.0
	bestTag := "NOTAG"
	var bestParent *Node

	for prevTag, prevNode := range prevMap {
		prevProb := prevNode.prob
		possibleTagMap := this.tagBigramCounts[prevTag]
		var maxCount int64 = 0
		nextTag := "NOTAG"
		for possibleTag, count := range possibleTagMap {
			if count > maxCount {
				maxCount = count
				nextTag = possibleTag
			}
		}

		prevProb *= this.fCalcPriorProb(prevTag, nextTag)

		if prevProb >= maxProb {
			maxProb = prevProb
			bestTag = nextTag
			bestParent = prevNode
		}
	}

	return NewFullNode(word, bestTag, bestParent, maxProb*this.fCalcLikelihood(bestTag, word))
}
