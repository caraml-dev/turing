import React from "react";
import { EuiCheckbox, EuiFlexItem, EuiForm, EuiFormRow } from "@elastic/eui";
import { DescribedFormGroup } from "../../../../../../components/form/described_form_group/DescribedFormGroup";
import { Panel } from "../../Panel";

const routingOrderDescription = {
  true: `All the routes applicable to the current request will be invoked, along with the experiment engine. This option is generally more performant.`,
  false: `Only the required route will be activated, based on the result from the experiment engine. This option is generally more cost-efficient.`
};

export const StandardEnsemblerRoutingOrderPanel = ({
  isLazyRouting,
  onChange,
  errors,
}) => (
  <EuiFlexItem>
    <Panel title="Routing Order">
      <EuiForm>
        <EuiFormRow
          isInvalid={!!errors}
          error={errors}
          fullWidth
          display="row">
          <DescribedFormGroup description={routingOrderDescription[isLazyRouting]}>
            <EuiFormRow
              fullWidth
              isInvalid={!!errors}
              error={errors}
              display="row">
              <EuiCheckbox
                id={"routing-order-checkbox"}
                label={"Lazy Routing"}
                checked={isLazyRouting}
                onChange={e => onChange(e.target.checked)}
              />
            </EuiFormRow>
          </DescribedFormGroup>
        </EuiFormRow>
      </EuiForm>
    </Panel>
  </EuiFlexItem>
);