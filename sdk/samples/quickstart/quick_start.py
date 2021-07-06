import os
import turing
import turing.batch
import turing.batch.config
from samples.common import MyEnsembler


def main(turing_api: str, project: str):
    # Initialize Turing client
    turing.set_url(turing_api)
    turing.set_project(project)

    # List projects
    projects = turing.Project.list()
    for p in projects:
        print(p)

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
        code_dir=[os.path.join(os.path.dirname(__file__), "../../samples")],
    )
    print("Updated:\n", ensembler)

    # List pyfunc ensemblers
    ensemblers = turing.PyFuncEnsembler.list()
    for e in ensemblers:
        print(e)


if __name__ == '__main__':
    import fire
    fire.Fire(main)
