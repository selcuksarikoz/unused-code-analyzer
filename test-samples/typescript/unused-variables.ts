// Test file for unused variables detection
// This file tests various variable declaration patterns

import { something } from "./utils";

// 1. Unused const
const UNUSED_CONST = "never used";

// 2. Unused let
let unusedLet = 42;

// 3. Unused var
var unusedVar = "old style";

// 4. Unused function declaration
function unusedFunction(): void {
  console.log("never called");
}

// 5. Unused class
class UnusedClass {
  private value: number;
  constructor() {
    this.value = 0;
  }
}

// 6. Unused interface
interface UnusedInterface {
  id: number;
  name: string;
}

// 7. Unused type alias
type UnusedType = string | number | boolean;

// 8. Unused enum
enum UnusedEnum {
  OptionA = "a",
  OptionB = "b",
}

// 9. Multiple unused in same declaration
const unusedA = 1, unusedB = 2, unusedC = 3;

// 10. Destructured but unused
const { unusedProp1, unusedProp2 } = { unusedProp1: 1, unusedProp2: 2 };

// 11. Array destructured but unused
const [unusedFirst, unusedSecond] = [1, 2];

// 12. Used const (should NOT be flagged)
const USED_CONST = "i am used";
console.log(USED_CONST);

// 13. Used function (should NOT be flagged)
function usedFunction(): string {
  return "used";
}
usedFunction();

// 14. Exported but unused in this file (should NOT be flagged - might be used elsewhere)
export const exportedUnused = "exported";

// 15. Variable used only in type position (edge case)
const typeValue = { id: 1 };
type TypeFromValue = typeof typeValue;

console.log(something);
