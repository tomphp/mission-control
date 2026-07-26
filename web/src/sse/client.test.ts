import { describe, expect, test, mock } from "bun:test";
import { SSEClient, type EventSourceLike } from "./client";
import type { Session } from "../types/session";

type Listener = (event: { data: string }) => void;

class FakeEventSource implements EventSourceLike {
  listeners = new Map<string, Listener[]>();
  closed = false;

  addEventListener(type: string, listener: Listener): void {
    const list = this.listeners.get(type) ?? [];
    list.push(listener);
    this.listeners.set(type, list);
  }

  close(): void {
    this.closed = true;
  }

  emit(type: string, data: unknown): void {
    for (const listener of this.listeners.get(type) ?? []) {
      listener({ data: JSON.stringify(data) });
    }
  }
}

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

function makeSink() {
  return {
    snapshot: mock((_sessions: Session[]) => {}),
    upsert: mock((_session: Session) => {}),
    remove: mock((_id: string) => {}),
  };
}

describe("SSEClient", () => {
  test("connect opens exactly one EventSource against the given url", () => {
    const sources: FakeEventSource[] = [];
    const factory = (url: string) => {
      expect(url).toBe("http://localhost:1234/events");
      const es = new FakeEventSource();
      sources.push(es);
      return es;
    };

    const client = new SSEClient("http://localhost:1234/events", makeSink(), factory);
    client.connect();

    expect(sources.length).toBe(1);
  });

  test("dispatches snapshot events to sink.snapshot", () => {
    let source!: FakeEventSource;
    const sink = makeSink();
    const client = new SSEClient("url", sink, () => (source = new FakeEventSource()));
    client.connect();

    const list = [makeSession({ id: "a" }), makeSession({ id: "b" })];
    source.emit("snapshot", list);

    expect(sink.snapshot).toHaveBeenCalledTimes(1);
    expect(sink.snapshot.mock.calls[0][0]).toEqual(list);
  });

  test("dispatches update events to sink.upsert", () => {
    let source!: FakeEventSource;
    const sink = makeSink();
    const client = new SSEClient("url", sink, () => (source = new FakeEventSource()));
    client.connect();

    const s = makeSession({ id: "a", state: "waiting_for_input" });
    source.emit("update", s);

    expect(sink.upsert).toHaveBeenCalledTimes(1);
    expect(sink.upsert.mock.calls[0][0]).toEqual(s);
  });

  test("dispatches remove events to sink.remove", () => {
    let source!: FakeEventSource;
    const sink = makeSink();
    const client = new SSEClient("url", sink, () => (source = new FakeEventSource()));
    client.connect();

    source.emit("remove", { id: "a" });

    expect(sink.remove).toHaveBeenCalledTimes(1);
    expect(sink.remove.mock.calls[0][0]).toBe("a");
  });

  test("reconnects with backoff after an error, opening a new EventSource", async () => {
    const sources: FakeEventSource[] = [];
    const factory = () => {
      const es = new FakeEventSource();
      sources.push(es);
      return es;
    };

    const client = new SSEClient("url", makeSink(), factory, { baseDelayMs: 5, maxDelayMs: 20 });
    client.connect();
    expect(sources.length).toBe(1);

    sources[0].emit("error", null);
    expect(sources[0].closed).toBe(true);

    await new Promise((r) => setTimeout(r, 50));
    expect(sources.length).toBe(2);
  });

  test("close() stops reconnection attempts", async () => {
    const sources: FakeEventSource[] = [];
    const factory = () => {
      const es = new FakeEventSource();
      sources.push(es);
      return es;
    };

    const client = new SSEClient("url", makeSink(), factory, { baseDelayMs: 5, maxDelayMs: 20 });
    client.connect();
    sources[0].emit("error", null);
    client.close();

    await new Promise((r) => setTimeout(r, 50));
    expect(sources.length).toBe(1);
  });
});
