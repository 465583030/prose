package prose

// Token ...
type Token struct {
	Text string
	Tag  string
}

// Sentence ...
type Sentence struct {
	Text      string // the actual text
	Length    int    // the number of words
	Paragraph int
}

// A RankedParagraph is a paragraph ranked by its number of keywords.
type RankedParagraph struct {
	Sentences []Sentence
	Position  int // the zero-based position within a Document
	Rank      int
}
