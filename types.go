package prose

import (
	"os"
)

// Token ...
type Token struct {
	Tag   string
	Text  string
	Label string
}

// Entity ...
type Entity struct {
	Text  string
	Label string
}

// Sentence ...
type Sentence struct {
	Text   string // the actual text
	Length int    // the number of words
}

// Model ...
type Model struct {
	Tagger     *PerceptronTagger
	Classifier *EntityExtracter
}

// Marshal ...
func (m *Model) Marshal(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	m.Tagger.model.Marshal(path)
	m.Classifier.model.Marshal(path)
	return err
}
