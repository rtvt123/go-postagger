package engine

import (
	"bytes"
	"container/list"
	"encoding/gob"
	"io/ioutil"
	"log"
)

type HMMParser struct {
	scanner          *Scanner
	TagCounts        map[string]int64
	WordCounts       map[string]map[string]int64
	TagBigramCounts  map[string]map[string]int64
	TagForWordCounts map[string]map[string]int64

	MostFreqTag        string
	MostFreqTagCount   int64
	NumTrainingBigrams int64
}

func NewHMMParser(filename string) HMMParser {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Panicln(err.Error())
	}
	scanner := NewScannerString(string(file))
	return HMMParser{
		scanner:            scanner,
		TagCounts:          make(map[string]int64),
		WordCounts:         make(map[string]map[string]int64),
		TagBigramCounts:    make(map[string]map[string]int64),
		TagForWordCounts:   make(map[string]map[string]int64),
		MostFreqTag:        "",
		MostFreqTagCount:   0,
		NumTrainingBigrams: 0,
	}
}

func (this *HMMParser) FParseTrainer() {
	prevTag := this.scanner.Next()
	this.scanner.Next()

	for this.scanner.HasNext() {
		currentTag := this.scanner.Next()
		currentWord := this.scanner.Next()

		this.f2AddOne(this.TagCounts, currentTag)
		this.f3AddOne(this.WordCounts, currentTag, currentWord)
		this.f3AddOne(this.TagBigramCounts, prevTag, currentTag)
		this.f3AddOne(this.TagForWordCounts, currentWord, currentTag)

		if this.TagCounts[currentTag] >= this.MostFreqTagCount {
			this.MostFreqTagCount = this.TagCounts[currentTag]
			this.MostFreqTag = currentTag
		}

		this.NumTrainingBigrams++
		prevTag = currentTag
	}
}

func (this *HMMParser) FWordSequence() *list.List {
	l := list.New()
	for this.scanner.HasNext() {
		this.scanner.Next()
		l.PushBack(this.scanner.Next())
	}
	return l
}

func (this *HMMParser) f2AddOne(m map[string]int64, key1 string) {
	m[key1]++
}

func (this *HMMParser) f3AddOne(m map[string]map[string]int64, key1 string, key2 string) {
	if _, exists := m[key1]; exists {
		this.f2AddOne(m[key1], key2)
	} else {
		sm := make(map[string]int64)
		sm[key2] = 1
		m[key1] = sm
	}
}

func (this *HMMParser) Save() {
	m := new(bytes.Buffer)
	enc := gob.NewEncoder(m)

	err := enc.Encode(this)
	if err != nil {
		log.Panicln(err)
	}

	err = ioutil.WriteFile("hmm.dat", m.Bytes(), 0600)
	if err != nil {
		log.Panicln(err)
	}
}

func (this *HMMParser) Load() {
	n, err := ioutil.ReadFile("hmm.dat")
	if err != nil {
		log.Println(err.Error())
		this.FParseTrainer()
		this.Save()
	} else {
		var hmmParser HMMParser
		p := bytes.NewBuffer(n)
		dec := gob.NewDecoder(p)

		err = dec.Decode(&hmmParser)
		if err != nil {
			log.Println(err.Error())
		} else {
			this.TagCounts = hmmParser.TagCounts
			this.WordCounts = hmmParser.WordCounts
			this.TagBigramCounts = hmmParser.TagBigramCounts
			this.TagForWordCounts = hmmParser.TagForWordCounts
			this.MostFreqTag = hmmParser.MostFreqTag
			this.MostFreqTagCount = hmmParser.MostFreqTagCount
			this.NumTrainingBigrams = hmmParser.NumTrainingBigrams
		}
	}
}
