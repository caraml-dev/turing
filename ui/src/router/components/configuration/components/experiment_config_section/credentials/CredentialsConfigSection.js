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

import { getFormattedHomepageUrl } from "../../../../../../router/components/configuration/components/config";

export const CredentialsConfigSection = ({
  projectId,
  engineType,
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
          <EuiTextColor color="secondary">
            <EuiLink
              href={getFormattedHomepageUrl(homepageUrl, projectId)}
              target="_blank"
              external>
              {engineProps.name}
            </EuiLink>
          </EuiTextColor>
        ) : (
          <EuiDescriptionListTitle>{engineProps.name}</EuiDescriptionListTitle>
        )}
      </EuiTitle>
      <EuiDescriptionList
        textStyle="reverse"
        type="responsiveColumn"
        compressed>
        <EuiDescriptionListTitle>Client</EuiDescriptionListTitle>
        <EuiDescriptionListDescription title={username}>
          {username}
        </EuiDescriptionListDescription>
      </EuiDescriptionList>
    </Fragment>
  );
};
