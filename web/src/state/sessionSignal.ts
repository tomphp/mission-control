import { signal } from "@preact/signals";
import type { Session } from "../types/session";

export const sessions = signal<Map<string, Session>>(new Map());

export function snapshot(list: Session[]): void {
  sessions.value = new Map(list.map((s) => [s.id, s]));
}

export function upsert(session: Session): void {
  const next = new Map(sessions.value);
  next.set(session.id, session);
  sessions.value = next;
}

export function remove(id: string): void {
  if (!sessions.value.has(id)) return;
  const next = new Map(sessions.value);
  next.delete(id);
  sessions.value = next;
}
