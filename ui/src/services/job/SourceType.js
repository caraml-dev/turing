import { Enum, EnumValue } from "../enum/Enum";

export const SourceType = Enum({
  BQ: EnumValue("BQ", {
    label: "Google BigQuery",
  }),
});
