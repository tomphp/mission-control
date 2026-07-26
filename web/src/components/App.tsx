import { SessionList } from "./SessionList";
import { SoundToggle } from "./SoundToggle";

export function App() {
  return (
    <main>
      <header class="app-header">
        <div>
          <h1>Mission Control</h1>
          <p class="app-subtitle">Live Claude Code sessions</p>
        </div>
        <SoundToggle />
      </header>
      <SessionList />
    </main>
  );
}
