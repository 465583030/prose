package prose

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
	Text      string
	Tokens    []Token
	Sentences []Sentence
	Entities  []Entity
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

	if base.Model != "" {
		// We've found a custom model.
	} else {
		// Load the default model.
	}

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
		doc.Entities = classifier.Chunk(doc.Tokens)
	}

	return &doc, pipeError
}
