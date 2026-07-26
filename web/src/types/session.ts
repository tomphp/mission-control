export type SessionState = "working" | "waiting_for_input" | "idle";

export interface Session {
  id: string;
  cwd: string;
  folder: string;
  gitBranch: string;
  state: SessionState;
  startedAt: string;
  lastEventAt: string;
  lastEvent: string;
}
