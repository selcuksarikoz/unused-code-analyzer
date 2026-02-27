import { usedFunction, unusedExport, anotherUnused } from "./utils";

export function main(): void {
  usedFunction();
}

export const mainConst = "main";

export class MainClass {
  name: string;
}

const localUnused = "unused";
let localLetUnused = 1;
var localVarUnused = 2;

function funcWithUnusedParam(a: number, b: string, unusedParam: boolean): void {
  console.log(a, b);
}

function funcWithUnusedReturn(): number {
  return 1;
}

function funcWithMultipleParams(used: string, unused1: number, unused2: boolean): void {
  console.log(used);
}
