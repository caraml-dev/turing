from typing import DefaultDict
import yaml
from pyspark import SparkConf, SparkContext
from pyspark.sql import SparkSession
from ensembler.job import BatchEnsemblingJob
from ensembler.ensembler import Ensembler
from ensembler.sink import Sink


def build_spark_session(
        app_name: str,
        spark_config: DefaultDict[str, str] = None,
        hadoop_config: DefaultDict[str, str] = None) -> SparkSession:
    if hadoop_config is None:
        hadoop_config = {}
    conf = SparkConf()
    if spark_config:
        conf.setAll(spark_config.items())

    sc = SparkContext(conf=conf)

    if hadoop_config:
        for k, v in hadoop_config.items():
            sc._jsc.hadoopConfiguration().set(k, v)

    return SparkSession.builder \
        .appName(app_name) \
        .config(conf=sc.getConf()) \
        .getOrCreate()


class SparkApplication(object):
    _ANNOTATION_GROUP_SPARK = 'spark'
    _ANNOTATION_GROUP_HADOOP = 'hadoopConfiguration'

    def __init__(self, args):
        self.job, raw_config = BatchEnsemblingJob.from_yaml(args.job_spec)
        print(
            'Job Configuration: \n'
            f'{yaml.dump(raw_config)}'
        )

    def run(self):
        annotations = {}
        for name, value in self.job.annotations().items():
            group, *keys = name.split('/', 1)
            annotations[group] = {
                **annotations.get(group, {}),
                "".join(keys): value
            }
        annotations.update()
        spark = build_spark_session(
            self.job.name(),
            annotations.get(self._ANNOTATION_GROUP_SPARK),
            annotations.get(self._ANNOTATION_GROUP_HADOOP)
        )
        self.job.run(spark)


__all__ = [
    'SparkApplication'
]
