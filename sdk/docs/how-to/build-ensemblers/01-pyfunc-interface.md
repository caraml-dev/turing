# Pyfunc Interface

In order to build a pyfunc ensembler and register it with Turing API, you would need to implement the 
`turing.ensembler.PyFunc` interface. This requires implementation 2 methods, `initialize` and `ensemble`.

```python
class MyEnsembler(turing.ensembler.PyFunc):
    """
    Simple ensembler, that returns predictions from the `model_odd`
    if `customer_id` is odd, or predictions from `model_even` otherwise
    """

    def initialize(self, artifacts: dict):
        pass

    def ensemble(
            self,
            features: pandas.Series,
            predictions: pandas.Series,
            treatment_config: Optional[dict]) -> Any:
        customer_id = features["customer_id"]
        if (customer_id % 2) == 0:
            return predictions['model_even']
        else:
            return predictions['model_odd']
```

`initialize` is essentially a method that gets called when an object of your implemented class gets instantiated. 

`ensemble` on the other hand, works a little differently depending on whether you plan to implement the 
`turing.ensembler.PyFunc` interface as a [batch ensembler](#batch-ensembler) or a 
[real-time ensembler service](#real-time-ensembler-service).

## Batch Ensembler
As a batch ensembler, `ensemble` will be called when performing ensembling jobs.

The source and predictions columns specified in your ensembling job will be passed **row-by-row** to `ensemble` as the 
arguments `features` and `predictions` respectively. These two arguments are `pandas.Series` objects which you can 
manipulate freely with in your implementation of the `ensemble` method. 

Note that for the moment, the argument `treatment_config` is not supported in batch ensembling; i.e. it will be 
passed as `None` in each call of `ensemble`.

The results of `ensemble` will then be stored in the sink that you have specified as part of the ensembling job.

The following example shows how the `ensemble` method may look like when being implemented for batch ensembling: 
```python
def ensemble(
        self,
        features: pandas.Series,
        predictions: pandas.Series,
        treatment_config: Optional[dict]) -> Any:
    customer_id = features["customer_id"]
    if (customer_id % 2) == 0:
        return predictions['model_even']
    else:
        return predictions['model_odd']
```

## Real-Time Ensembler Service
As a real-time ensembler service, `ensemble` will be called each time an individual request gets sent to the 
ensembler service (which gets automatically built and run as part of a Turing router deployment).

Each time a Turing Router sends a request to a 
[pyfunc ensembler service](../../../../docs/how-to/create-a-router/configure-ensembler.md#pyfunc-ensembler), 
`ensemble` will be called, with the request payload being passed as a `dict` object for the `features` argument, and 
the route responses as a `list` of `dict` for the `predictions` argument. 

If an experiment has been set up, the experiment returned would also be passed as a `dict` object for the 
`treatment_config` argument.

The results of `ensemble` will then be returned as a `json` payload to the Turing Router.

The following example shows how the `ensemble` method may look like when being implemented for real-time ensembling:

```python
def ensemble(
        self,
        features: dict,
        predictions: List[dict],
        treatment_config: dict) -> Any:
    if features['control_first'] is True:
        result = ['control', 'treatment-a']
    else:
        result = ['treatment-a', 'control']
    return "-ensembled-with-".join(result)
```