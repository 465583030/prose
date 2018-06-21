package prose

// Component ...
type Component func(*Pipline)

// Pipline ...
type Pipline struct {
	Summarize bool
	Tag       bool
	Tokenize  bool
}

// WithTokenization ...
func WithTokenization(include bool) Component {
	return func(pipe *Pipline) {
		// Tagging and summarization both require tokenization.
		pipe.Tokenize = include || pipe.Tag || pipe.Summarize
	}
}

// WithTagging ...
func WithTagging(include bool) Component {
	return func(pipe *Pipline) {
		pipe.Tag = include
	}
}

// WithSummarization ...
func WithSummarization(include bool) Component {
	return func(pipe *Pipline) {
		pipe.Summarize = include
	}
}

// Document ...
type Document struct {
	Text      string
	Tokens    []Token
	Sentences []Sentence
}

var defaultPipeline = Pipline{
	Tokenize:  true,
	Tag:       true,
	Summarize: false,
}

// NewDocument ...
func NewDocument(text string, pipeline ...Component) (*Document, error) {
	var pipeError error

	base := defaultPipeline
	for _, applyComponent := range pipeline {
		applyComponent(&base)
	}

	segmenter := NewPunktSentenceTokenizer()
	doc := Document{Text: text, Sentences: segmenter.Segment(text)}
	if base.Tokenize {
		tokenizer := NewTreebankWordTokenizer()
		for _, sent := range doc.Sentences {
			doc.Tokens = append(doc.Tokens, tokenizer.Tokenize(sent.Text)...)
		}
	}
	if base.Tag {
		tagger := NewPerceptronTagger()
		doc.Tokens = tagger.Tag(doc.Tokens)
	}

	return &doc, pipeError
}

// MatchString ...
func (doc *Document) MatchString(query string) []string {
	return []string{}
}

// People ...
func (doc *Document) People(query string) []string {
	return []string{}
}

// Places ...
func (doc *Document) Places(query string) []string {
	return []string{}
}

// Organizations ...
func (doc *Document) Organizations(query string) []string {
	return []string{}
}
