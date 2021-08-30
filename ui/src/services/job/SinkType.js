import { Enum, EnumValue } from "../enum/Enum";

export const SinkType = Enum({
  BQ: EnumValue("BQ", {
    label: "Google BigQuery",
  }),
});
