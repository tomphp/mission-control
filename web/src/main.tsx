import { render } from "preact";
import { App } from "./components/App";
import { SSEClient } from "./sse/client";
import { snapshot, upsert, remove } from "./state/sessionSignal";

const client = new SSEClient(`${location.origin}/events`, { snapshot, upsert, remove });
client.connect();

const root = document.getElementById("app");
if (root) {
  render(<App />, root);
}
