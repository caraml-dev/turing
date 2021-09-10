import React, { useCallback, useContext, useEffect } from "react";
import {
  EuiButton,
  EuiDragDropContext,
  euiDragDropReorder,
  EuiDraggable,
  EuiDroppable,
  EuiFlexGroup,
  EuiFlexItem,
  EuiSpacer,
} from "@elastic/eui";
import { FormLabelWithToolTip } from "../../../../../../../components/form/label_with_tooltip/FormLabelWithToolTip";
import { ExperimentCard } from "./experiment_card/ExperimentCard";
import ExperimentContext from "../../providers/context";
import { makeId } from "../../config";
import { Panel } from "../../../Panel";
import { useOnChangeHandler } from "../../../../../../../components/form/hooks/useOnChangeHandler";

export const ExperimentsConfigPanel = ({
  experiments,
  onChangeHandler,
  errors = [],
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const {
    experiments: allExperiments,
    isLoaded,
    setExperimentsValidated,
  } = useContext(ExperimentContext);

  const experimentsOptions = useCallback(
    (experiment) => {
      const expNames = experiments.map((e) => e.name);
      // Exclude already selected experiments
      return allExperiments.filter(
        ({ name }) => experiment.name === name || !expNames.includes(name)
      );
    },
    [experiments, allExperiments]
  );

  // Update selected experiments if new experiments list does not include some of them
  useEffect(() => {
    if (isLoaded("experiments")) {
      const allExperimentIds = allExperiments.map((e) => e.id);
      // Filter selected experiments against available (!e.id check is for the dummy experiment card)
      const filteredExps = experiments.filter(
        (e) => !e.id || allExperimentIds.includes(e.id)
      );
      if (filteredExps.length < experiments.length) {
        onChangeHandler(filteredExps);
      } else if (!experiments.length) {
        // Add a dummy experiment
        onChangeHandler([{ uuid: makeId() }]);
      }
      setExperimentsValidated(true);
    }
  }, [
    experiments,
    allExperiments,
    isLoaded,
    setExperimentsValidated,
    onChangeHandler,
  ]);

  // Define onchange handlers
  const onAddExperiment = useCallback(() => {
    onChangeHandler([...experiments, { uuid: makeId() }]);
  }, [onChangeHandler, experiments]);

  const onDeleteExperiment = (idx) => () => {
    experiments.splice(idx, 1);
    onChangeHandler([...experiments]);
  };

  const onDragEnd = ({ source, destination }) => {
    if (source && destination) {
      const items = euiDragDropReorder(
        experiments,
        source.index,
        destination.index
      );
      onChangeHandler([...items]);
    }
  };

  return (
    <Panel
      title={
        <FormLabelWithToolTip
          label="Experiments"
          size="m"
          content="Select one or more experiments. Drag and drop to re-order."
        />
      }>
      <EuiSpacer size="s" />
      <EuiDragDropContext onDragEnd={onDragEnd}>
        <EuiFlexGroup direction="column" gutterSize="s">
          <EuiDroppable droppableId="CUSTOM_HANDLE_DROPPABLE_AREA" spacing="m">
            {experiments.map((experiment, idx) => (
              <EuiDraggable
                key={`${experiment.uuid || experiment.id}`}
                index={idx}
                draggableId={`${experiment.uuid || experiment.id}`}
                customDragHandle={true}
                disableInteractiveElementBlocking>
                {(provided) => (
                  <EuiFlexItem>
                    <ExperimentCard
                      experiment={experiment}
                      experiments={experimentsOptions(experiment)}
                      onChangeHandler={onChange(`${idx}`)}
                      onDelete={
                        experiments.length > 1
                          ? onDeleteExperiment(idx)
                          : undefined
                      }
                      errors={errors[idx]}
                      dragHandleProps={provided.dragHandleProps}
                    />
                    <EuiSpacer size="s" />
                  </EuiFlexItem>
                )}
              </EuiDraggable>
            ))}
          </EuiDroppable>
          <EuiFlexItem>
            <EuiButton
              onClick={onAddExperiment}
              disabled={experiments.length >= allExperiments.length}>
              + Add Experiment
            </EuiButton>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiDragDropContext>
    </Panel>
  );
};
