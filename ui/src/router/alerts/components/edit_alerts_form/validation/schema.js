/* eslint-disable no-template-curly-in-string */
import * as yup from "yup";

const durationRegex = /^(?!0+[smh])[0-9]+[smh]$/;

const durationSchema = yup
  .string()
  .required("Duration must have a positive value.")
  .matches(durationRegex, "Duration must have a positive value.");

const individualThresholdSchema = yup
  .number()
  .min(0, "Threshold cannot be less than 0");

const individualPercentageThresholdSchema = yup
  .number()
  .min(0, "Percentage threshold cannot be less than 0.")
  .max(100, "Percentage threshold cannot be more than 100.");

const atLeastOneThreshold = ({ path, options, createError }) => {
  const alert = options.originalValue;

  return (
    !!alert.warning_threshold ||
    !!alert.critical_threshold ||
    createError({
      message: "Alert must have at least 1 threshold value.",
      path: `${path}.overall`
    })
  );
};

const metricSchema = yup
  .object()
  .shape({
    warning_threshold: individualThresholdSchema,
    critical_threshold: individualThresholdSchema,
    duration: durationSchema
  })
  .test({
    name: "atLeast1",
    test: function() {
      return atLeastOneThreshold(this);
    }
  });

const percentageMetricSchema = yup
  .object()
  .shape({
    warning_threshold: individualPercentageThresholdSchema,
    critical_threshold: individualPercentageThresholdSchema,
    duration: durationSchema
  })
  .test({
    name: "atLeast1",
    test: function() {
      return atLeastOneThreshold(this);
    }
  });

const schema = {
  team: yup.object().shape({
    team: yup.string().required("Team name is required.")
  }),
  throughput: yup.object().shape({
    alerts: yup.object().shape({
      throughput: metricSchema
    })
  }),
  latency95p: yup.object().shape({
    alerts: yup.object().shape({
      latency95p: metricSchema
    })
  }),
  error_rate: yup.object().shape({
    alerts: yup.object().shape({
      error_rate: percentageMetricSchema
    })
  }),
  cpu_util: yup.object().shape({
    alerts: yup.object().shape({
      cpu_util: percentageMetricSchema
    })
  }),
  memory_util: yup.object().shape({
    alerts: yup.object().shape({
      memory_util: percentageMetricSchema
    })
  })
};

export default schema;