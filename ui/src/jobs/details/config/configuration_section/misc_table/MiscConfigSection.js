import { EuiDescriptionList, EuiLink } from "@elastic/eui";
import React, { useContext } from "react";
import { getGCSDashboardUrl } from "../../../../../utils/gcp";
import { ProjectsContext } from "@gojek/mlp-ui";

export const MiscConfigSection = ({
  infra_config: { artifact_uri, service_account_name },
}) => {
  const { currentProject = {} } = useContext(ProjectsContext);

  const items = [
    {
      title: "Ensembler Artifact URI",
      description: (
        <EuiLink href={getGCSDashboardUrl(artifact_uri)} target="_blank">
          {artifact_uri}
        </EuiLink>
      ),
    },
    {
      title: "Service Account",
      description: (
        <EuiLink href={`/projects/${currentProject.id}/settings/secrets-management`}>
          {service_account_name}
        </EuiLink>
      ),
    },
  ];

  return (
    <EuiDescriptionList
      compressed
      textStyle="reverse"
      type="responsiveColumn"
      listItems={items}
      titleProps={{ style: { width: "30%" } }}
      descriptionProps={{ style: { width: "70%" } }}
    />
  );
};
