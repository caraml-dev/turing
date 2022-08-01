import turing


def test_update_router_invalid_config():
    # get existing router that has been created in 01_create_router_test.py
    retrieved_router = turing.Router.get(1)
    assert retrieved_router is not None
