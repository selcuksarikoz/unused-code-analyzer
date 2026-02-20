import { usedFunction, usedConst, UsedClass, anotherName } from "./utils";

const result = usedFunction();
console.log(result, usedConst);

const obj = new UsedClass("test");
console.log(obj.name);

anotherName();

import * as vscode from "vscode";

vscode.window.showInformationMessage("test");
