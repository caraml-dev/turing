import React, { Fragment } from "react";
import {
  EuiDescriptionList,
  EuiDescriptionListDescription,
  EuiDescriptionListTitle,
  EuiLink,
  EuiTextColor,
  EuiTitle
} from "@elastic/eui";
import "./CredentialsConfigSection.scss";

export const CredentialsConfigSection = ({
  engineType,
  deployment,
  client,
  engineProps
}) => {
  const username = client.username || "N/A";

  return (
    <Fragment>
      <EuiTitle size="xs">
        {!!engineProps.home_page_url ? (
          <EuiTextColor color="secondary">
            <EuiLink href={engineProps.home_page_url} target="_blank" external>
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
        <EuiDescriptionListTitle>Timeout</EuiDescriptionListTitle>
        <EuiDescriptionListDescription title={deployment.timeout}>
          {deployment.timeout}
        </EuiDescriptionListDescription>
        <EuiDescriptionListTitle>Client</EuiDescriptionListTitle>
        <EuiDescriptionListDescription title={username}>
          {username}
        </EuiDescriptionListDescription>
      </EuiDescriptionList>
    </Fragment>
  );
};
