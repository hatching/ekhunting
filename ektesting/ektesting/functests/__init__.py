# Copyright (C) 2019 Hatching B.V.
# All rights reserved.

from ektesting.abstracts import FuncTest
from ektesting.helpers import enumerate_plugins

tests = enumerate_plugins(
    __file__, "ektesting.functests", globals(), FuncTest, as_dict=True
)

