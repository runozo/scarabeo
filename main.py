"""Scarabeo solver."""
WORDFILE = 'dicts/italia-1a'
import random
from engine.solver import Solver

SCRABBLE_GLOBAL_CRATE = 'aaaaaaaaaaaabbbbcccccccddddeeeeeeeeeeee' \
                        'ffffgggghhiiiiiiiiiiillllllmmmmmmnnnnnnoo' \
                        'ooooooooooppppqqrrrrrrrssssssstttttttuuuuvvvvzz'

if __name__ == '__main__':
    print("A game simulation")
    global_crate = list(SCRABBLE_GLOBAL_CRATE)
    solver = Solver(WORDFILE)
    player1_words, player2_words = [], []
    for i in range(len(SCRABBLE_GLOBAL_CRATE)):
        caret = []
        for _ in range(8):
            if len(global_crate) <= 0:
                exit()
            choice = random.choice(global_crate)
            global_crate.pop(global_crate.index(choice))
            caret.append(choice)
        CROSSING_LETTERS = ''
        if i % 2 == 0:
            print("Player 1 CARET='%s'" % "".join(caret))
            player1_words.append(solver.find_words(
                "".join(caret), CROSSING_LETTERS)[0])
        else:
            print("Player 2 CARET='%s'" % "".join(caret))
            player2_words.append(solver.find_words(
                "".join(caret), CROSSING_LETTERS)[0])
        if player1_words:
            print("Player 1 word: %s" % player1_words[-1][0])
            print("Player 1 total points: %s" %
                  sum([x[1] for x in player1_words]))
        if player2_words:
            print("Player 2 word: %s" % player2_words[-1][0])
            print("Player 2 total points: %s" %
                  sum([x[1] for x in player2_words]))
        print("Player 1 total maccarone: %s" % player1_words)
        print("Player 2 total maccarone: %s" % player2_words)

    print("Player 1 total maccarone: %s" % player1_words)
    print("Player 2 total maccarone: %s" % player2_words)
