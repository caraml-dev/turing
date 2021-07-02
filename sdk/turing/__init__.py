import turing.session
from turing.version import VERSION as __version__
from turing.project import Project
from turing.ensembler import Ensembler, EnsemblerType, PyFuncEnsembler
from turing.job import EnsemblingJob

set_url = turing.session.set_url
set_project = turing.session.set_project

__all__ = ["set_url", "set_project"]
