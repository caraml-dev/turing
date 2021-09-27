export const supportedAlerts = [
  {
    title: "Throughput",
    iconType: "inputOutput",
    metric: "throughput",
    unit: "rps",
    comparator: "lower",
  },
  {
    title: "Latency",
    iconType: "clock",
    metric: "latency95p",
    unit: "ms",
    comparator: "higher",
  },
  {
    title: "Error Rate",
    iconType: "cross",
    metric: "error_rate",
    unit: "%",
    comparator: "higher",
  },
  {
    title: "CPU Utilisation",
    iconType: "indexMapping",
    metric: "cpu_util",
    unit: "%",
    comparator: "higher",
  },
  {
    title: "Memory Utilisation",
    iconType: "memory",
    metric: "memory_util",
    unit: "%",
    comparator: "higher",
  },
];

export const durationOptions = [
  {
    value: "s",
    inputDisplay: "s",
  },
  {
    value: "m",
    inputDisplay: "m",
  },
  {
    value: "h",
    inputDisplay: "h",
  },
];
