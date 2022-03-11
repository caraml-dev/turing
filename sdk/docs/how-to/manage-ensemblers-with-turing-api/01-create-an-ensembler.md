# Create an Ensembler

An ensembler can be registered with Turing API to be used for batch ensembling or [real-time ensembling as a 
service](../../../../docs/how-to/create-a-router/configure-ensembler.md#pyfunc-ensembler). At the moment, only 
ensemblers written in `mlflow`'s Pyfunc flavour can be registered on Turing API. 

To create an (Pyfunc) ensembler using Turing SDK, you will need to first implement the `turing.ensembler.PyFunc` interface:

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

Then, you will simply need to run, specifying the name of your ensembler, your implementation of the 
`turing.ensembler.PyFunc` class as well as other `python` dependencies need to run it:

```python
# Save pyfunc ensembler in Turing's backend
ensembler = turing.PyFuncEnsembler.create(
    name="my-ensembler",
    ensembler_instance=MyEnsembler(),
    conda_env={
        'dependencies': [
            'python>=3.8.0',
            # other dependencies, if required
        ]
    }
)
```

The return value would be a `PyFuncEnsembler` object representing the ensembler that has been created if Turing API has 
created it successfully. 

