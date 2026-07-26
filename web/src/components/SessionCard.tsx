import type { Session, SessionState } from "../types/session";
import { now } from "../state/clock";

const STATE_LABEL: Record<SessionState, string> = {
  working: "Working",
  waiting_for_input: "Waiting for input",
  idle: "Idle",
};

function relativeTime(iso: string): string {
  const deltaSec = Math.max(0, (now.value - new Date(iso).getTime()) / 1000);
  if (deltaSec < 5) return "just now";
  if (deltaSec < 60) return `${Math.floor(deltaSec)}s ago`;
  if (deltaSec < 3600) return `${Math.floor(deltaSec / 60)}m ago`;
  return `${Math.floor(deltaSec / 3600)}h ago`;
}

export function SessionCard({ session }: { session: Session }) {
  return (
    <li class={`session-card session-card--${session.state}`}>
      <div class="session-card__header">
        <span class="session-card__folder">{session.folder}</span>
        {session.gitBranch && <span class="session-card__branch">{session.gitBranch}</span>}
      </div>
      <span class="state-badge" data-state={session.state}>
        {STATE_LABEL[session.state]}
      </span>
      <div class="session-card__meta">
        <span class="session-card__path" title={session.cwd}>
          {session.cwd}
        </span>
        <span class="session-card__last-event">
          {session.lastEvent} · {relativeTime(session.lastEventAt)}
        </span>
      </div>
    </li>
  );
}
