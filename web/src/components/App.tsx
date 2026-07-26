import { SessionList } from "./SessionList";

export function App() {
  return (
    <main>
      <header class="app-header">
        <h1>Mission Control</h1>
        <p class="app-subtitle">Live Claude Code sessions</p>
      </header>
      <SessionList />
    </main>
  );
}
