from typing import Optional, Any
import pandas
import turing


def main(turing_api: str, project: str):
    # Initialize Turing client
    turing.set_url(turing_api)
    turing.set_project(project)

    # List projects
    projects = turing.Project.list()
    for p in projects:
        print(p)

    # Implement new pyfunc ensembler
    class MyEnsembler(turing.ensembler.PyFunc):

        def ensemble(
                self,
                features: pandas.Series,
                predictions: pandas.Series,
                treatment_config: Optional[dict]) -> Any:
            """
            Simple ensembler, that returns predictions from the `model_odd`
            if `customer_id` is odd, or predictions from `model_even` otherwise
            """

            customer_id = features["customer_id"]
            if (customer_id % 2) == 0:
                return predictions['model_even']
            else:
                return predictions['model_odd']

    # Save pyfunc ensembler in Turing's backend
    ensembler = turing.PyFuncEnsembler.create(
        name="my-ensembler",
        ensembler_instance=MyEnsembler(),
        conda_env={
            'dependencies': [
                'python>=3.8.0',
                'cloudpickle==0.5.8',
                # other dependencies, if required
            ]
        }
    )
    print("Ensembler created:\n", ensembler)

    # Update Ensembler's name
    ensembler.update(name="my-ensembler-updated")
    print("Updated:\n", ensembler)

    # Update Ensembler's implementation
    ensembler.update(
        ensembler_instance=MyEnsembler(),
        conda_env={
            'channels': ['defaults'],
            'dependencies': [
                'python=3.7.0',
                "cookiecutter>=1.7.2",
                "numpy"
            ]
        },
        code_dir=["../samples"],
    )
    print("Updated:\n", ensembler)

    # List pyfunc ensemblers
    ensemblers = turing.PyFuncEnsembler.list()
    for e in ensemblers:
        print(e)


if __name__ == '__main__':
    import fire
    fire.Fire(main)
