package builder

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const sentenceStart = "__start__"
const sentenceEnd = "__end__"

type Node struct {
	Count  int
	Weight float64
}

var randomSource *rand.Rand
var sentenceEnders = []byte{33, 46, 63} // ? ! .

func isEndOfSentence(word string) bool {
	lastChar := word[len(word)-1]
	for _, r := range sentenceEnders {
		if lastChar == r {
			return true
		}
	}
	return false
}

// Build takes in raw text and creates a word map suitable for generating sentences
func Build(input string) map[string]map[string]*Node {
	// first replace all newlines with spaces
	text := strings.Replace(input, "\n", " ", -1)
	words := strings.Split(text, " ")

	// create our map
	wordMap := make(map[string]map[string]*Node)
	addOrUpdateMap := func(where, word string) {
		// create child map if necessary
		if _, ok := wordMap[where]; !ok {
			wordMap[where] = make(map[string]*Node)
		}
		// see if this word is already in the map
		if _, found := wordMap[where][word]; found {
			wordMap[where][word].Count++
		} else {
			wordMap[where][word] = &Node{
				Count: 1,
			}
		}
	}

	atSentenceStart := false
	for idx, w := range words {
		// trim and lowercase all words
		word := strings.ToLower(strings.TrimSpace(w))

		// if word is empty skip it
		if len(word) < 1 {
			continue
		}

		// if at beginning of sentence, we want to
		// consider the first words children of __start__
		if idx == 0 || atSentenceStart {
			addOrUpdateMap(sentenceStart, word)
			atSentenceStart = false
		}

		// if we are at the end of a sentence, deal with that in a similar way, and set a flag
		if idx == len(words)-1 || isEndOfSentence(word) {
			addOrUpdateMap(sentenceEnd, word)
			atSentenceStart = true
		}

		// confirm we aren't at the end of the word list
		if idx+1 == len(words) {
			break
		}

		addOrUpdateMap(word, strings.ToLower(strings.TrimSpace(words[idx+1])))
	}

	// now need to update Weights of each word
	for word, nextWords := range wordMap {
		sum := 0
		for _, props := range nextWords {
			sum += props.Count
		}
		for nextWord, props := range nextWords {
			wordMap[word][nextWord].Weight = float64(props.Count) / float64(sum)
		}
	}

	return wordMap
}

func getNextWord(currentWord string, wordMap map[string]map[string]*Node) string {
	possibleNextWords := wordMap[currentWord]

	for {
		r := randomSource.Float64()
		possibilities := make([]string, 0)
		for possibleNextWord, props := range possibleNextWords {
			if r < props.Weight {
				possibilities = append(possibilities, possibleNextWord)
			}
		}
		if len(possibilities) > 0 {
			randomIndex := randomSource.Intn(len(possibilities))
			return possibilities[randomIndex]
		}
	}
}

// GenerateSentence uses a word map to generate a sentence
func GenerateSentence(wordMap map[string]map[string]*Node) string {
	sentence := make([]string, 1)

	currentWord := sentenceStart
	done := false
	for {
		if done {
			break
		}

		nextWord := getNextWord(currentWord, wordMap)
		sentence = append(sentence, nextWord)
		currentWord = nextWord

		// see if we're possibly at the end
		for w := range wordMap[sentenceEnd] {
			if currentWord == w {
				done = true
				break
			}
		}
	}

	// remove leading space and return
	return strings.Join(sentence[1:], " ")
}

// ReadMap takes raw bytes of the json output of a word map and unmarshals it
func ReadMap(file []byte) map[string]map[string]*Node {
	wordMap := make(map[string]map[string]*Node)
	json.Unmarshal(file, &wordMap)
	return wordMap
}

// PrintMap takes a word map and pretty prints it
func PrintMap(wordMap map[string]map[string]*Node) {
	for word, nextWords := range wordMap {
		fmt.Println(word)
		for nextWord, props := range nextWords {
			fmt.Println("\t", nextWord, props.Count, props.Weight)
		}
	}
}

func init() {
	randomSource = rand.New(rand.NewSource(time.Now().UnixNano()))
}
