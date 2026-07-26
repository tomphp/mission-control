import { signal } from "@preact/signals";

// A shared ticking clock so components rendering relative timestamps (e.g.
// "3m ago") re-render as time passes, not just when session data changes.
export const now = signal(Date.now());

export function startClock(intervalMs = 5000): () => void {
  const id = setInterval(() => {
    now.value = Date.now();
  }, intervalMs);
  return () => clearInterval(id);
}
