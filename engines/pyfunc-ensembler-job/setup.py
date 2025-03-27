import setuptools
import pathlib
import pkg_resources
import imp
import os

version = imp.load_source(
    "pyfuncserver.version", os.path.join("version.py")
).VERSION

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
    name="pyfunc-ensembler-job",
    version=version,
    packages=setuptools.find_packages(),
    install_requires=requirements,
    dev_requirements=dev_requirements,
    python_requires=">=3.8,<3.11",
    setup_requires=["setuptools_scm"],
)
