import os

import turing

TEST_ORDER = [
    "test_create_router",
    "test_update_router_invalid_config",
    "test_create_router_version",
    "test_undeploy_router",
    "test_deploy_router_invalid_config",
    "test_deploy_router_valid_config",
    "test_delete_router",
    "test_deploy_router_with_traffic_rules",
]


def pytest_collection_modifyitems(items):
    # modify test items in place to ensure that the e2e tests run in order
    item_mappings = {item.originalname: item for item in items}
    sorted_items = []
    for test_name in TEST_ORDER:
        sorted_items.append(item_mappings[test_name])
    items[:] = sorted_items


def pytest_sessionstart():
    turing.set_url(os.getenv("API_BASE_PATH"), use_google_oauth=False)
    turing.set_project(os.getenv("PROJECT_NAME"))
