"""The scrabble solver engine"""
import re
from collections import Counter

POINTS = {'a': 1,
          'b': 4,
          'c': 1,
          'd': 4,
          'e': 1,
          'f': 4,
          'g': 4,
          'h': 8,
          'i': 1,
          'l': 2,
          'm': 2,
          'n': 2,
          'o': 1,
          'p': 3,
          'q': 10,
          'r': 1,
          's': 1,
          't': 1,
          'u': 4,
          'v': 4,
          'z': 8
          }


class Solver(object):
    """You need to istantiate one of this to make things work. ;)"""

    def __init__(self, wordfile):
        """
        Args:
            wordfile: a text file with a list of
            (possibly) all the words in a language.
        """
        self.wordfile = wordfile
        self.caret = ''
        self.crossing_letters = ''

    def calculate_score(self, word):
        total_sum = 0
        for char in word:
            total_sum += POINTS.get(char)
        chars_from_caret = len(
            word) - len(self.crossing_letters.replace(' ', ''))
        if chars_from_caret == 6:
            total_sum += 10
        elif chars_from_caret == 7:
            total_sum += 30
        elif chars_from_caret == 8:
            total_sum += 50
        # print(word, total_sum)
        return total_sum

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
        self.caret = caret
        self.crossing_letters = crossing_letters
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
        return [(x, self.calculate_score(x)) for x in sorted(all_words, key=self.calculate_score, reverse=True)]
