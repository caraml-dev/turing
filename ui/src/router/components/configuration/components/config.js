import template from "lodash/template";
import templateSettings from "lodash/templateSettings";

export const getFormattedHomepageUrl = (homepageUrl, projectId) => {
  // If {{projectId}} present in the URL, substitute with actual project id.
  templateSettings.interpolate = /{{([\s\S]+?)}}/g;
  return template(homepageUrl)({
    projectId: projectId,
  });
};

export const getExperimentUrl = (homepageUrl, experiment) => {
  return !!homepageUrl
    ? `${getFormattedHomepageUrl(homepageUrl)}/experiments/${experiment.id}`
    : "";
};
