from typing import DefaultDict, Optional
import logging
from pyspark import SparkConf, SparkContext
from pyspark.sql import SparkSession
from ensembler.job import BatchEnsemblingJob
from ensembler.ensembler import Ensembler


def build_spark_session(
    app_name: str,
    spark_config: Optional[DefaultDict[str, str]] = None,
    hadoop_config: Optional[DefaultDict[str, str]] = None,
) -> SparkSession:
    conf = SparkConf()
    if spark_config:
        conf.setAll(spark_config.items())

    sc = SparkContext(conf=conf)

    if hadoop_config:
        for k, v in hadoop_config.items():
            sc._jsc.hadoopConfiguration().set(k, v)

    return (
        SparkSession.builder.appName(app_name).config(conf=sc.getConf()).getOrCreate()
    )


class SparkApplication:
    _ANNOTATION_GROUP_SPARK = "spark"
    _ANNOTATION_GROUP_HADOOP = "hadoopConfiguration"

    def __init__(self, args):
        self._job, raw_config = BatchEnsemblingJob.from_yaml(args.job_spec)
        self._logger = logging.getLogger("SparkApplication")
        self._logger.debug(
            "Job Specification:\n"
            "===============================================================\n"
            "%s\n"
            "===============================================================\n",
            raw_config,
        )

    def run(self):
        annotations = {}
        for name, value in self._job.annotations().items():
            group, *keys = name.split("/", 1)
            annotations[group] = {**annotations.get(group, {}), "".join(keys): value}
        annotations.update()
        spark = build_spark_session(
            self._job.name(),
            annotations.get(self._ANNOTATION_GROUP_SPARK),
            annotations.get(self._ANNOTATION_GROUP_HADOOP),
        )
        self._job.run(spark)


__all__ = ["SparkApplication"]
