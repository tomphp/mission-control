import { describe, expect, test } from "bun:test";
import { now, startClock } from "./clock";

describe("startClock", () => {
  test("advances now.value on each tick and stops on cleanup", async () => {
    const before = now.value;
    const stop = startClock(5);

    await new Promise((resolve) => setTimeout(resolve, 20));
    const afterTicks = now.value;
    expect(afterTicks).toBeGreaterThan(before);

    stop();
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(now.value).toBe(afterTicks);
  });
});
