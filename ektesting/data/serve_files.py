#!/usr/bin/env python2

# Copyright (C) 2019 Hatching B.V.
# All rights reserved.

import argparse
import sys
import os

try:
    from flask import Flask, request, send_from_directory, abort
except ImportError:
    sys.stderr.write(
        "Flask is required to run this server. Run 'pip install Flask'\n"
    )
    sys.exit(1)

app = Flask("Flask EK Testing", static_url_path="/files")

@app.route("/<path>")
def serve_page(path):
    if os.path.isfile(os.path.join("files", path)):
        return send_from_directory("files", path)
    abort(404)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Run development server to serve page to the EK hunting "
                    "functional testing tasks"
    )
    parser.add_argument("--host", nargs="?", default="127.0.0.1")
    parser.add_argument("--port", nargs="?", default="8090")
    args = parser.parse_args()

    app.run(host=args.host, port=int(args.port))

