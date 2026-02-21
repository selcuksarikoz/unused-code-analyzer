# main.py
from utils import used_function, USED_CONST, UsedClass
from utils import unused_function as used_alias

result = used_function()
print(result, USED_CONST)

obj = UsedClass("test")
print(obj.name)

print(used_alias())

try:
    from utils import maybe_missing
except ImportError:
    pass

import os
from os import path

print(os.getcwd())
print(path.join("a", "b"))

unused_var = "this is unused"
unused_class = "unused"


def local_func():
    pass


class LocalClass:
    pass
