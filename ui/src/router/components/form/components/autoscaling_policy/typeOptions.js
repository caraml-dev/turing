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
                      When the specified target (%) is reached, a new replica will be added.`,
  },
  {
    value: "memory",
    inputDisplay: "Memory",
    description: `Traditional memory-based autoscaling.
                      When the specified target (in MiB) is reached, a new replica will be added.`,
  },
];

export const autoscalingPolicyMetrics = autoscalingPolicyOptions.map(e => e.value);
