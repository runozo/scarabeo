"""Scarabeo solver."""
WORDFILE = 'dicts/italia-1a'
import random
from engine.solver import Solver

SCRABBLE_GLOBAL_CRATE = 'aaaaaaaaaaaabbbbcccccccddddeeeeeeeeeeee' \
                        'ffffgggghhiiiiiiiiiiillllllmmmmmmnnnnnnoo' \
                        'ooooooooooppppqqrrrrrrrssssssstttttttuuuuvvvvzz'

if __name__ == '__main__':
    global_crate = list(SCRABBLE_GLOBAL_CRATE)
    solver = Solver(WORDFILE)
    for _ in range(len(SCRABBLE_GLOBAL_CRATE)):
        caret = []
        for _ in range(8):
            if len(global_crate) <= 0:
                exit()
            choice = random.choice(global_crate)
            global_crate.pop(global_crate.index(choice))
            caret.append(choice)
        CROSSING_LETTERS = ''
        caret = str(caret)
        print("CARET='%s'" % caret)
        print("CROSSING_LETTERS='%s'" % CROSSING_LETTERS)
        found_words = solver.find_words(caret, CROSSING_LETTERS)
        print(found_words)
        print(len(found_words))
