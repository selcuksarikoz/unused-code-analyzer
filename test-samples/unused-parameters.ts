// Test file for unused parameters detection
// This file tests various function parameter patterns

import { something } from "./utils";

// 1. Single unused parameter
function singleUnused(unusedParam: string): void {
  console.log("no param used");
}

// 2. Multiple unused parameters
function multipleUnused(unused1: number, unused2: string, unused3: boolean): void {
  console.log("none used");
}

// 3. Partially unused parameters
function partiallyUnused(used: string, unused: number): void {
  console.log(used);
}

// 4. Unused rest parameters
function unusedRest(...unusedArgs: any[]): void {
  console.log("rest not used");
}

// 5. Unused destructured parameter
function unusedDestructured({ unusedName, unusedAge }: { unusedName: string; unusedAge: number }): void {
  console.log("destructured not used");
}

// 6. Arrow function with unused parameter
const arrowUnused = (unusedX: number): number => {
  return 42;
};

// 7. Arrow function with multiple unused
const arrowMultipleUnused = (unusedA: string, unusedB: number): void => {
  console.log("arrow");
};

// 8. Method with unused parameter
class MyClass {
  methodWithUnused(unusedArg: string): void {
    console.log("method");
  }
}

// 9. Callback with unused (common pattern - might be intentionally unused)
const callbackExample = [1, 2, 3].map((unusedItem, usedIndex) => {
  return usedIndex;
});

// 10. Underscore prefix (convention for intentionally unused)
function underscorePrefix(_unused: string): void {
  console.log("underscore prefix");
}

// 11. All parameters used (should NOT be flagged)
function allUsed(param1: string, param2: number): string {
  return `${param1}${param2}`;
}

// 12. Used in function body (should NOT be flagged)
function paramUsed(param: string): void {
  console.log(param);
}

console.log(something, callbackExample, MyClass, allUsed, paramUsed);
