import setuptools
import pathlib

with pathlib.Path("turing/version.py").open() as version_py:
    _locals = locals()
    exec(version_py.read(), globals(), _locals)
    version = _locals["VERSION"]

def parse_requirements(path):
    """Requirement lines of a requirements file, without comments and blank lines."""
    lines = (line.split(" #")[0].strip() for line in pathlib.Path(path).read_text().splitlines())
    return [line for line in lines if line and not line.startswith(("#", "-"))]


requirements = parse_requirements("requirements.txt")
dev_requirements = parse_requirements("requirements.dev.txt")

setuptools.setup(
    name="turing-sdk",
    version=version,
    packages=setuptools.find_packages(exclude=["docs", "e2e", "samples", "tests"]),
    install_requires=requirements,
    extras_require={"dev": dev_requirements},
    python_requires=">=3.8,<3.11",
    long_description=pathlib.Path("./docs/README.md").read_text(),
    long_description_content_type="text/markdown",
)
