package prose

// Token ...
type Token struct {
	Tag  string
	Text string
}

// Sentence ...
type Sentence struct {
	Text      string // the actual text
	Length    int    // the number of words
	Paragraph int
}
