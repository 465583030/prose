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

	dec := getJSONAsset("Maxent", "mapping.json")
	checkError(dec.Decode(&mapping))

	dec = getJSONAsset("Maxent", "weights.json")
	checkError(dec.Decode(&weights))

	dec = getJSONAsset("Maxent", "words.json")
	checkError(dec.Decode(&words))

	dec = getJSONAsset("Maxent", "labels.json")
	checkError(dec.Decode(&labels))

	return &EntityExtracter{classifier: NewMaxentClassifier(weights, mapping, labels, words)}
}

// Encode ...
func (e *EntityExtracter) Encode(features map[string]string, label string) map[int]int {
	encoding := make(map[int]int)
	for key, val := range features {
		entry := strings.Join([]string{key, val, label}, "-")
		if _, found := e.classifier.Mapping[entry]; found {
			//fmt.Println("found", e.classifier.Mapping[entry])
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

// quadString creates a string containing all of the tags, each padded to 4
// characters wide.
func quadsString(tagged []Token) string {
	tagQuads := ""
	for _, tok := range tagged {
		padding := ""
		pos := tok.Tag
		switch len(pos) {
		case 0:
			padding = "____" // should not exist
		case 1:
			padding = "___"
		case 2:
			padding = "__"
		case 3:
			padding = "_"
		case 4: // no padding required
		default:
			pos = pos[:4] // longer than 4 ... truncate!
		}
		tagQuads += pos + padding
	}

	return tagQuads
}

// TreebankNamedEntities matches proper names, excluding prior adjectives,
// possibly including numbers and a linkage by preposition or subordinating
// conjunctions (for example "Bank of England").
var TreebankNamedEntities = regexp.MustCompile(
	`((CD__)*(NNP.)+(CD__|NNP.)*)+` +
		`((IN__)*(CD__)*(NNP.)+(CD__|NNP.)*)*`)

// Chunk returns a slice containing the chunks of interest according to the
// regexp.
//
// This is a convenience wrapper around Locate, which should be used if you
// need access the to the in-text locations of each chunk.
func Chunk(tagged []Token, rx *regexp.Regexp) []string {
	chunks := []string{}
	for _, loc := range Locate(tagged, rx) {
		res := ""
		for t, tt := range tagged[loc[0]:loc[1]] {
			if t != 0 {
				res += " "
			}
			res += tt.Text
		}
		chunks = append(chunks, res)
	}
	return chunks
}

// Locate finds the chunks of interest according to the regexp.
func Locate(tagged []Token, rx *regexp.Regexp) [][]int {
	rx.Longest() // make sure we find the longest possible sequences
	rs := rx.FindAllStringIndex(quadsString(tagged), -1)
	for i, ii := range rs {
		for j := range ii {
			// quadsString makes every offset 4x what it should be
			rs[i][j] /= 4
		}
	}
	return rs
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
