export const isEmpty = (obj) => !obj || !Object.keys(obj).length;

// Creates a copy of original object and removes specific keys from it
export const stripKeys = (obj, keys, deep = true) =>
  obj !== Object(obj)
    ? obj
    : Array.isArray(obj)
    ? deep
      ? obj.map((item) => stripKeys(item, keys, deep))
      : obj
    : Object.keys(obj)
        .filter((k) => !keys.includes(k))
        .reduce(
          (acc, x) =>
            Object.assign(acc, {
              [x]: deep ? stripKeys(obj[x], keys, deep) : obj[x],
            }),
          {}
        );
