export function usedFunction(): string {
  return "hello";
}

export const usedConst = "world";

export class UsedClass {
  name: string;
  constructor(name: string) {
    this.name = name;
  }
}

export interface UsedInterface {
  id: number;
}

export type UsedType = string | number;

export enum UsedEnum {
  A = "a",
  B = "b",
}

function unusedFunction(): void {
  console.log("unused");
}

const unusedConst = 123;

let unusedLet = "unused";

var unusedVar = "unused";

class UnusedClass {
  value: string;
}

interface UnusedInterface {
  id: number;
}

type UnusedType = string;

enum UnusedEnum {
  A = "a",
}

export { usedFunction as anotherName };

export default function defaultUsed(): void {}

export const unusedExport = "unused";
export function anotherUnused(): void {}

export * from "./reexports";
