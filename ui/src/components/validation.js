import * as yup from "yup";

const httpHeaderSchema = yup
  .string()
  .required('Valid Request Header value is required, e.g. "x-session-id"');

const jsonPathSchema = yup
  .string()
  .required(
    'Valid Request Payload json path is required, e.g. "my_object.session_id"'
  );

export const fieldSourceSchema = yup
  .mixed()
  .required("Valid Field Source should be selected")
  .oneOf(["header", "payload"], "Valid Field Source should be selected");

export const fieldSchema = fieldSource =>
  yup
    .mixed()
    .when(fieldSource, {
      is: "header",
      then: httpHeaderSchema
    })
    .when(fieldSource, {
      is: "payload",
      then: jsonPathSchema
    });
