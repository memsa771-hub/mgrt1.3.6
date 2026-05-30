import Database from "bun:sqlite";
import type { Statement } from "bun:sqlite";
import { resolve } from "path";
import { ensureDataDir } from "../paths";

export interface TypedDatabase extends Omit<Database, "run" | "query" | "prepare"> {
  run(sql: string, ...params: any[]): import("bun:sqlite").Changes;
  query<T = any>(sql: string): Statement<T>;
  prepare<T = any>(sql: string): Statement<T>;
}

const dataDir = ensureDataDir();
export const dbPath = resolve(dataDir, "overlord.db");
export const db: TypedDatabase = new Database(dbPath) as any;

console.log(`[db] Using database at: ${dbPath}`);
