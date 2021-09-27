import React, { useEffect, useState } from "react";
import { EuiComboBox } from "@elastic/eui";
import {
  ComboboxSuggestItem,
  EuiComboboxSuggestItemRowHeight,
} from "./ComboboxSuggestItem";

export const EuiComboBoxSelect = ({ value, onChange, options, ...props }) => {
  const [selected, setSelected] = useState([]);

  useEffect(() => {
    let selected = options.filter((suggestion) => suggestion.label === value);

    if (value && selected.length === 0) {
      selected = [{ label: value }];
    }
    setSelected(selected);
  }, [value, options, setSelected]);

  return (
    <EuiComboBox
      className={props.className}
      fullWidth={props.fullWidth}
      compressed={props.compressed}
      placeholder={props.placeholder}
      isLoading={props.isLoading}
      isInvalid={props.isInvalid}
      singleSelection={{ asPlainText: true }}
      noSuggestions={props.noSuggestions || !options}
      options={options}
      onChange={(selected) => {
        onChange(selected.length ? selected[0].label : undefined);
      }}
      isClearable={props.isClearable || true}
      selectedOptions={selected}
      rowHeight={EuiComboboxSuggestItemRowHeight}
      onCreateOption={props.onCreateOption}
      renderOption={(option, searchValue, optionContentClassname) => (
        <ComboboxSuggestItem
          {...{ option, searchValue, optionContentClassname }}
        />
      )}
      prepend={props.prepend}
    />
  );
};
