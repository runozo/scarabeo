package engine

// Alphabet size used throughout the engine. Only the 21 letters of the
// Italian alphabet are worth points; j, k, w, x and y are kept at 0 so that
// loanwords loaded from a dirty dictionary never crash the scorer.
const alphabetSize = 26

// Jolly is the wildcard tile accepted in a rack. It stands for any letter and
// is scored as the letter it substitutes.
const Jolly = '?'

// letterPoints maps a letter (index = letter-'a') to its Scarabeo value.
var letterPoints = [alphabetSize]int{
	// a  b  c  d  e  f  g  h  i  j  k  l  m
	1, 4, 1, 4, 1, 4, 4, 8, 1, 0, 0, 2, 2,
	// n  o  p  q  r  s  t  u  v  w  x  y  z
	2, 1, 3, 10, 1, 1, 1, 4, 4, 0, 0, 0, 8,
}

// letterValue returns the score of a single lowercase letter, or 0 when the
// letter is unknown (foreign letters, punctuation, ...).
func letterValue(ch byte) int {
	if ch < 'a' || ch > 'z' {
		return 0
	}
	return letterPoints[ch-'a']
}

// countLetters fills a letter histogram for an already lowercase, a-z word.
func countLetters(word string) (counts [alphabetSize]uint8) {
	for i := 0; i < len(word); i++ {
		if c := word[i]; c >= 'a' && c <= 'z' {
			counts[c-'a']++
		}
	}
	return counts
}
