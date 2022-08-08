import os
import pytest
import shutil
from pyspark import SparkConf, SparkContext
from pyspark.sql import SparkSession

shutil.rmtree("mlruns", ignore_errors=True)


@pytest.fixture(scope="session")
def spark(request):
    conf = SparkConf()
    conf.set(
        "spark.jars",
        "https://storage.googleapis.com/hadoop-lib/gcs/gcs-connector"
        "-hadoop2-2.0.1.jar",
    )
    conf.set(
        "spark.jars.packages",
        "com.google.cloud.spark:spark-bigquery-with-dependencies_2.12:0.19.1",
    )
    conf.set("spark.driver.host", "127.0.0.1")
    sc = SparkContext(master="local", conf=conf)

    sc._jsc.hadoopConfiguration().set(
        "fs.gs.impl", "com.google.cloud.hadoop.fs.gcs.GoogleHadoopFileSystem"
    )
    sc._jsc.hadoopConfiguration().set(
        "fs.AbstractFileSystem.gs.impl", "com.google.cloud.hadoop.fs.gcs.GoogleHadoopFS"
    )
    sc._jsc.hadoopConfiguration().set(
        "google.cloud.auth.service.account.enable", "true"
    )

    sa_path = os.environ.get("GOOGLE_APPLICATION_CREDENTIALS")
    if sa_path is not None:
        sc._jsc.hadoopConfiguration().set(
            "google.cloud.auth.service.account.json.keyfile", sa_path
        )

    spark = SparkSession.builder.config(conf=sc.getConf()).getOrCreate()

    request.addfinalizer(lambda: spark.stop())

    return spark
