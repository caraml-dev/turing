import setuptools
import pathlib
import pkg_resources

# get version from version.py
with pathlib.Path("version.py").open() as version_py:
    _locals = locals()
    exec(version_py.read(), globals(), _locals)
    version = _locals["VERSION"]

with pathlib.Path("requirements.txt").open() as requirements_txt:
    requirements = [
        str(requirement)
        for requirement in pkg_resources.parse_requirements(requirements_txt)
    ]

with pathlib.Path("requirements.dev.txt").open() as dev_requirements_test:
    dev_requirements = [
        str(requirement)
        for requirement in pkg_resources.parse_requirements(dev_requirements_test)
    ]

setuptools.setup(
    name="turing-pyfunc-ensembler-job",
    version=version,
    packages=setuptools.find_packages(),
    install_requires=requirements,
    extras_require={
        "dev": dev_requirements,
    },
    python_requires=">=3.8,<3.11",
    setup_requires=["setuptools_scm"],
)
