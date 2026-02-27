from utils import used_function, unused_function, UNUSED_CONST
from utils import UsedClass


class Feature:
    def __init__(self):
        self.value = used_function()

    def process(self):
        return self.value


obj = UsedClass("test")

unused_imported = "this is not used"


def local_function():
    pass


class LocalClass:
    pass
