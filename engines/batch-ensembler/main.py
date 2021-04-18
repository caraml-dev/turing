import argparse
import ensembler


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Run a PySpark batch ensembling job')

    parser.add_argument('--job-spec', dest='job_spec', type=str, required=True,
                        help='Path to the ensembling job YAML file specification')

    args = parser.parse_args()
    print(f'Called with arguments: {args}')

    application = ensembler.SparkApplication(args)
    application.run()
