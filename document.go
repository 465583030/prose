package prose

import (
	"path/filepath"
)

// A Component represents a stage in the Document pipeline.
type Component func(*Pipline)

// Pipline controls the Document creation process:
type Pipline struct {
	Extract  bool
	Model    string
	Tag      bool
	Tokenize bool
}

// WithTokenization can enable (the default) or disable tokenization.
func WithTokenization(include bool) Component {
	return func(pipe *Pipline) {
		// Tagging and entity extraction both require tokenization.
		pipe.Tokenize = include || pipe.Tag || pipe.Extract
	}
}

// WithTagging can enable (the default) or disable POS tagging.
func WithTagging(include bool) Component {
	return func(pipe *Pipline) {
		pipe.Tag = include
	}
}

// WithNER can enable (the default) or disable named-entity extraction.
func WithNER(include bool) Component {
	return func(pipe *Pipline) {
		pipe.Extract = include
	}
}

// UsingModel ...
//
// en-v2.0.0
//
// doc.Model.Tagger.Train(), doc.Model.Marshal("name")
func UsingModel(path string) Component {
	return func(pipe *Pipline) {
		// load model from disk ...
		pipe.Model = path
	}
}

// A Document represents a parsed body of text.
type Document struct {
	Entities  []Entity
	Model     *Model
	Sentences []Sentence
	Text      string
	Tokens    []Token
}

var defaultPipeline = Pipline{
	Tokenize: true,
	Tag:      true,
	Extract:  true,
}

// NewDocument creates a Document according to the user-specified pipeline.
func NewDocument(text string, pipeline ...Component) (*Document, error) {
	var pipeError error

	base := defaultPipeline
	for _, applyComponent := range pipeline {
		applyComponent(&base)
	}

	segmenter := NewPunktSentenceTokenizer()
	doc := Document{
		Text:      text,
		Sentences: segmenter.Segment(text),
		Model:     &Model{},
	}

	if base.Tokenize {
		tokenizer := NewTreebankWordTokenizer()
		for _, sent := range doc.Sentences {
			doc.Tokens = append(doc.Tokens, tokenizer.Tokenize(sent.Text)...)
		}
	}
	if base.Tag {
		var tagger *PerceptronTagger
		if base.Model != "" {
			tagger = loadTagger(base.Model)
		} else {
			tagger = NewPerceptronTagger()
		}
		doc.Model.Tagger = tagger
		doc.Tokens = tagger.Tag(doc.Tokens)
	}
	if base.Extract {
		var classifier *EntityExtracter
		if base.Model != "" {
			classifier = loadClassifier(base.Model)
		} else {
			classifier = NewEntityExtracter()
		}
		doc.Model.Classifier = classifier
		doc.Tokens = classifier.Classify(doc.Tokens)
		doc.Entities = classifier.Chunk(doc.Tokens)
	}

	return &doc, pipeError
}

func loadTagger(path string) *PerceptronTagger {
	var wts map[string]map[string]float64
	var tags map[string]string
	var classes []string

	loc := filepath.Join(path, "AveragedPerceptron")
	dec := getDiskAsset(filepath.Join(loc, "weights.gob"))
	checkError(dec.Decode(wts))

	dec = getDiskAsset(filepath.Join(loc, "tags.gob"))
	checkError(dec.Decode(tags))

	dec = getDiskAsset(filepath.Join(loc, "classes.gob"))
	checkError(dec.Decode(classes))

	model := NewAveragedPerceptron(wts, tags, classes)
	return NewTrainedPerceptronTagger(model)
}

func loadClassifier(path string) *EntityExtracter {
	var mapping map[string]int
	var weights []float64
	var labels []string
	var words []string

	loc := filepath.Join(path, "Maxent")
	dec := getDiskAsset(filepath.Join(loc, "mapping.gob"))
	checkError(dec.Decode(&mapping))

	dec = getDiskAsset(filepath.Join(loc, "weights.gob"))
	checkError(dec.Decode(&weights))

	dec = getDiskAsset(filepath.Join(loc, "words.gob"))
	checkError(dec.Decode(&words))

	dec = getDiskAsset(filepath.Join(loc, "labels.gob"))
	checkError(dec.Decode(&labels))

	model := NewMaxentClassifier(weights, mapping, labels, words)
	return NewTrainedEntityExtracter(model)
}
