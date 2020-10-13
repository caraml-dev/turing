# Turing Experiments

At the heart of the capabilities offered by Turing is the Experiment Engine, which manages the experiment configurations and performs the necessary computations to select an applicable treatment, for every request made to the Turing system. Differtent organizations and use cases may require different behaviours from the Experiment Engine and may already have existing systems in place that carry out this function. Thus, Turing offers an extensible architecture where developers can create and plug in their own Experiment Engines. This sub-repo holds the interface definitions that are required to be implemented by an Experiment Engine that wishes to integrate itself into Turing.

## Get Started

* Begin by understanding the [Experiment Concepts](./docs/concepts.md) in Turing
* Learn how to set up your local environment and start creating Experiment Engine plugins from the [Developer Guide](./docs/developer_guide.md)
