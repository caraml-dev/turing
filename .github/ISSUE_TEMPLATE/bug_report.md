name: Bug Report
description: Report a software bug on Turing components.
labels: ['type: bug', 'status: waiting triage']
body:
  - type: dropdown
    attributes:
      label: Turing Component
      description: What Turing component is this issue related to?
      multiple: true
      options:
        - Core
        - Experiment Engine Plugin
        - Pyfunc Ensembler
        - SDK
        - UI
        - Others
    validations:
      required: true
  - type: textarea
    attributes:
      label: Describe the Issue
      description: A clear and concise description of the issue you encountered.
    validations:
      required: true
  - type: textarea
    attributes:
      label: Steps to Reproduce
      description: Please provide the steps to reproduce the issue.
      value: |
        1. 
        2. 
        3. 
    validations:
      required: true
  - type: textarea
    attributes:
      label: Expected behavior
      description: A clear and concise description of what you expected to happen.
    validations:
      required: true
  - type: textarea
    attributes:
      label: System
      description: Add your system information which surfaced this bug.
      value: |
        - Device: [e.g. Macbook, Windows Laptop]
        - OS: [e.g. macOS Monterey, Windows 10, Ubuntu 20.04]
        - Browser: [e.g. Chrome, Safari]
        - Version: [e.g. 22]
    validations:
      required: true
  - type: textarea
    attributes:
      label: Additional context
      description: Add any other context about the problem here. Or, screenshots (if applicable) to help explain your problem. You can drag and drop `png`, `jpg`, `gif`, etc. in this box.
    validations:
      required: false
