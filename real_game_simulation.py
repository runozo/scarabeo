"""A Scrabble game simulation between 2 Solver instances."""
import random
from engine.solver import Solver

WORDFILE = 'dicts/italia-1a'
SCRABBLE_GLOBAL_CRATE = 'aaaaaaaaaaaabbbbcccccccddddeeeeeeeeeeee' \
                        'ffffgggghhiiiiiiiiiiillllllmmmmmmnnnnnnoo' \
                        'ooooooooooppppqqrrrrrrrssssssstttttttuuuuvvvvzz'


class Player(object):
    """A class representing a Scrabble player."""
    def __init__(self, name):
        self.name = name
        self.crate = []
        self.words = []
        self.score = 0
        self.solver = Solver(WORDFILE)

    def pick_from_crate(self, game_crate):
        """Fill the caret with new chars."""
        picked = []
        for _ in range(8 - len(self.crate)):
            if len(game_crate) > 0:
                choice = random.choice(game_crate)
                game_crate.pop(game_crate.index(choice))
                picked.append(choice)
        self.crate = self.crate + picked
        return game_crate

    def play_one_move(self):
        """Write the most valuable word on the gametable."""
        try:
            found_word = self.solver.find_words(self.crate, "")[0]
            self.words.append(found_word)
            for char in found_word[0]:
                self.crate.pop(self.crate.index(char))
            self.score = sum([x[1] for x in self.words])
            return found_word
        except IndexError:
            return False


if __name__ == '__main__':
    print("A game simulation")
    global_crate = list(SCRABBLE_GLOBAL_CRATE)
    players = (Player("King Kong"),Player("Hulk"))  # 2 brand new players
    play, count = True, 0
    while play:
        print("Turn %s" % count)
        for player in players:
            global_crate = player.pick_from_crate(global_crate)
            print("%s crate: %s, len: %s, score: %s, words: %s" %
                  (player.name, player.crate, len(player.crate), player.score, player.words))
            play = player.play_one_move()
            print("%s plays: %s" % (player.name, play))
            print("%s crate: %s, len: %s, score: %s, words: %s" %
                  (player.name, player.crate, len(player.crate), player.score, player.words))
        print("Global crate: %s" % global_crate)
        count += 1
