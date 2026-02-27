// Test file for unused imports detection
// This file tests various import patterns

// 1. Named imports - unused
import { unusedNamed1, unusedNamed2 } from "./utils";

// 2. Named imports - partially used
import { usedNamed, unusedNamed3 } from "./utils";

// 3. Default import - unused
import UnusedDefault from "./utils";

// 4. Default import - used
import UsedDefault from "./utils";

// 5. Namespace import - unused
import * as UnusedNamespace from "./utils";

// 6. Namespace import - used
import * as UsedNamespace from "./utils";

// 7. Type-only import - unused
import type { UnusedType1 } from "./utils";

// 8. Type-only import - used
import type { UsedType1 } from "./utils";

// 9. Mixed type and value imports - partially used
import { type UnusedType2, unusedValue, type UsedType2, usedValue } from "./utils";

// 10. Side-effect import (should not be flagged as unused)
import "./utils";

// Use some imports
console.log(usedNamed);
console.log(UsedDefault);
console.log(UsedNamespace);

const x: UsedType1 = "test";
const y: UsedType2 = 123;
console.log(usedValue, x, y);

export { usedNamed };
