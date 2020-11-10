export const enricherTypeOptions = [
  {
    value: "nop",
    inputDisplay: "None",
    description:
      "No enrichment gonna happen. Original request will be sent to configured routes"
  },
  {
    value: "docker",
    inputDisplay: "Docker",
    description:
      "Turing will deploy specified image as a pre-processor and will send original request to it for enrichment"
  },
  {
    value: "external",
    inputDisplay: "External (Coming Soon)",
    description:
      "Turing will send original request to the external URL for enrichment",
    disabled: true
  }
];
