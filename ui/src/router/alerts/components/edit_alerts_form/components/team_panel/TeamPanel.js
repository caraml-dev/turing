import React, { Fragment, useCallback, useContext, useMemo } from "react";
import { Panel } from "../../../../../components/form/components/Panel";
import { EuiDescribedFormGroup, EuiFormRow, EuiSpacer } from "@elastic/eui";
import { EuiComboBoxSelect } from "../../../../../../components/form/combo_box/EuiComboBoxSelect";
import TeamsContext from "../../../../../../providers/teams/context";

export const TeamPanel = ({ team, onChange, errors }) => {
  const teams = useContext(TeamsContext);

  const teamOptions = useMemo(
    () =>
      teams.sort((a, b) => (a > b ? 1 : -1)).map((team) => ({ label: team })),
    [teams]
  );

  const onCreateOption = useCallback(
    (value) => {
      onChange(!!value ? value.trim().toLowerCase() : undefined);
    },
    [onChange]
  );

  return (
    <Panel contentWidth="65%">
      <EuiSpacer size="m" />
      <div>
        <EuiDescribedFormGroup
          title={<h3>Team Name</h3>}
          description={
            <Fragment>
              The notification will be sent to the specified team's Slack
              alerting channel.
              <EuiSpacer size="m" />
              Not sure about your team name? Please check{" "}
              <a
                href="https://go-jek.atlassian.net/wiki/spaces/DSP/pages/1731037258/Model+Endpoint+Alert#What-is-my-Team-Name?-How-to-get-one?"
                target="_blank"
                rel="noopener noreferrer">
                this guide
              </a>
              .
            </Fragment>
          }>
          <EuiFormRow hasEmptyLabelSpace isInvalid={!!errors} error={errors}>
            <EuiComboBoxSelect
              singleSelection={{ asPlainText: true }}
              placeholder="e.g. fraud, gofood, gopay"
              value={team}
              options={teamOptions}
              onChange={onChange}
              onCreateOption={onCreateOption}
              isInvalid={!!errors}
            />
          </EuiFormRow>
        </EuiDescribedFormGroup>
      </div>
      <EuiSpacer size="s" />
    </Panel>
  );
};
