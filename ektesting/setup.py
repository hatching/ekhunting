# Copyright (C) 2019 Hatching B.V.
# All rights reserved.

import setuptools

setuptools.setup(
    name="ektesting",
    version="0.1",
    author="Hatching International B.V",
    description="A functional testing package for Cuckoo MassURL analysis",
    author_email="info@hatching.io",
    packages=[
        "ektesting",
    ],
    include_package_data=True,
    classifiers=[
        "Environment :: Console",
        "Natural Language :: English",
        "Programming Language :: Python :: 2.7",
    ],
    entry_points={
        "console_scripts": [
            "ektest = ektesting.main:main",
        ]
    },
    install_requires=[
        "click>=6.6",
        "cuckoo>=2.0.7a1"
    ]
)
