# fluentd-bigquery

Docker image for fluentd with BigQuery output plugin. This image currently accepts JSON / MessagePack data as http messages using `in_http` plugin and events on a `tcp` socket using `in_forward` plugin.

## Running the Image
A number of ENV vars are required to be set in the container, for `fluentd.conf`. They are documented in `.env.template`. A container can be created from the image and run as follows.
1. Build the image locally (if not using a published image):

```sh
make build-image
```
2. Run, supplying the relevant env vars, volumes and ports. As an example, assuming `FLUENTD_GCP_JSON_KEY_PATH` is set to `/key.json`, the local key file is `gcp_key.json` and the `FLUENTD_LOG_PATH` to set to `/fluentd/log/bq_load_logs.*.buffer`:

```sh
docker run -it --rm --name docker-fluent-bq \
    --env-file .env \
    -v $(pwd)/log:/fluentd/log \
    -v $(pwd)/gcp_key.json:/key.json \
    -p 24224:24224 \
    -p 9880:9880 \
    fluentd-bigquery:latest
```

## Reference
* [Fluentd Docker image](https://hub.docker.com/r/fluent/fluentd/)
* [Fluentd BigQuery plugin](https://github.com/fluent-plugins-nursery/fluent-plugin-bigquery)