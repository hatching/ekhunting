# Copyright (C) 2019 Hatching B.V.
# All rights reserved.

import logging

import click
from ektesting.helpers import settings, run_tests
from ektesting.functests import tests

log = logging.getLogger(__name__)

@click.group()
@click.argument("cuckoo_cwd", required=True, type=click.Path(exists=True, readable=True, writable=True))
@click.argument("eventserver_ip", required=True)
@click.argument("eventserver_port", required=True, type=click.INT)
@click.argument("ekwebserver", required=True)
@click.option("-d", "--debug", is_flag=True, help="Enable debug logging")
def main(cuckoo_cwd, eventserver_ip, eventserver_port, ekwebserver, debug):
    """Run functional tests on a Exploit kit hunting/Cuckoo MassURL setup.

    EK_WEBSERVER must be remote web server IP:PORT that serves the files
    shipped with this package. Files cannot be served from the same server
    that Cuckoo is running on, as Cuckoo will block requests from the VM to
    the host machine"""
    logging.basicConfig(
        level=logging.DEBUG if debug else logging.INFO,
        format="%(asctime)s [%(name)s] %(levelname)s: %(message)s"
    )

    try:
        from cuckoo import massurl
    except ImportError:
        log.error("Cuckoo Massurl must be installed to use this tool!")

    settings.init(
        cwd=cuckoo_cwd, event_ip=eventserver_ip, event_port=eventserver_port,
        webserver=ekwebserver
    )

@main.command()
def list():
    """List all available tests"""
    print "All tests: %s\n" % " ".join(tests.keys())
    for name, test in tests.iteritems():
        print "%s | %s " % (name, test.description)

@main.command()
@click.argument("test_names", nargs=-1, required=True)
def run(test_names):
    """Run one or more specified tests"""
    run_tests(tests=[tests.get(n) for n in test_names if n in tests])

@main.command()
def runall():
    """Run all available tests"""
    run_tests(tests=sorted(tests.values(), key=lambda x: x.order, reverse=True))
