package prose

// Token ...
type Token interface {
}

// Tokenizer is the interface implemented by an object that takes a string
// and returns a slice of substrings.
type Tokenizer interface {
	Tokenize(text string) []Token
}
