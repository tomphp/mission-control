import { soundEnabled, toggleSound } from "../state/sound";

export function SoundToggle() {
  return (
    <button
      type="button"
      class="sound-toggle"
      data-enabled={soundEnabled.value}
      aria-pressed={soundEnabled.value}
      title={soundEnabled.value ? "Disable notification sound" : "Enable notification sound"}
      onClick={toggleSound}
    >
      <span aria-hidden="true">{soundEnabled.value ? "🔊" : "🔇"}</span>
      <span class="sound-toggle__label">Sound</span>
    </button>
  );
}
