import * as yup from "yup";

const httpHeaderSchema = (_) => yup
  .string()
  .required('Valid Request Header value is required, e.g. "x-session-id"');

const jsonPathSchema = (_) => yup
  .string()
  .required('Valid Request Payload json path is required, e.g. "my_object.session_id"');

const predictionContextSchema = (_) => yup
  .string()
  .required('Valid Prediction Context value is required, e.g. "model_name"');
  
export const fieldSourceSchema = (_) => yup
  .string()
  .required("Valid Field Source should be selected")
  .oneOf(["header", "payload", "prediction_context"], "Valid Field Source should be selected");

export const fieldSchema = (fieldSource) => (_) =>
  yup
    .string()
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
