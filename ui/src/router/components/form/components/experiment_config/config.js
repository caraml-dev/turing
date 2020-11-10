import { htmlIdGenerator } from "@elastic/eui/lib/services";

export const makeId = htmlIdGenerator();

export const initConfig = () => ({
  engine: {}, // Will be used for validation
  client: {},
  experiments: [],
  variables: { config: [] }
});

export const newVariableConfig = () => ({
  field_source: "header",
  required: false,
  field: "",
  name: ""
});

export const resetVariableSelection = v => {
  v.field = "";
  delete v.selected;
};
