# Copyright (C) 2019 Hatching B.V.
# All rights reserved.

import string
import random
import os
import logging
import importlib
import sys

from cuckoo.misc import decide_cwd

log = logging.getLogger(__name__)

class Settings(object):
    cwd = None
    event_ip = None
    event_port = None

    def init(cls, cwd, event_ip, event_port, webserver):
        decide_cwd(cwd)
        cls.cwd = cwd
        cls.event_ip = event_ip
        cls.event_port = event_port
        cls.webserver = webserver

settings = Settings()

def rand_string(n):
    return ''.join(
        random.choice(string.ascii_uppercase + string.digits) for _ in range(n)
    )

def run_tests(tests=[]):
    passcount = 0
    for functest in tests:
        test = functest()

        passed = False
        log.info("Starting test '%s'", test.name)
        log.info("Test description: %s", test.description)
        try:
            test.init()
            passed = test.start()
        except Exception as e:
            log.exception("Exception while starting test '%s'", test.name)
            continue
        finally:
            test.cleanup()

        if passed:
            log.info("Test '%s': Passed", test.name)
            passcount += 1
        else:
            log.info("Test '%s': Failed", test.name)
            log.info("Reason: %s", test.failreason)
            log.info("Possible fix: %s", test.failfix)
            if test.stop_on_fail:
                log.warning(
                    "Cannot continue because test '%s' failed", test.name
                )
                sys.exit(1)
    log.info(
        "Tests passed: %s. Tests failed: %s", passcount, len(tests) - passcount
    )

def enumerate_plugins(dirpath, module_prefix, namespace, class_,
                      attributes={}, as_dict=False):
    """Import plugins of type `class` located at `dirpath` into the
    `namespace` that starts with `module_prefix`. If `dirpath` represents a
    filepath then it is converted into its containing directory. The
    `attributes` dictionary allows one to set extra fields for all imported
    plugins. Using `as_dict` a dictionary based on the module name is
    returned."""
    if os.path.isfile(dirpath):
        dirpath = os.path.dirname(dirpath)

    for fname in os.listdir(dirpath):
        if fname.endswith(".py") and not fname.startswith("__init__"):
            module_name, _ = os.path.splitext(fname)
            try:
                importlib.import_module(
                    "%s.%s" % (module_prefix, module_name)
                )
            except ImportError as e:
                log.error("Failed to import %s", e)

    subclasses = class_.__subclasses__()[:]

    plugins = []
    while subclasses:
        subclass = subclasses.pop(0)

        # Include subclasses of this subclass (there are some subclasses, e.g.,
        # LibVirtMachinery, that fail the fail the following module namespace
        # check and as such we perform this logic here).
        subclasses.extend(subclass.__subclasses__())

        # Check whether this subclass belongs to the module namespace that
        # we're currently importing. It should be noted that parent and child
        # namespaces should fail the following if-statement.
        if module_prefix != ".".join(subclass.__module__.split(".")[:-1]):
            continue

        namespace[subclass.__name__] = subclass
        for key, value in attributes.items():
            setattr(subclass, key, value)

        plugins.append(subclass)

    if as_dict:
        ret = {}
        for plugin in plugins:
            ret[plugin.__module__.split(".")[-1]] = plugin
        return ret

    return sorted(plugins, key=lambda x: x.__name__.lower())