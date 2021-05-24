import turing.session
from turing.version import VERSION as __version__

set_url = turing.session.set_url
set_project = turing.session.set_project

__all__ = ["set_url", "set_project"]
