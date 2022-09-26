export const autoscalingPolicyOptions = [
  {
    value: "concurrency",
    inputDisplay: "Concurrency",
    description: `The default metric, represents the maximum number of in-flight requests
                      to a single replica.`,
  },
  {
    value: "rps",
    inputDisplay: "RPS",
    description: `The maximum number of requests per second, to a single replica.`,
  },
  {
    value: "cpu",
    inputDisplay: "CPU",
    description: `Traditional CPU-based autoscaling.
                      When the specified target (% request) is reached, a new replica will be added.`,
    unit: "%",
  },
  {
    value: "memory",
    inputDisplay: "Memory",
    description: `Traditional memory-based autoscaling.
                      When the specified target (% request) is reached, a new replica will be added.`,
    unit: "%",
  },
];

export const autoscalingPolicyMetrics = autoscalingPolicyOptions.map(e => e.value);
