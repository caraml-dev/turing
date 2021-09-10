import { useContext, useEffect, useState } from "react";
import FormValidationContext from "../form/validation";
import { slugify } from "@gojek/mlp-ui";
import { scroller } from "react-scroll";
import { animatedScrollConfig } from "./scroll";
import { isSectionInvalid } from "./AccordionForm";

export const AccordionFormScrollController = ({ sections }) => {
  const [isFormSubmissionInProgress, setFormSubmissionInProgress] =
    useState(false);
  const { isSubmitting, errors } = useContext(FormValidationContext);

  useEffect(() => {
    !!isSubmitting && setFormSubmissionInProgress(true);
  }, [isSubmitting]);

  useEffect(() => {
    // after submission is completed, scroll to the first section that has errors
    if (isFormSubmissionInProgress && !isSubmitting) {
      const errorSectionIdx = errors.findIndex(isSectionInvalid);
      if (errorSectionIdx !== -1)
        scroller.scrollTo(
          `${slugify(sections[errorSectionIdx].title)}`,
          animatedScrollConfig
        );
      setFormSubmissionInProgress(false);
    }
  }, [errors, sections, isFormSubmissionInProgress, isSubmitting]);

  return null;
};
