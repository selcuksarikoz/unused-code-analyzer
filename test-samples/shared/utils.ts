// Utility module for testing unused code detection
// Contains various exports for testing different import patterns

// ========== BASIC EXPORTS ==========

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

// ========== TYPE EXPORTS FOR TYPE-IMPORT TESTS ==========

export type { UsedType as UsedSimpleType };
export type { UsedType as UsedType1 };
export type { UsedType as UsedTypeX };
export type { UsedInterface as UsedDefaultType };

export interface UsedTypeNamespace {
  SomeType: string;
}

export type InlineUsedType = string;
export type InlineUnusedType = number;
export type MixedUsedType = string;
export type MixedUnusedType = number;

export const valueExport = "value";
export const usedValueExport = "used value";
export const mixedUsedValue = "mixed used";
export const something = "something";

// ========== EXPORTS FOR UNUSED-IMPORTS TEST ==========

export const unusedNamed1 = "unused1";
export const unusedNamed2 = "unused2";
export const unusedNamed3 = "unused3";
export const usedNamed = "used";
export const usedValue = "value used";
export const UnusedDefault = "default export";
export const UsedDefault = "used default";
export const UnusedNamespace = {};
export const UsedNamespace = {};
export type UnusedType1 = string;
export type UnusedType2 = number;
export type UsedType2 = boolean;

// ========== EXPORTS FOR UNUSED-VARIABLES TEST ==========

export const someValue = 123;

// ========== EXPORTS FOR UNUSED-PARAMETERS TEST ==========

export const someConstant = "constant";

// ========== EXPORTS FOR MIXED TEST ==========

export const usedUtil = "util";
export const unusedUtil1 = "unused1";
export const unusedUtil2 = "unused2";

// ========== EXPORTS FOR REACT TEST ==========

export const formatData = (data: string): string => data.toUpperCase();
export const unusedHelper = "not used";

export interface ComponentProps {
  title: string;
}

export interface UnusedProps {
  unused: boolean;
}

// ========== ALIASED EXPORTS ==========

export { usedFunction as anotherName };
export { usedConst as usedUtilAlias };

export default function defaultUsed(): void {
  console.log("default");
}

// ========== UNUSED EXPORTS (for testing detection) ==========

export const unusedExport = "unused";
export function anotherUnused(): void {}

// Re-exports
export * from "./reexports";
