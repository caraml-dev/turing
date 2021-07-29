export const EnumValue = (name, props) =>
  Object.freeze({
    ...props,
    toJSON: () => name,
    toString: () => name
  });
