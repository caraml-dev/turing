import React from "react";
import { DockerRegistryPopover } from "./DockerRegistryPopover";
import "./SelectDockerImageComboBox.scss";
import { EuiComboBoxSelect } from "../combo_box/EuiComboBoxSelect";

const extractRegistry = (image, registries) => {
  if (image) {
    const registry = registries.find((o) => image.startsWith(o.value));
    if (registry && registry.value) {
      image = image.substr(registry.value.length);
      image && image.startsWith("/") && (image = image.substr(1));
      return [registry.value, image];
    }
  }

  return ["", image];
};

export const SelectDockerImageComboBox = ({
  value,
  registryOptions,
  imageOptions,
  onChange,
  ...props
}) => {
  const [registry, image] = extractRegistry(value, registryOptions);

  const onRegistryChange = (value) => {
    onChange(value ? `${value}/${image}` : image);
  };

  const onImageChange = (value) => {
    onChange(registry ? `${registry}/${value}` : value);
  };

  return (
    <EuiComboBoxSelect
      className="euiComboBox--dockerImage"
      fullWidth={props.fullWidth}
      compressed={props.compressed}
      placeholder={props.placeholder}
      isInvalid={props.isInvalid}
      options={imageOptions}
      value={image}
      onChange={(value) => onImageChange(value || "")}
      onCreateOption={onImageChange}
      prepend={
        <DockerRegistryPopover
          value={registry}
          registryOptions={registryOptions}
          onChange={onRegistryChange}
        />
      }
    />
  );
};
