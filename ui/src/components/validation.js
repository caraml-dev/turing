import * as yup from "yup";

const httpHeaderSchema = yup
  .string()
  .required('Valid Request Header value is required, e.g. "x-session-id"');

const jsonPathSchema = yup
  .string()
  .required(
    'Valid Request Payload json path is required, e.g. "my_object.session_id"'
  );

const predictionContextSchema = yup
.string()
.required(
  'Valid Prediction Context value is required, e.g. "model-a" in  "models"'
);
  
export const fieldSourceSchema = yup
  .mixed()
  .required("Valid Field Source should be selected")
  .oneOf(["header", "payload", "prediction_context"], "Valid Field Source should be selected");

export const fieldSchema = (fieldSource) =>
  yup
    .mixed()
    .when(fieldSource, {
      is: "header",
      then: httpHeaderSchema,
    })
    .when(fieldSource, {
      is: "payload",
      then: jsonPathSchema,
    })
    .when(fieldSource, {
      is: "prediction_context",
      then: predictionContextSchema,
    });
