package prose

// A Component represents a stage in the Document pipeline.
type Component func(*Pipline)

// Pipline controls the Document creation process:
type Pipline struct {
	EntityMap map[string]string
	Extract   bool
	Tag       bool
	Tokenize  bool
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

// WithExtraction can enable (the default) or disable named-entity extraction.
func WithExtraction(include bool) Component {
	return func(pipe *Pipline) {
		pipe.Extract = include
	}
}

// A Document represents a parsed body of text.
type Document struct {
	Text      string
	Tokens    []Token
	Sentences []Sentence
	Entities  []string // TODO: Use more fine-grained labels -- e.g., ORG, etc.
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
	if base.Extract {
		classifier := NewEntityExtracter()
		doc.Tokens = classifier.Classify(doc.Tokens)
	}

	return &doc, pipeError
}
