import React from "react";
import {
  EuiCard,
  EuiFormRow,
  EuiSpacer,
  EuiSuperSelect,
  EuiText,
} from "@elastic/eui";
import { InMemoryTableForm } from "../../../../../../components/form/in_memory_table_form/InMemoryTableForm";
import "./TreatmentMappingCard.scss";

export const TreatmentMappingCard = ({
  experimentName,
  items = [],
  routeOptions,
  onChange,
  getError,
}) => {
  const columns = [
    {
      name: "Treatment",
      field: "treatment",
      width: "40%",
      render: (treatment) => <EuiText size="s">{treatment}</EuiText>,
    },
    {
      name: "Route",
      field: "route",
      width: "60%",
      render: (routeId, item) => (
        <EuiFormRow
          fullWidth
          isInvalid={!!getError(item.experiment, item.treatment)}
          error={getError(item.experiment, item.treatment)}>
          <EuiSuperSelect
            fullWidth
            hasDividers
            compressed
            valueOfSelected={routeId || "nop"}
            onChange={(routeId) =>
              onChange(item.experiment, item.treatment, routeId)
            }
            options={routeOptions}
            isInvalid={!!getError(item.experiment, item.treatment)}
          />
        </EuiFormRow>
      ),
    },
  ];

  return (
    <EuiCard
      className="euiCard--treatmentMappingCard"
      title=""
      description=""
      image={
        <div>
          <EuiSpacer size="m" />
          <EuiText textAlign="center" size="s">
            <b>{experimentName}</b>
          </EuiText>
          <EuiSpacer size="m" />
        </div>
      }
      textAlign="left">
      <InMemoryTableForm
        columns={columns}
        items={items}
        loading={items.length === 0}
      />
    </EuiCard>
  );
};
