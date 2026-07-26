import { sessions } from "../state/sessionSignal";
import { SessionCard } from "./SessionCard";

export function SessionList() {
  const list = [...sessions.value.values()].sort(
    (a, b) => a.folder.localeCompare(b.folder) || a.id.localeCompare(b.id),
  );

  if (list.length === 0) {
    return <p class="empty-state">No active Claude Code sessions.</p>;
  }

  return (
    <ul class="session-list">
      {list.map((s) => (
        <SessionCard key={s.id} session={s} />
      ))}
    </ul>
  );
}
