package prose

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io/ioutil"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// readDataFile reads data from a file, panicking on any errors.
func readDataFile(path string) []byte {
	p, err := filepath.Abs(path)
	checkError(err)

	data, ferr := ioutil.ReadFile(p)
	checkError(ferr)

	return data
}

// checkError panics if `err` is not `nil`.
func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

// min returns the minimum of `a` and `b`.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isPunct determines if the string represents a number.
func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// isPunct determines if a character is a punctuation symbol.
func isPunct(c byte) bool {
	for _, r := range []byte("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~") {
		if c == r {
			return true
		}
	}
	return false
}

// isSpace determines if a character is a whitespace character.
func isSpace(c byte) bool {
	for _, r := range []byte("\t\n\r\f\v") {
		if c == r {
			return true
		}
	}
	return false
}

// isLetter determines if a character is letter.
func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isAlnum determines if a character is a letter or a digit.
func isAlnum(c byte) bool {
	return (c >= '0' && c <= '9') || isLetter(c)
}

// stringInSlice determines if `slice` contains the string `a`.
func stringInSlice(a string, slice []string) bool {
	for _, b := range slice {
		if a == b {
			return true
		}
	}
	return false
}

// hasAnySuffix determines if the string a has any suffixes contained in the
// slice b.
func hasAnySuffix(a string, slice []string) bool {
	for _, b := range slice {
		if strings.HasSuffix(a, b) {
			return true
		}
	}
	return false
}

// containsAny determines if the string a contains any fo the strings contained
// in the slice b.
func containsAny(a string, b []string) bool {
	for _, s := range b {
		if strings.Contains(a, s) {
			return true
		}
	}
	return false
}

// getAsset returns the named Asset.
func getJSONAsset(folder, name string) *json.Decoder {
	b, err := Asset(path.Join("data", folder, name))
	checkError(err)
	return json.NewDecoder(bytes.NewReader(b))
}

func getGobAsset(folder, name string) *gob.Decoder {
	b, err := Asset(path.Join("data", folder, name))
	checkError(err)
	return gob.NewDecoder(bytes.NewReader(b))
}
