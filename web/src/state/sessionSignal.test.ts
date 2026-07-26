import { describe, expect, test, beforeEach } from "bun:test";
import { sessions, snapshot, upsert, remove } from "./sessionSignal";
import type { Session } from "../types/session";

function makeSession(overrides: Partial<Session> = {}): Session {
  return {
    id: "sess-1",
    cwd: "/tmp/proj",
    folder: "proj",
    gitBranch: "main",
    state: "working",
    startedAt: "2026-01-01T00:00:00Z",
    lastEventAt: "2026-01-01T00:00:00Z",
    lastEvent: "PreToolUse",
    ...overrides,
  };
}

beforeEach(() => {
  snapshot([]);
});

describe("snapshot", () => {
  test("replaces the entire session set", () => {
    upsert(makeSession({ id: "old" }));
    snapshot([makeSession({ id: "a" }), makeSession({ id: "b" })]);

    const ids = [...sessions.value.keys()].sort();
    expect(ids).toEqual(["a", "b"]);
  });

  test("empty snapshot clears all sessions", () => {
    upsert(makeSession({ id: "a" }));
    snapshot([]);
    expect(sessions.value.size).toBe(0);
  });
});

describe("upsert", () => {
  test("adds a new session", () => {
    upsert(makeSession({ id: "a" }));
    expect(sessions.value.get("a")?.state).toBe("working");
  });

  test("updates an existing session in place without affecting others", () => {
    upsert(makeSession({ id: "a", state: "working" }));
    upsert(makeSession({ id: "b", state: "idle" }));

    upsert(makeSession({ id: "a", state: "waiting_for_input" }));

    expect(sessions.value.get("a")?.state).toBe("waiting_for_input");
    expect(sessions.value.get("b")?.state).toBe("idle");
    expect(sessions.value.size).toBe(2);
  });
});

describe("remove", () => {
  test("removes a session by id", () => {
    upsert(makeSession({ id: "a" }));
    upsert(makeSession({ id: "b" }));

    remove("a");

    expect(sessions.value.has("a")).toBe(false);
    expect(sessions.value.has("b")).toBe(true);
  });

  test("removing an unknown id is a no-op", () => {
    upsert(makeSession({ id: "a" }));
    remove("does-not-exist");
    expect(sessions.value.size).toBe(1);
  });
});
