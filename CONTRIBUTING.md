# Contributing to Turing

Thank you for your interest in contributing to Turing.

You can contribute to Turing in several ways:
- Contribute to the Turing codebase
- Report bugs
- Create community extensions
- Create articles and documentation for users and contributors
- Help others answering questions about Turing

The following sections provide some suggestions and guidelines you should follow:
- [Code of Conduct](#coc)
- [Issues and Bugs](#issue)
- [Feature Requests](#feature)
- [Submission Guidelines](#submit)
- [Coding Rules](#rules)

## <a name="coc"></a> Code of Conduct

Help us keep Turing open and inclusive.
Please read and follow our [Code of Conduct](https://github.com/caraml-dev/turing/blob/main/CODE_OF_CONDUCT.md).

## <a name="issue"></a> Found a Bug?

If you find a bug in the source code, you can help us by [submitting an issue](#submit-issue) to our [GitHub Repository][github].
Even better, you can [submit a Pull Request](#submit-pr) with a fix.


## <a name="feature"></a> Missing a Feature?
You can *request* a new feature by [submitting an issue](#submit-issue) to our GitHub Repository.
If you would like to *implement* a new feature, please consider the size of the change in order to determine the right steps to proceed:

* For a **Major Feature**, first open an issue and outline your proposal so that it can be discussed.
  This process allows us to better coordinate our efforts, prevent duplication of work, and help you to craft the change so that it is successfully accepted into the project.

  **Note**: Adding a new topic to the documentation, or significantly re-writing a topic, counts as a major feature.

* **Small Features** can be crafted and directly [submitted as a Pull Request](#submit-pr).

## <a name="submit"></a> Submission Guidelines

### <a name="submit-issue"></a> Submitting an Issue

Before you submit an issue, please search the issue tracker. An issue for your problem might already exist and the discussion might inform you of workarounds readily available.

We want to fix all the issues as soon as possible, but before fixing a bug, we need to reproduce and confirm it.
In order to reproduce bugs, we require that you provide a minimal reproduction.
Having a minimal reproducible scenario gives us the necessary information without going back and forth to you with additional questions.

A minimal reproduction allows us to quickly confirm a bug (or point out a coding problem) as well as confirm that we are fixing the right problem.

We require a minimal reproduction to save maintainers' time and ultimately be able to fix more bugs.

Unfortunately, we are not able to investigate / fix bugs without a minimal reproduction, so if we don't hear back from you, we are going to close an issue that doesn't have enough information to be reproduced.

You can file new issues by selecting from our [new issue templates](https://github.com/caraml-dev/turing/issues/new/choose) and filling out the issue template.

### <a name="submit-pr"></a> Submitting a Pull Request (PR)

Before you submit your Pull Request (PR), consider the following guidelines:

1. Search [GitHub](https://github.com/caraml-dev/turing/pulls) for an open or closed PR that relates to your submission.
   You don't want to duplicate existing efforts.

2. Be sure that an issue describes the problem you're fixing, or documents the design for the feature you'd like to add.
   Discussing the design upfront helps to ensure that we're ready to accept your work.

3. Fork this Turing repository
4. Create a new branch in your forked repository
5. Commit the changes to the codebase
6. Follow our [Coding Rules](#rules).
7. Make sure all the tests pass
8. Submit a pull request to the main Turing codebase from your forked repository

## <a name="rules"></a> Coding Rules
To ensure consistency throughout the source code, keep these rules in mind as you are working:

* All features or bug fixes **must be tested** by one or more specs (unit-tests).

### Code Style & Linting

We are using [golangci-lint](https://github.com/golangci/golangci-lint), and we can run the following commands for formatting.

```sh
# Formatting for linting issues
make format

# Checking for linting issues
make lint
```

### Go tests

For **Unit** tests, we follow the convention of keeping it beside the main source file.

### Pre-commit Hooks

Setup [`pre-commit`](https://pre-commit.com/) to automatically lint and format the codebase on commit:

1. Ensure that you have Python (3.7 and above) with `pip`, installed.
2. Install `pre-commit` with `pip` &amp; install pre-push hooks

    ```sh
    # Clear existing hooks    
    git config --unset-all core.hooksPath
    rm -rf .git/hooks
    # Install hooks
    make setup
    ```

3. On push, the pre-commit hook will run. This runs `make format` and `make lint`.
