# main.py
from utils import used_function, USED_CONST, UsedClass

result = used_function()
print(result, USED_CONST)

obj = UsedClass("test")
print(obj.name)

from utils import unused_function as used_alias
