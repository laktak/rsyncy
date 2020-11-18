from setuptools import setup
import os

with open(os.path.join(os.path.dirname(__file__), "README.md"), encoding="utf-8") as f:
    readme = f.read()

setup(
    name="rsyncy",
    version="0.0.4",
    url="https://github.com/laktak/rsyncy",
    author="Christian Zangl",
    author_email="laktak@cdak.net",
    description="A status/progress bar for rsync.",
    long_description=readme,
    long_description_content_type="text/markdown",
    packages=[],
    install_requires=[],
    scripts=["rsyncy"],
    py_modules=["rsyncy"],
)
