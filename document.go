package prose

// Component ...
type Component func(*Pipline)

// Pipline ...
type Pipline struct {
	Segment   bool
	Summarize bool
	Tag       bool
	Tokenize  bool
}

// WithTokenization ...
func WithTokenization(include bool) Component {
	return func(pipe *Pipline) {
		pipe.Tokenize = include
	}
}

// WithSegmentation ...
func WithSegmentation(include bool) Component {
	return func(pipe *Pipline) {
		pipe.Segment = include
	}
}

// WithTagging ...
func WithTagging(include bool) Component {
	return func(pipe *Pipline) {
		pipe.Tag = include
	}
}

// Document ...
type Document struct {
	Text      string
	Tokens    []Token
	Sentences []Sentence
}

var defaultPipeline = Pipline{
	Tokenize: true,
	Segment:  true,
	Tag:      true,
}

// NewDocument ...
func NewDocument(text string, pipeline ...Component) (*Document, error) {
	var pipeError error

	base := defaultPipeline
	for _, applyComponent := range pipeline {
		applyComponent(&base)
	}

	doc := Document{Text: text}
	if base.Tokenize || base.Tag {
		tokenizer := NewTreebankWordTokenizer()
		doc.Tokens = tokenizer.Tokenize(text)
	}
	if base.Segment {
		segmenter := NewPunktSentenceTokenizer()
		doc.Sentences = segmenter.Segment(text)
	}
	if base.Tag {
		tagger := NewPerceptronTagger()
		doc.Tokens = tagger.Tag(doc.Tokens)
	}

	return &doc, pipeError
}
