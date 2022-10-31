export const enricherTypeOptions = {
  HTTP_JSON: [
    {
      value: "nop",
      inputDisplay: "None",
      description:
        "There will be no enrichment being done. The original request will be sent to the configured routes.",
    },
    {
      value: "docker",
      inputDisplay: "Docker",
      description:
        "Turing will deploy the specified image as a pre-processor and send the original request to it for enrichment.",
    },
    {
      value: "external",
      inputDisplay: "External (Coming Soon)",
      description:
        "Turing will send the original request to the external URL for enrichment.",
      disabled: true,
    },
  ],
  UPI_V1: [
    {
      value: "nop",
      inputDisplay: "None",
      description: "Enricher is not yet supported for the UPI protocol.",
      disabled: true,
    },
  ],
};
