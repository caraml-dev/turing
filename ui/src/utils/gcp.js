export const getBigQueryConsoleUrl = (table) => {
  if (table) {
    const parts = table.split(".");
    if (parts.length === 3) {
      const [project, dataset, table] = parts;
      return `https://console.cloud.google.com/bigquery?project=${project}&p=${project}&d=${dataset}&t=${table}&page=table`;
    }
  }
  return undefined;
};

export const getGCSDashboardUrl = (gcsBucketUri) => {
  const uri = gcsBucketUri.replace("gs://", "");
  return `https://console.cloud.google.com/storage/browser/${uri}`;
};
