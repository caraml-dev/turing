export const autoscalingPolicyOptions = [
  {
    value: "concurrency",
    inputDisplay: "Concurrency",
    description: `The default metric, represents the maximum number of in-flight requests
                      to a single replica.`,
    defaultValue: "1",
  },
  {
    value: "rps",
    inputDisplay: "RPS",
    description: `The maximum number of requests per second, to a single replica.`,
    defaultValue: "100",
  },
  {
    value: "cpu",
    inputDisplay: "CPU",
    description: `Traditional CPU-based autoscaling.
                      When the specified target (% request) is reached, a new replica will be added.`,
    defaultValue: "80",
    unit: "%",
  },
  {
    value: "memory",
    inputDisplay: "Memory",
    description: `Traditional memory-based autoscaling.
                      When the specified target (% request) is reached, a new replica will be added.`,
    defaultValue: "80",
    unit: "%",
  },
];

export const autoscalingPolicyMetrics = autoscalingPolicyOptions.map(e => e.value);
