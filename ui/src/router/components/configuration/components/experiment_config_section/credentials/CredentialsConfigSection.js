import React, { Fragment } from "react";
import {
  EuiDescriptionList,
  EuiDescriptionListDescription,
  EuiDescriptionListTitle,
  EuiLink,
  EuiTextColor,
  EuiTitle,
} from "@elastic/eui";
import "./CredentialsConfigSection.scss";

import { getFormattedHomepageUrl } from "../../config";

export const CredentialsConfigSection = ({
  projectId,
  client,
  engineProps,
}) => {
  const username = client.username || "N/A";
  const homepageUrl =
    engineProps?.standard_experiment_manager_config?.home_page_url;

  return (
    <Fragment>
      <EuiTitle size="xs">
        {!!homepageUrl ? (
          <span>
            <EuiTextColor color="success">
              <EuiLink
                href={getFormattedHomepageUrl(homepageUrl, projectId)}
                target="_blank"
                external>
                {engineProps.display_name || engineProps.name}
              </EuiLink>
            </EuiTextColor>
          </span>
        ) : (
          <EuiDescriptionListTitle>{engineProps.name}</EuiDescriptionListTitle>
        )}
      </EuiTitle>
      <EuiDescriptionList
        textStyle="reverse"
        type="responsiveColumn"
        columnWidths={["70px", "auto"]}
        compressed>
        <EuiDescriptionListTitle>Client</EuiDescriptionListTitle>
        <EuiDescriptionListDescription title={username}>
          {username}
        </EuiDescriptionListDescription>
      </EuiDescriptionList>
    </Fragment>
  );
};
