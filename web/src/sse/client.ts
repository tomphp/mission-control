import type { Session } from "../types/session";

export interface EventSourceLike {
  addEventListener(type: string, listener: (event: { data: string }) => void): void;
  close(): void;
}

export type EventSourceFactory = (url: string) => EventSourceLike;

export interface SessionSink {
  snapshot(sessions: Session[]): void;
  upsert(session: Session): void;
  remove(id: string): void;
}

export interface SSEClientOptions {
  baseDelayMs?: number;
  maxDelayMs?: number;
}

const defaultFactory: EventSourceFactory = (url) => new EventSource(url) as unknown as EventSourceLike;

/**
 * SSEClient consumes the server's session event stream and dispatches each
 * message to a SessionSink. The EventSource constructor is injected (not
 * read off globalThis) so it can be faked in tests.
 */
export class SSEClient {
  private es: EventSourceLike | null = null;
  private reconnectAttempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private stopped = true;
  private readonly baseDelayMs: number;
  private readonly maxDelayMs: number;

  constructor(
    private readonly url: string,
    private readonly sink: SessionSink,
    private readonly createEventSource: EventSourceFactory = defaultFactory,
    options: SSEClientOptions = {},
  ) {
    this.baseDelayMs = options.baseDelayMs ?? 1000;
    this.maxDelayMs = options.maxDelayMs ?? 30000;
  }

  connect(): void {
    this.stopped = false;
    this.openConnection();
  }

  close(): void {
    this.stopped = true;
    if (this.reconnectTimer !== null) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.es?.close();
    this.es = null;
  }

  private openConnection(): void {
    const es = this.createEventSource(this.url);
    this.es = es;

    es.addEventListener("snapshot", (event) => {
      this.reconnectAttempts = 0;
      this.sink.snapshot(JSON.parse(event.data) as Session[]);
    });
    es.addEventListener("update", (event) => {
      this.sink.upsert(JSON.parse(event.data) as Session);
    });
    es.addEventListener("remove", (event) => {
      const { id } = JSON.parse(event.data) as { id: string };
      this.sink.remove(id);
    });
    es.addEventListener("error", () => {
      es.close();
      this.scheduleReconnect();
    });
  }

  private scheduleReconnect(): void {
    if (this.stopped) return;
    const delay = Math.min(this.baseDelayMs * 2 ** this.reconnectAttempts, this.maxDelayMs);
    this.reconnectAttempts++;
    this.reconnectTimer = setTimeout(() => this.openConnection(), delay);
  }
}
