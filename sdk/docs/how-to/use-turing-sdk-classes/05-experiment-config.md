# ExperimentConfig

The `ExperimentConfig` class is a simple container to carry configuration related to an experiment to be used by a 
Turing Router. Note that as Turing does not create experiments automatically, you would need to create your 
experiments separately prior to specifying their configuration here.

Also, notice that `ExperimentConfig` does not contain any fixed schema as it simply carries configuration for 
generic experiment engines, which are used as plug-ins for Turing. When building an `ExperimentConfig` from scratch, 
you would need to consider the underlying schema for the `config` attribute as well as the appropriate `type` that 
corresponds to your selected experiment engine:

```python
@dataclass
class ExperimentConfig:
    type: str = "nop"
    config: Dict = None
```