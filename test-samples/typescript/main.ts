import { usedFunction, usedConst } from "./utils";
import type { UsedInterface, UsedType } from "./utils";
import * as UtilsNamespace from "./utils";

const result = usedFunction();
console.log(result, usedConst);

const iface: UsedInterface = { id: 1 };
console.log(iface);

UtilsNamespace.usedFunction();

export { usedFunction as reExported };
export type { UsedInterface as ReExportedInterface } from "./utils";

import { unusedImport } from "./utils";

console.log("local");
