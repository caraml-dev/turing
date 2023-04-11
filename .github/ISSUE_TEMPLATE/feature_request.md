name: Feature Request
description: Request a new feature for Turing components.
labels: ['type: feature request']
body:
  - type: textarea
    attributes:
      label: Is your feature request related to a problem? Please describe.
      description: A clear and concise description of what the problem is. Ex. I'm facing [...] when trying to [...]
    validations:
      required: true
  - type: textarea
    attributes:
      label: Describe the solution you would like
      description: A clear and concise description of what you want to happen.
    validations:
      required: true
  - type: textarea
    attributes:
      label: Describe alternatives you have considered
      description: A clear and concise description of any alternative solutions or features you have considered.
    validations:
      required: true
  - type: textarea
    attributes:
      label: Additional context
      description: Add any other context or screenshots about the feature request here.
    validations:
      required: false
