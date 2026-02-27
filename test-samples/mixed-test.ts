// Comprehensive mixed test file
// Contains unused imports, variables, and parameters all together

// ========== IMPORTS ==========
// Used imports
import { usedUtil } from "./utils";

// Unused named imports
import { unusedUtil1, unusedUtil2 } from "./utils";

// Unused default import
import UnusedDefault from "./utils";

// Unused namespace import
import * as UnusedNamespace from "./utils";

// Unused type imports
import type { UnusedType1, UnusedType2 } from "./utils";

// Used type import
import type { UsedType } from "./utils";

// ========== VARIABLES ==========
// Unused const
const UNUSED_CONFIG = { key: "value" };

// Unused let
let unusedCounter = 0;

// Unused function
function calculateSomething(): number {
  return 42;
}

// Unused class
class DataProcessor {
  process(data: string): string {
    return data.toUpperCase();
  }
}

// Unused interface
interface ConfigOptions {
  debug: boolean;
  timeout: number;
}

// Unused type
type HandlerFunction = (data: string) => void;

// ========== FUNCTIONS WITH PARAMETERS ==========

// Function with unused parameters
function processData(unusedData: string, unusedOptions: object): void {
  console.log("processing...");
}

// Function with mixed parameters
function formatOutput(usedValue: string, unusedFormat: string): string {
  return usedValue.trim();
}

// Arrow function with unused
const transformData = (input: string, unusedModifier: number): string => {
  return input.toLowerCase();
};

// Method with unused parameter
class Formatter {
  format(unusedPattern: string, value: string): string {
    return value;
  }
}

// ========== USED THINGS (should NOT be flagged) ==========
const USED_CONSTANT = "important";
console.log(USED_CONSTANT);

function actuallyUsed(): void {
  console.log("I am used");
}
actuallyUsed();

function usedWithParams(param1: string, param2: number): string {
  return `${param1}-${param2}`;
}
usedWithParams("test", 123);

// Use some imports
const result: UsedType = "test";
console.log(usedUtil, result);

// Export something
export const publicAPI = "available";
