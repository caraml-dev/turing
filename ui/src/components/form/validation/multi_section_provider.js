import * as React from "react";
import { useCallback, useContext, useEffect, useState } from "react";
import { FormContext } from "../context";
import FormValidationContext from ".";
import { extractErrors } from "./errors";
import zip from "lodash/zip";

export const MultiSectionFormValidationContextProvider = ({
  schemas,
  contexts,
  onSubmit,
  children,
}) => {
  const { data: formData } = useContext(FormContext);

  // identifies if user tried to submit this form
  const [isTouched, setIsTouched] = useState(false);

  // identifies if the form was validated
  const [isValidated, setIsValidated] = useState(false);

  // identifies if the form is in submission state
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errors, setErrors] = useState([]);

  const isValid = (errors) =>
    errors.reduce(
      (isValid, errors) => isValid && !Object.keys(errors).length,
      true
    );

  const onStartSubmitting = (event) => {
    event && event.preventDefault();
    setIsTouched(true);
    setIsSubmitting(true);
  };

  const onFinishSubmitting = useCallback(() => {
    setIsTouched(false);
    setIsValidated(false);
    setIsSubmitting(false);
    onSubmit();
  }, [onSubmit]);

  useEffect(() => {
    if (isTouched) {
      if (schemas) {
        Promise.all(
          zip(schemas, contexts).map(([schema, ctx]) => {
            return !!schema
              ? new Promise((resolve, reject) => {
                  schema
                    .validate(formData, {
                      abortEarly: false,
                      context: ctx,
                    })
                    .then(
                      () => resolve({}),
                      (err) => resolve(extractErrors(err))
                    );
                })
              : Promise.resolve({});
          })
        )
          .then(setErrors)
          .then(() => setIsValidated(true));
      } else {
        setIsValidated(true);
      }
    }
  }, [isTouched, schemas, contexts, formData]);

  useEffect(() => {
    if (isSubmitting && isValidated) {
      isValid(errors) ? onFinishSubmitting() : setIsSubmitting(false);
    }
  }, [isSubmitting, isValidated, errors, onFinishSubmitting]);

  return (
    <FormValidationContext.Provider
      value={{ onSubmit: onStartSubmitting, isSubmitting, errors }}>
      {children}
    </FormValidationContext.Provider>
  );
};
