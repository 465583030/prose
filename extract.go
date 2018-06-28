package prose

import (
	"regexp"
	"strconv"
	"strings"
)

// MaxentClassifier ...
type MaxentClassifier struct {
	Labels  []string
	Words   []string
	Mapping map[string]int
	Weights []float64
}

// NewMaxentClassifier ...
func NewMaxentClassifier(weights []float64, mapping map[string]int, labels, words []string) *MaxentClassifier {
	return &MaxentClassifier{labels, words, mapping, weights}
}

// EntityExtracter ...
type EntityExtracter struct {
	classifier *MaxentClassifier
}

// NewEntityExtracter ...
func NewEntityExtracter() *EntityExtracter {
	var mapping map[string]int
	var weights []float64
	var labels []string
	var words []string

	dec := getAsset("Maxent", "mapping.gob")
	checkError(dec.Decode(&mapping))

	dec = getAsset("Maxent", "weights.gob")
	checkError(dec.Decode(&weights))

	dec = getAsset("Maxent", "words.gob")
	checkError(dec.Decode(&words))

	dec = getAsset("Maxent", "labels.gob")
	checkError(dec.Decode(&labels))

	return &EntityExtracter{classifier: NewMaxentClassifier(weights, mapping, labels, words)}
}

// Chunk ...
func (e *EntityExtracter) Chunk(tokens []Token) []Entity {
	entities := []Entity{}
	end := ""

	parts := []Token{}
	idx := 0

	for _, tok := range tokens {
		label := tok.Label
		if (label != "O" && label != end) || (idx > 0 && tok.Tag == parts[idx-1].Tag) {
			end = strings.Replace(label, "B", "I", 1)
			parts = append(parts, tok)
			idx++
		} else if (label == "O" && end != "") || label == end {
			// We've found the end of an entity.
			if label != "O" {
				parts = append(parts, tok)
			}
			entities = append(entities, coalesce(parts))

			end = ""
			parts = []Token{}
			idx = 0
		}
	}

	return entities
}

func coalesce(parts []Token) Entity {
	labels := []string{}
	tokens := []string{}
	for _, tok := range parts {
		tokens = append(tokens, tok.Text)
		labels = append(labels, tok.Label)
	}
	return Entity{
		Label: strings.Split(labels[0], "-")[1],
		Text:  strings.Join(tokens, " "),
	}
}

// Encode ...
func (e *EntityExtracter) Encode(features map[string]string, label string) map[int]int {
	encoding := make(map[int]int)
	for key, val := range features {
		entry := strings.Join([]string{key, val, label}, "-")
		if _, found := e.classifier.Mapping[entry]; found {
			encoding[e.classifier.Mapping[entry]] = 1
		}
	}
	return encoding
}

// Classify ...
func (e *EntityExtracter) Classify(tokens []Token) []Token {
	history := []string{}
	labeled := []Token{}

	for i, tok := range tokens {
		scores := make(map[string]float64)
		features := extract(i, tokens, history, e.classifier.Words)
		//fmt.Println("Looking", features)
		for _, label := range e.classifier.Labels {
			total := 0.0
			for id, val := range e.Encode(features, label) {
				total += e.classifier.Weights[id] * float64(val)
			}
			scores[label] = total
		}
		label := max(scores)
		labeled = append(labeled, Token{tok.Tag, tok.Text, label})
		history = append(history, simplePOS(label))
	}

	return labeled
}

func extract(i int, ctx []Token, history, vocab []string) map[string]string {
	feats := make(map[string]string)

	word := ctx[i].Text
	prevShape := "None"

	feats["bias"] = "True"
	feats["word"] = word
	feats["pos"] = ctx[i].Tag
	if stringInSlice(word, vocab) {
		feats["en-wordlist"] = "True"
	} else {
		feats["en-wordlist"] = "False"
	}
	feats["word.lower"] = strings.ToLower(word)
	feats["suffix3"] = strings.ToLower(word[len(word)-min(len(word), 3):])
	feats["prefix3"] = strings.ToLower(word[:min(len(word), 3)])
	feats["shape"] = shape(word)
	feats["wordlen"] = strconv.Itoa(len(word))

	if i == 0 {
		feats["prevtag"] = "None"
		feats["prevword"], feats["prevpos"] = "None", "None"
	} else if i == 1 {
		feats["prevword"] = strings.ToLower(ctx[i-1].Text)
		feats["prevpos"] = ctx[i-1].Tag
		feats["prevtag"] = history[i-1]
	} else {
		feats["prevword"] = strings.ToLower(ctx[i-1].Text)
		feats["prevpos"] = ctx[i-1].Tag
		feats["prevtag"] = history[i-1]
		prevShape = shape(ctx[i-1].Text)
	}

	if i == len(ctx)-1 {
		feats["nextword"], feats["nextpos"] = "None", "None"
	} else {
		feats["nextword"] = strings.ToLower(ctx[i+1].Text)
		feats["nextpos"] = strings.ToLower(ctx[i+1].Tag)
	}

	feats["word+nextpos"] = strings.Join(
		[]string{feats["word.lower"], feats["nextpos"]}, "+")
	feats["pos+prevtag"] = strings.Join(
		[]string{feats["pos"], feats["prevtag"]}, "+")
	feats["shape+prevtag"] = strings.Join(
		[]string{prevShape, feats["prevtag"]}, "+")

	return feats
}

func shape(word string) string {
	if isNumeric(word) {
		return "number"
	} else if match, _ := regexp.MatchString(`\W+$`, word); match {
		return "punct"
	} else if match, _ := regexp.MatchString(`\w+$`, word); match {
		if strings.ToLower(word) == word {
			return "downcase"
		} else if strings.Title(word) == word {
			return "upcase"
		} else {
			return "mixedcase"
		}
	}
	return "other"
}

func simplePOS(pos string) string {
	if strings.HasPrefix(pos, "V") {
		return "v"
	}
	return strings.Split(pos, "-")[0]
}
