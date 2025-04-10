import setuptools
import pathlib
import pkg_resources
import importlib.util
import os

# get version from version.py
spec = importlib.util.spec_from_file_location(
    "pyfuncserver.version", os.path.join("version.py")
)

v_module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(v_module)

version = v_module.VERSION

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
    name="turing-pyfunc-ensembler-service",
    version=version,
    packages=setuptools.find_packages(),
    install_requires=requirements,
    extras_require={
        "dev": dev_requirements,
    },
    python_requires=">=3.8,<3.11",
    setup_requires=["setuptools_scm"],
)
