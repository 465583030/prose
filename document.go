package prose

// Component ...
type Component func(*Pipline) error

// Pipline ...
type Pipline struct {
	Tokenizer Tokenizer
}

// UsingTokenizer ...
func UsingTokenizer(tokenizer Tokenizer) Component {
	return func(pipe *Pipline) error {
		pipe.Tokenizer = tokenizer
		return nil
	}
}

// Document ...
type Document struct {
	Text   string
	Tokens []Token
}

var defaultPipeline = Pipline{
	Tokenizer: NewTreebankWordTokenizer(),
}

// NewDocument ...
func NewDocument(text string, pipeline ...Component) (*Document, error) {
	var pipeError error

	base := defaultPipeline
	for _, applyComponent := range pipeline {
		pipeError = applyComponent(&base)
	}

	doc := Document{Text: text}
	if base.Tokenizer != nil {
		doc.Tokens = base.Tokenizer.Tokenize(text)
	}

	return &doc, pipeError
}
