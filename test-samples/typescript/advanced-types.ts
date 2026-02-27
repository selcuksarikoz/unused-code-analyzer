// TypeScript Advanced Types Test File
// Testing unused types, interfaces, and generics

// ========== TYPE IMPORTS ==========
import type { BaseEntity } from './shared';
import type { UnusedBaseType } from './shared';

// ========== INTERFACES ==========
// Used interface
export interface User {
  id: number;
  name: string;
  email: string;
}

// Unused interfaces
interface UnusedUser {
  id: number;
  name: string;
}

interface AnotherUnusedInterface {
  value: string;
}

// ========== TYPES ==========
// Used types
export type UserId = number;
export type UserRole = 'admin' | 'user' | 'guest';

// Unused types
type UnusedStatus = 'active' | 'inactive';
type UnusedConfig = { debug: boolean };

// ========== GENERICS ==========
// Used generic
export interface Repository<T> {
  findById(id: number): T | null;
  save(entity: T): void;
}

// Unused generic interface
interface UnusedGeneric<T, U> {
  first: T;
  second: U;
}

// Unused generic type
type UnusedPair<T, U> = [T, U];

// ========== TYPE ALIASES ==========
// Used
export type Nullable<T> = T | null;

// Unused
type UnusedPartial<T> = Partial<T>;

// ========== MAPPED TYPES ==========
// Unused mapped type
type UnusedReadonly<T> = {
  readonly [P in keyof T]: T[P];
};

// ========== CONDITIONAL TYPES ==========
// Unused conditional type
type UnusedIsString<T> = T extends string ? true : false;

// ========== UNION TYPES ==========
// Used
export type Result = Success | Failure;

// Unused
type UnusedVariant = OptionA | OptionB;

interface Success {
  type: 'success';
  data: unknown;
}

interface Failure {
  type: 'failure';
  error: Error;
}

interface OptionA {
  kind: 'a';
}

interface OptionB {
  kind: 'b';
}

// ========== ENUMS ==========
// Used enum
export enum Status {
  Active = 'active',
  Inactive = 'inactive',
}

// Unused enum
enum UnusedPriority {
  Low = 1,
  High = 2,
}

// ========== CONST ASSERTIONS ==========
// Unused const assertion
const unusedConfig = {
  api: 'test',
  timeout: 5000
} as const;

// ========== NAMESPACE ==========
// Unused namespace
namespace UnusedNamespace {
  export interface Config {
    key: string;
  }
}

// ========== DECORATORS (if enabled) ==========
// function unusedDecorator(target: any) {
//   console.log(target);
// }

// Use some types
const user: User = { id: 1, name: 'Test', email: 'test@test.com' };
console.log(user, Status.Active);
