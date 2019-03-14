# Copyright (C) 2019 Hatching B.V.
# All rights reserved.

from helpers import settings

class FuncTest(object):

    name = ""
    description = ""
    order = 1
    stop_on_fail = False

    def __init__(self):
        self.settings = settings
        self.failreason = ""
        self.failfix = ""
        self._finished = False
        self._passed = False

    @property
    def finished(self):
        return self._finished

    @property
    def passed(self):
        return self._passed

    def markfail(self, reason, fix=""):
        self.failreason = reason
        self.failfix = fix
        self._finished = True
        self._passed = False
        return False

    def markpass(self):
        self._finished = True
        self._passed = True
        return True

    def init(self):
        pass

    def start(self):
        pass

    def cancel(self):
        pass

    def cleanup(self):
        pass
