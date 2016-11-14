"""Scarabeo solver."""
import re
from collections import Counter
WORDFILE = 'dicts/italia-1a'


def find_words(caret, crossing_letters=None):
    """
    Find all the words in a caret.

    Args:

    caret:
        your caret of letters
    crossing_letters:
        is a string with a pattern representing letters
        we must to cross in composing the word.
        Example: "a b" match an a, a space and a b
    """
    ret_words = []
    regexp = re.compile(".*" +
                        crossing_letters.replace(' ', '[a-z]') +
                        ".*") if crossing_letters else re.compile(".*")
    caret = caret + \
        crossing_letters.replace(' ', '') if crossing_letters else caret
    c_caret = Counter(caret)

    with open(WORDFILE) as words:
        for word in words:
            word = word.strip()

            if len(word) <= len(caret) and regexp.match(word):
                c_word = Counter(word)
                if len(word) <= len(caret):
                    if all(char in caret and c_caret[char] >= c_word[char]
                            for char in word):
                        if len(word) > 0:
                            ret_words.append(word)
    return sorted(ret_words, key=str.__len__, reverse=True)

if __name__ == '__main__':
    CARET = 'scarabeo'
    RESTRICTIONS = 'a e'
    print("CARET='%s'" % CARET)
    print("RESTRICTIONS='%s'" % RESTRICTIONS)
    found_words = find_words(CARET, RESTRICTIONS)
    print(found_words)
    print(len(found_words))
