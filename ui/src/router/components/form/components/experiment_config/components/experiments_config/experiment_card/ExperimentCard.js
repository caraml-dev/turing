import React, { useCallback, useContext, useMemo } from "react";
import { EuiFlexGroup, EuiFlexItem, EuiFormRow, EuiPanel } from "@elastic/eui";
import { EuiComboBoxSelect } from "../../../../../../../../components/form/combo_box/EuiComboBoxSelect";
import { ExperimentCardHeader } from "./card_header/ExperimentCardHeader";
import ExperimentContext from "../../../providers/context";
import { makeId } from "../../../config";
import sortBy from "lodash/sortBy";

export const ExperimentCard = ({
  experiment,
  experiments,
  onChangeHandler,
  onDelete,
  errors = {},
  ...props
}) => {
  const { experiments: allExperiments } = useContext(ExperimentContext);

  const experimentOptions = useMemo(() => {
    return sortBy(experiments, "name").map((exp) => ({
      icon: "beaker",
      label: exp.name,
    }));
  }, [experiments]);

  // Define onchange handler
  const onExperimentNameChange = useCallback(
    (expName) => {
      onChangeHandler({
        ...allExperiments.find((e) => e.name === expName),
        uuid: experiment.uuid || makeId(),
      });
    },
    [allExperiments, experiment.uuid, onChangeHandler]
  );

  return (
    <EuiPanel>
      <ExperimentCardHeader
        onDelete={onDelete}
        dragHandleProps={props.dragHandleProps}
      />
      <EuiFlexGroup direction="column">
        <EuiFlexItem>
          <EuiFormRow
            label="Experiment *"
            isInvalid={!!errors.name}
            error={errors.name}
            fullWidth>
            <ExperimentContext.Consumer>
              {({ isLoading }) => (
                <EuiComboBoxSelect
                  fullWidth
                  isLoading={isLoading("experiments")}
                  placeholder="Select Experiment..."
                  value={experiment.name || ""}
                  options={experimentOptions}
                  onChange={onExperimentNameChange}
                  isInvalid={!!errors.name}
                />
              )}
            </ExperimentContext.Consumer>
          </EuiFormRow>
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiPanel>
  );
};
