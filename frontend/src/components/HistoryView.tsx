import { useCallback, useEffect, useState } from "react";
import {
  QRService,
  type HistoryEntry,
} from "../lib/api";

type Props = { onToast: (msg: string, err: boolean) => void };

export default function HistoryView({ onToast }: Props) {
  const [entries, setEntries] = useState<HistoryEntry[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [preview, setPreview] = useState<{ url: string; content: string } | null>(
    null,
  );
  const [busy, setBusy] = useState(false);

  const load = useCallback(async () => {
    try {
      const list = await QRService.ListHistory();
      setEntries(list || []);
    } catch (err) {
      onToast(String(err), true);
    }
  }, [onToast]);

  useEffect(() => {
    load();
  }, [load]);

  return (
    <div className="panel">
      <span className="panel-title">
        HISTORY :: {entries.length} ENTRIES (PINNED SURVIVE AUTO-PRUNE)
      </span>
      <div className="row" style={{ marginBottom: 12 }}>
        <input
          type="text"
          value={searchQuery}
          placeholder="FILTER BY PAYLOAD OR TYPE…"
          onChange={(e) => setSearchQuery(e.target.value)}
          style={{ flex: 1 }}
        />
        <button className="btn" onClick={load} disabled={busy}>
          REFRESH
        </button>
        <button
          className="btn"
          disabled={busy}
          onClick={async () => {
            if (!window.confirm("Delete ALL unpinned history?")) return;
            await QRService.ClearHistory("unpinned");
            load();
            onToast("unpinned entries cleared", false);
          }}
        >
          CLEAR UNPINNED
        </button>
        <button
          className="btn"
          disabled={busy}
          onClick={async () => {
            if (!window.confirm("Delete ENTIRE history?")) return;
            await QRService.ClearHistory("all");
            load();
            onToast("history wiped", false);
          }}
        >
          WIPE ALL
        </button>
      </div>

      {entries.length === 0 && (
        <div className="hint" style={{ padding: 30, textAlign: "center" }}>
          NO HISTORY YET_
        </div>
      )}

      {(() => {
        const q = searchQuery.trim().toLowerCase();
        const shown = [...entries]
          .filter(
            (e) =>
              !q ||
              e.content.toLowerCase().includes(q) ||
              String(e.content_type).toLowerCase().includes(q) ||
              String(e.kind).toLowerCase().includes(q),
          )
          .sort((a, b) =>
            a.pinned === b.pinned
              ? b.timestamp.localeCompare(a.timestamp)
              : a.pinned
                ? -1
                : 1,
          );
        if (shown.length === 0) {
          return (
            <div className="hint" style={{ padding: 30, textAlign: "center" }}>
              {q && entries.length > 0
                ? "NO MATCHES FOR FILTER_"
                : "NO HISTORY YET_"}
            </div>
          );
        }
        return shown.map((e) => (
          <div className="history-item" key={e.id}>
            <div style={{ flex: 1, minWidth: 0 }}>
              <div className="row" style={{ gap: 6, marginBottom: 4 }}>
                <span className="badge">{e.kind === "scan" ? "SCAN" : "GEN"}</span>
                <span className="badge meta">{e.content_type}</span>
                {e.error_level && (
                  <span className="badge meta">ECC {e.error_level}</span>
                )}
              </div>
              <div className="content-preview">{e.content}</div>
              <div className="when">
                {new Date(e.timestamp).toLocaleString()}
                {e.source ? ` · via ${e.source}` : ""}
              </div>
            </div>
            <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
              <button
                className={"icon-btn" + (e.pinned ? " pinned" : "")}
                title={e.pinned ? "Unpin" : "Pin"}
                onClick={async () => {
                  await QRService.SetPinned(e.id, !e.pinned);
                  load();
                }}
              >
                ★
              </button>
              <button
                className="icon-btn"
                title="Copy content"
                onClick={() => {
                  navigator.clipboard.writeText(e.content);
                  onToast("copied", false);
                }}
              >
                COPY
              </button>
              {e.kind === "generate" && (
                <button
                  className="icon-btn"
                  title="Regenerate preview with saved style"
                  onClick={async () => {
                    try {
                      setBusy(true);
                      const url = await QRService.RegenerateFromHistory(e.id);
                      setPreview({ url, content: e.content });
                    } catch (err) {
                      onToast(String(err), true);
                    } finally {
                      setBusy(false);
                    }
                  }}
                >
                  IMG
                </button>
              )}
              <button
                className="icon-btn"
                title="Delete"
                onClick={async () => {
                  await QRService.DeleteHistory(e.id);
                  load();
                }}
              >
                DEL
              </button>
            </div>
          </div>
        ));
      })()}

      {preview && (
        <div
          className="overlay"
          style={{
            position: "fixed",
            inset: 0,
            zIndex: 50,
            background: "rgba(0,0,0,0.88)",
          }}
          onClick={() => setPreview(null)}
        >
          <h1>REGENERATED</h1>
          <div className="qr-preview">
            <img src={preview.url} alt="regenerated QR" />
          </div>
          <div className="payload-box" style={{ maxWidth: 480 }}>
            {preview.content}
          </div>
          <div className="hint">(click anywhere to close)</div>
        </div>
      )}
    </div>
  );
}
