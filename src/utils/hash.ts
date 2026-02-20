import * as crypto from "crypto";

export function computeHash(content: string): string {
  return crypto.createHash("md5").update(content).digest("hex");
}
