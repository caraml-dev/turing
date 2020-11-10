const getExperimentUrl = (engineProps, experiment) => {
  return engineProps && engineProps.home_page_url
    ? `${engineProps.home_page_url}/experiments/${experiment.id}`
    : "";
};

export default getExperimentUrl;
