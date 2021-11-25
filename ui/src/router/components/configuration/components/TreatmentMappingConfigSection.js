import React, { useContext } from "react";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiInMemoryTable,
  EuiLink,
  EuiTextColor,
  EuiTitle,
} from "@elastic/eui";
import { ConfigSectionPanel } from "../../../../components/config_section";
import { getExperimentUrl } from "./config";
import ExperimentEngineContext from "../../../../providers/experiments/context";

const TreatmentMappingConfigTable = ({ items }) => {
  const columns = [
    {
      field: "treatment",
      width: "50%",
      name: "Variant",
    },
    {
      field: "route",
      width: "50%",
      name: "Route",
    },
  ];

  return (
    <EuiInMemoryTable items={items} columns={columns} isSelectable={false} />
  );
};

export const TreatmentMappingConfigSection = ({
  engine,
  experiments = [],
  mappings = [],
}) => {
  const { getEngineProperties } = useContext(ExperimentEngineContext);
  const engineProps = getEngineProperties(engine);

  const experimentNames = [
    ...new Set(mappings.map((mapObj) => mapObj.experiment)),
  ];
  const experimentInfo = experiments.reduce((acc, exp) => {
    acc[exp.name] = exp;
    return acc;
  }, {});

  return (
    <ConfigSectionPanel title="Treatment Mapping" className="treatmentMapping">
      <EuiFlexGroup gutterSize="xl">
        {experimentNames.map((exp) => (
          <EuiFlexItem key={`${exp}--mappingConfig`}>
            <EuiTitle size="xxs">
              <EuiTextColor color="secondary">
                <EuiLink
                  href={getExperimentUrl(
                    engineProps?.standard_experiment_manager_config
                      ?.home_page_url,
                    experimentInfo[exp] || {}
                  )}
                  target="_blank"
                  external>
                  <code>{exp}</code>
                </EuiLink>
              </EuiTextColor>
            </EuiTitle>
            <TreatmentMappingConfigTable
              items={mappings.filter((mapObj) => mapObj.experiment === exp)}
            />
          </EuiFlexItem>
        ))}
      </EuiFlexGroup>
    </ConfigSectionPanel>
  );
};
