import React, { useEffect } from "react";
import { EuiFlexGroup, EuiFlexItem, EuiSpacer } from "@elastic/eui";
import { get } from "../../../../../../components/form/utils";
import { isEmpty } from "../../../../../../utils/object";
import { Panel } from "../../Panel";
import { TreatmentMappingCard } from "./TreatmentMappingCard";
import { useOnChangeHandler } from "../../../../../../components/form/hooks/useOnChangeHandler";
import { newMapping } from "../../../../../../services/ensembler";

export const StandardEnsemblerPanel = ({
  experiments = [],
  mappings = [],
  routeOptions,
  onChangeHandler,
  errors = {},
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const experimentTreatments = experiments.reduce((acc, exp) => {
    if (!!exp.name) {
      acc[exp.name] = exp.variants.map((v) => v.name);
    }
    return acc;
  }, {});
  const experimentNames = experiments
    .filter((e) => !!e.name)
    .map((e) => e.name);

  useEffect(() => {
    const mappingExperiments = [...new Set(mappings.map((m) => m.experiment))];
    const isSame =
      experimentNames.length === mappingExperiments.length &&
      experimentNames.every((name) => mappingExperiments.includes(name));
    if (!isSame && !isEmpty(experimentTreatments)) {
      const editedMappings = mappings.filter((m) =>
        experimentNames.includes(m.experiment)
      );
      experimentNames.forEach((expName) => {
        if (!mappingExperiments.includes(expName)) {
          experimentTreatments[expName].forEach((t) =>
            editedMappings.push(newMapping(expName, t))
          );
        }
      });
      onChangeHandler(editedMappings);
    }
  }, [experimentNames, experimentTreatments, mappings, onChangeHandler]);

  const onChangeMapping = (experiment, treatment, route) => {
    const idx = mappings.findIndex(
      (mapObj) =>
        mapObj.experiment === experiment && mapObj.treatment === treatment
    );
    onChange(`${idx}.route`)(route);
  };

  const getError = (experiment, treatment) => {
    const idx = mappings.findIndex(
      (mapObj) =>
        mapObj.experiment === experiment && mapObj.treatment === treatment
    );
    return get(errors, `${idx}.route`);
  };

  return (
    <Panel title="Treatment Mapping">
      <EuiSpacer size="l" />
      <EuiFlexGroup direction="column" gutterSize="s">
        {experimentNames.length > 0 &&
          experimentNames.map((experiment, idx) => (
            <EuiFlexItem key={`variantMapping-${idx}`}>
              <TreatmentMappingCard
                experimentName={experiment}
                items={mappings.filter(
                  (mapObj) => mapObj.experiment === experiment
                )}
                routeOptions={routeOptions}
                onChange={onChangeMapping}
                getError={getError}
              />
              <EuiSpacer size="m" />
            </EuiFlexItem>
          ))}
      </EuiFlexGroup>
    </Panel>
  );
};
