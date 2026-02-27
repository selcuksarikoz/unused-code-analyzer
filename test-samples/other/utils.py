def used_function():
    return "hello"


USED_CONST = "world"


class UsedClass:
    def __init__(self, name):
        self.name = name


def unused_function():
    return "unused"


UNUSED_CONST = 123


class UnusedClass:
    pass


def func_with_unused_param(a, b, unused):
    print(a, b)


def func_with_unused_return():
    return 1
