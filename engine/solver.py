"""The scrabble solver engine"""
import re
from collections import Counter


class Solver(object):
    """You need to istantiate one of this to make things work. ;)"""

    def __init__(self, wordfile):
        """
        Args:
            wordfile: a text file with a list of
            (possibly) all the words in a language.
        """
        self.wordfile = wordfile

    def find_words(self, caret, crossing_letters=None):
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
        all_words = []
        regexp = re.compile(".*" +
                            crossing_letters.replace(' ', '[a-z]') +
                            ".*") if crossing_letters else re.compile(".*")
        caret = caret + \
            crossing_letters.replace(' ', '') if crossing_letters else caret
        c_caret = Counter(caret)

        with open(self.wordfile) as words:
            for word in words:
                word = word.strip()

                if len(word) <= len(caret) and regexp.match(word):
                    if len(word) > 0:
                        c_word = Counter(word)
                        if all(char in caret and c_caret[char] >= c_word[char]
                                for char in word):
                            all_words.append(word)
        return sorted(all_words, key=str.__len__, reverse=True)
