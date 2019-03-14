Cuckoo Massurl/EKHunting functional testing
===========================================

This is a functional testing framework that can be used to test a Cuckoo Massurl setup.
It should be installed in the same environment as the Cuckoo Massurl installation.

In order to install and use this tool:

    pip install .
    ektest --debug CUCKOO_CWD EVENTSERVER_IP EVENTSERVER_PORT FILE_WEBSERVER runall
