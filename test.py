"""Not a real test."""
WORDFILE = 'dicts/italia-1a'
from engine.solver import Solver


if __name__ == '__main__':
    CARET = 'scarabeo'
    RESTRICTIONS = 'a e'
    print("CARET='%s'" % CARET)
    print("RESTRICTIONS='%s'" % RESTRICTIONS)
    solver = Solver(WORDFILE)
    found_words = solver.find_words(CARET, RESTRICTIONS)
    print(found_words)
    print(len(found_words))
