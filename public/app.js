const output = document.getElementById("output");
const runState = document.getElementById("run-state");
const historyList = document.getElementById("history");

const history = [];
const MAX_HISTORY = 6;

function statusClass(status) {
  if (status >= 500) return "pill-error";
  if (status >= 400) return "pill-warn";
  if (status >= 200) return "pill-ok";
  return "";
}

function renderHistory() {
  if (history.length === 0) {
    historyList.innerHTML = '<span class="empty-history">No tests in this session yet.</span>';
    return;
  }

  historyList.innerHTML = history
    .map(
      entry => `
        <button class="history-row" data-json="${encodeURIComponent(JSON.stringify(entry.payload, null, 2))}">
          <span>${entry.label}</span>
          <span class="pill ${statusClass(entry.status)}">${entry.status}</span>
        </button>
      `
    )
    .join("");

  historyList.querySelectorAll(".history-row").forEach(row => {
    row.addEventListener("click", () => {
      output.textContent = decodeURIComponent(row.dataset.json);
    });
  });
}

async function report(label, promise, trigger) {
  if (trigger) trigger.disabled = true;
  runState.textContent = "SENDING…";
  runState.className = "status-pill";

  try {
    const startedAt = performance.now();
    const response = await promise;
    const durationMs = Math.round(performance.now() - startedAt);
    const body = await response.json().catch(() => ({}));
    const payload = { receivedStatus: response.status, durationMs, body };

    runState.textContent = `${response.status}`;
    runState.className = `status-pill ${statusClass(response.status)}`;
    output.textContent = JSON.stringify(payload, null, 2);

    history.unshift({ label, status: response.status, payload });
    history.length = Math.min(history.length, MAX_HISTORY);
    renderHistory();
  } catch (error) {
    runState.textContent = "ERROR";
    runState.className = "status-pill pill-error";
    output.textContent = String(error);
  } finally {
    if (trigger) trigger.disabled = false;
  }
}

document.querySelectorAll("#status-buttons button").forEach(button => {
  button.addEventListener("click", () => {
    report(`HTTP ${button.dataset.status}`, fetch(`/demo/http/${button.dataset.status}`), button);
  });
});
