// JavaScript test file for unused detection
// Testing plain JS without TypeScript types

// 1. Unused imports
import { unusedHelper } from "./utils";
import UsedHelper from "./utils";

// 2. Unused variables
const unusedConst = "never used";
let unusedLet = 42;
var unusedVar = "var style";

// 3. Unused function
function unusedFunction() {
  console.log("not called");
}

// 4. Unused class
class UnusedClass {
  constructor() {
    this.value = 0;
  }
}

// 5. Function with unused parameter
function greet(name, unusedGreeting) {
  return `Hello ${name}`;
}

// 6. Arrow function with unused
const multiply = (a, unusedB) => {
  return a * 2;
};

// USED - should not be flagged
const usedConst = "I am used";
console.log(usedConst);

function usedFunction() {
  return "used";
}
usedFunction();

function properFunction(usedParam) {
  return usedParam + 1;
}
properFunction(10);

console.log(UsedHelper);

export const exportedValue = "exported";
