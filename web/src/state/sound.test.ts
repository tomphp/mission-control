import { describe, expect, test, beforeEach } from "bun:test";
import { soundEnabled, setSoundEnabled, toggleSound, notifyStateChange } from "./sound";

beforeEach(() => {
  setSoundEnabled(true);
});

describe("setSoundEnabled / toggleSound", () => {
  test("updates the soundEnabled signal", () => {
    setSoundEnabled(false);
    expect(soundEnabled.value).toBe(false);

    toggleSound();
    expect(soundEnabled.value).toBe(true);
  });
});

describe("notifyStateChange", () => {
  test("does not throw when transitioning into a notifiable state", () => {
    expect(() => notifyStateChange("working", "idle")).not.toThrow();
    expect(() => notifyStateChange("working", "waiting_for_input")).not.toThrow();
  });

  test("does not throw for a no-op transition", () => {
    expect(() => notifyStateChange("idle", "idle")).not.toThrow();
  });

  test("does not throw for a transition into a non-notifiable state", () => {
    expect(() => notifyStateChange("idle", "working")).not.toThrow();
  });

  test("does not throw while sound is disabled", () => {
    setSoundEnabled(false);
    expect(() => notifyStateChange("working", "idle")).not.toThrow();
  });
});
