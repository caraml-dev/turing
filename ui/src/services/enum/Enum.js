export const EnumValue = (name, props) =>
  Object.freeze({
    ...props,
    toJSON: () => name,
    toString: () => name,
    valueOf: () => name,
  });

export const Enum = (values) => {
  const allValues = Object.values(values);

  return Object.freeze({
    ...values,
    values: allValues,
    fromValue: (name) => allValues.find((s) => name === s.toString()),
  });
};
