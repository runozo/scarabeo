"""A game simulation between 2 solvers."""
WORDFILE = 'dicts/italia-1a'
import random
from engine.solver import Solver

SCRABBLE_GLOBAL_CRATE = 'aaaaaaaaaaaabbbbcccccccddddeeeeeeeeeeee' \
                        'ffffgggghhiiiiiiiiiiillllllmmmmmmnnnnnnoo' \
                        'ooooooooooppppqqrrrrrrrssssssstttttttuuuuvvvvzz'


class Player:

    def __init__(self, name):
        self.name = name
        self.crate = []
        self.words = []
        self.score = 0
        self.solver = Solver(WORDFILE)

    def pick_from_crate(self, global_crate):
        picked = []
        for _ in range(8 - len(self.crate)):
            if len(global_crate) > 0:
                choice = random.choice(global_crate)
                global_crate.pop(global_crate.index(choice))
                picked.append(choice)
        self.crate = self.crate + picked
        return global_crate

    def play_one_move(self):
        try:
            found_word = self.solver.find_words(self.crate, "")[0]
            self.words.append(found_word)
            for char in found_word[0]:
                self.crate.pop(self.crate.index(char))
            self.score = sum([x[1] for x in self.words])
            return True
        except IndexError as exp:
            return False


if __name__ == '__main__':
    print("A game simulation")
    global_crate = list(SCRABBLE_GLOBAL_CRATE)
    player1 = Player("King Kong")
    player2 = Player("Hulk")
    play, count = True, 0
    while play:
        print("Turn %s" % str(count / 2))

        if count % 2 == 0:
            global_crate = player1.pick_from_crate(global_crate)
            print("%s crate: %s, len: %s, score: %s, words: %s" %
                  (player1.name, player1.crate, len(player1.crate), player1.score, player1.words))
            play = player1.play_one_move()
        else:
            global_crate = player2.pick_from_crate(global_crate)
            print("%s crate: %s, len: %s, score: %s, words: %s" %
                  (player2.name, player2.crate, len(player2.crate), player2.score, player2.words))
            play = player2.play_one_move()
        print("Global crate: %s" % global_crate)
        count += 1
