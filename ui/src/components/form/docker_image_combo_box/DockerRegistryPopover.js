import {  EuiSuperSelect } from "@elastic/eui";

export const DockerRegistryPopover = ({ value, registryOptions, onChange }) => {
  return (
    <EuiSuperSelect
          fullWidth
          options={registryOptions}
          valueOfSelected={value}
          itemLayoutAlign="top"
          onChange={onChange}
          isInvalid={false}
        />
  );
};