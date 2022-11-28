export const routingStrategyOptions = [
  {
    inputDisplay: "Selective",
    value: "selective",
    description: `Only the required route will be activated, based on the result from the experiment engine. This option is generally more cost-efficient.`,
    flag: "true",
  },
  {
    inputDisplay: "Exhaustive",
    value: "exhaustive",
    description: `All the routes applicable to the current request will be invoked, along with the experiment engine, in parallel. This option is generally more performant.`,
    flag: "false",
  }
];
