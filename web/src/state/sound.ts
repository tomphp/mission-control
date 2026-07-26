import { signal } from "@preact/signals";
import type { SessionState } from "../types/session";
import notificationSound from "../assets/notification.wav";

const STORAGE_KEY = "mission-control:sound-enabled";

// States that warrant a nudge: the agent has stopped and is either done or
// blocked on the user, as opposed to still working.
const NOTIFY_STATES: ReadonlySet<SessionState> = new Set(["idle", "waiting_for_input"]);

function loadInitial(): boolean {
  if (typeof localStorage === "undefined") return true;
  const stored = localStorage.getItem(STORAGE_KEY);
  return stored === null ? true : stored === "true";
}

export const soundEnabled = signal<boolean>(loadInitial());

export function setSoundEnabled(enabled: boolean): void {
  soundEnabled.value = enabled;
  if (typeof localStorage !== "undefined") {
    localStorage.setItem(STORAGE_KEY, String(enabled));
  }
}

export function toggleSound(): void {
  setSoundEnabled(!soundEnabled.value);
}

export function notifyStateChange(previousState: SessionState, nextState: SessionState): void {
  if (previousState === nextState) return;
  if (!NOTIFY_STATES.has(nextState)) return;
  if (!soundEnabled.value) return;
  playSound();
}

function playSound(): void {
  if (typeof Audio === "undefined") return;
  const audio = new Audio(notificationSound);
  audio.play().catch(() => {
    // Ignore playback failures (e.g. browser autoplay restrictions).
  });
}
