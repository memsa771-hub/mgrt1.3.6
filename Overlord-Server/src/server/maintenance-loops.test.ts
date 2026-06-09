import { describe, expect, test } from "bun:test";
import { getHeartbeatBatchSize } from "./maintenance-loops";

describe("maintenance heartbeat batching", () => {
  test("spreads large fleets across the heartbeat interval", () => {
    expect(getHeartbeatBatchSize(1000, 30_000)).toBe(34);
  });

  test("keeps small fleets responsive", () => {
    expect(getHeartbeatBatchSize(10, 30_000)).toBe(1);
  });

  test("handles empty fleets", () => {
    expect(getHeartbeatBatchSize(0, 30_000)).toBe(0);
  });
});
