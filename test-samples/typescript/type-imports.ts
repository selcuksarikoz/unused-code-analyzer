// Test file for type-only imports
// Testing various type import patterns

import { valueExport } from "./utils";

// 1. Simple type-only import - unused
import type { UnusedSimpleType } from "./utils";

// 2. Type-only import - used
import type { UsedSimpleType } from "./utils";

// 3. Multiple type imports - all unused
import type { UnusedTypeA, UnusedTypeB, UnusedTypeC } from "./utils";

// 4. Multiple type imports - partially used
import type { UsedTypeX, UnusedTypeY } from "./utils";

// 5. Type-only default import - unused
import type UnusedDefaultType from "./utils";

// 6. Type-only default import - used
import type UsedDefaultType from "./utils";

// 7. Type-only namespace import - unused
import type * as UnusedTypeNamespace from "./utils";

// 8. Type-only namespace import - used
import type * as UsedTypeNamespace from "./utils";

// 9. Inline type specifier (TypeScript 4.5+) - unused
import { type InlineUnusedType } from "./utils";

// 10. Inline type specifier - used
import { type InlineUsedType, usedValueExport } from "./utils";

// 11. Mixed value and type imports - partially used
import { type MixedUnusedType, mixedUsedValue, type MixedUsedType } from "./utils";

// Use some types and values
const a: UsedSimpleType = "test";
const b: UsedTypeX = 123;
const c: UsedDefaultType = {};
const d: UsedTypeNamespace.SomeType = {};
const e: InlineUsedType = "inline";
const f: MixedUsedType = "mixed";

console.log(valueExport, usedValueExport, mixedUsedValue, a, b, c, d, e, f);
