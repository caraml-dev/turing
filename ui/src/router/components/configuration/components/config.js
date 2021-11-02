import template from "lodash/template";
import templateSettings from "lodash/templateSettings";

export const getHomePageUrl = (engineProps, projectId) => {
  // If {{projectId}} present in the URL, substitute with actual project id.
  templateSettings.interpolate = /{{([\s\S]+?)}}/g;
  return template(engineProps.home_page_url)({
    projectId: projectId,
  });
};

export const getExperimentUrl = (engineProps, experiment) => {
  return engineProps && engineProps.home_page_url
    ? `${getHomePageUrl(engineProps)}/experiments/${experiment.id}`
    : "";
};
