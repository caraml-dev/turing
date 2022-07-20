import turing
import turing.batch
import turing.batch.config
import turing.router.config.router_config


API_BASE_PATH = "http://turing-gateway.127.0.0.1.nip.io/api/turing/"
PROJECT_NAME = "default"


def setup_module():
    turing.set_url(API_BASE_PATH)
    turing.set_project(PROJECT_NAME)


def test_get_routers():
    routers = turing.Router.list()
    assert len(routers) == 0
