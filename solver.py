"""Scarabeo solver."""
WORDFILE = 'dicts/italia-1a'
import sys
import random
from engine.solver import Solver

SCRABBLE_GLOBAL_CRATE = 'aaaaaaaaaaaabbbbcccccccddddeeeeeeeeeeee' \
                        'ffffgggghhiiiiiiiiiiillllllmmmmmmnnnnnnoo' \
                        'ooooooooooppppqqrrrrrrrssssssstttttttuuuuvvvvzz'

if __name__ == '__main__':
    solver = Solver(WORDFILE)
    print(sys.argv[1], sys.argv[2])
    found_words = solver.find_words(sys.argv[1], sys.argv[2])
    print("CARET LEN: %s" % len(sys.argv[1]))
    print(found_words)
