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

function unusedFunction(): void {
  console.log("unused");
}

const unusedConst = 123;

export { usedFunction as anotherName };
