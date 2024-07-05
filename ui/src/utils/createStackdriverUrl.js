// This file in almost an exact copy of this file from Merlin
// Ref: https://github.com/caraml-dev/merlin/blob/8edef22b29d0bfb2728d62b1f880f1f753f9509e/ui/src/utils/createStackdriverUrl.js
const stackdriverAPI = "https://console.cloud.google.com/logs/viewer";

const stackdriverFilter = query => {
  return `resource.type:"k8s_query" OR "k8s_container" OR "k8s_pod"
resource.labels.project_id:${query.gcp_project}
resource.labels.cluster_name:${query.cluster}
resource.labels.namespace_name:${query.namespace}
resource.labels.pod_name:${query.pod_name}
timestamp>"${query.start_time}"
`;
};

const stackdriverImageBuilderFilter =  (query, imagebuilder) => {
  return `resource.type:"k8s_container"
resource.labels.project_id:${imagebuilder.gcp_project}
resource.labels.cluster_name:${imagebuilder.cluster}
resource.labels.namespace_name:${imagebuilder.namespace}
labels.k8s-pod/job-name:${query.job_name}
timestamp>"${query.start_time}"`;
}

export const createStackdriverUrl = (query, component, imagebuilder) => {
  const advanceFilter = component === "ensembler_image_builder" ? stackdriverImageBuilderFilter(query, imagebuilder) : stackdriverFilter(query);

  const url = {
    project: query.gcp_project || imagebuilder.gcp_project,
    minLogLevel: 0,
    expandAll: false,
    advancedFilter: advanceFilter,
  };

  const stackdriverParams = new URLSearchParams(url).toString();
  return stackdriverAPI + "?" + stackdriverParams;
};

