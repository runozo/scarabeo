"""Finds most score value words."""
WORDFILE = 'dicts/italia-1a'
import random
from engine.solver import Solver

SCRABBLE_GLOBAL_CRATE = 'aaaaaaaaaaaabbbbcccccccddddeeeeeeeeeeee' \
                        'ffffgggghhiiiiiiiiiiillllllmmmmmmnnnnnnoo' \
                        'ooooooooooppppqqrrrrrrrssssssstttttttuuuuvvvvzz'

if __name__ == '__main__':
    solver = Solver(WORDFILE)
    CROSSING_LETTERS = ''
    caret = SCRABBLE_GLOBAL_CRATE
    print("CARET='%s'" % caret)
    print("CROSSING_LETTERS='%s'" % CROSSING_LETTERS)
    found_words = solver.find_words(caret, CROSSING_LETTERS)
    print(found_words)[:20]
