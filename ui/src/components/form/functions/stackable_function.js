import { ExtensibleFunction } from "./extensible_function";

export class StackableFunction extends ExtensibleFunction {
  constructor(args, apply) {
    super((value) => {
      apply(args, value);
    });
    this.apply = apply;
    this.args = args;
  }

  withArg(arg) {
    return new StackableFunction([...this.args, arg], this.apply);
  }
}
