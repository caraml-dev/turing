import argparse
import logging
import ensembler


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Run a PySpark batch ensembling job")

    parser.add_argument(
        "--job-spec",
        dest="job_spec",
        type=str,
        required=True,
        help="Path to the ensembling job YAML file specification",
    )
    parser.add_argument(
        "-l",
        "--log-level",
        dest="log_level",
        choices=["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"],
        help="Set the logging level",
        default=logging.DEBUG,
    )

    args = parser.parse_args()

    logging.basicConfig(level=args.log_level)
    # Disable debug log from py4j, because it's not actionable
    logging.getLogger("py4j").setLevel(logging.INFO)
    logging.info(
        "Called with arguments:\n%s\n",
        "\n".join([f"{k}: {v}" for k, v in vars(args).items()]),
    )

    application = ensembler.SparkApplication(args)
    application.run()
