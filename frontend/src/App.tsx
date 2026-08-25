import { useEffect, useRef, useState } from "react";
// @ts-ignore
import { Events, Window } from "/wails/runtime.js";
import TitleBar, { ALL_TABS } from "./components/TitleBar";
import ContextMenu from "./components/ContextMenu";
import ScannerView from "./components/ScannerView";
import GeneratorView from "./components/GeneratorView";
import HistoryView from "./components/HistoryView";
import SettingsView from "./components/SettingsView";
import WorkerOverlay from "./components/WorkerOverlay";
import { taskFromURL } from "./lib/worker";

type Tab = "scan" | "generate" | "history" | "settings" | "workbench";

export default function App() {
  const [tab, setTab] = useState<Tab>("scan");
  const [task] = useState(() => taskFromURL());
  const [toast, setToast] = useState<{ msg: string; err: boolean } | null>(
    null,
  );
  const toastTimer = useRef<number | undefined>(undefined);
  const [viewport, setViewport] = useState(window.innerWidth);

  useEffect(() => {
    if (!task) return;
    document.title = "PIXALPEEK";
    return;
  }, [task]);

  useEffect(() => {
    const onResize = () => setViewport(window.innerWidth);
    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, []);

  useEffect(() => {
    const unsub = Events.On(
      "pixalpeek:scan-result",
      (data: unknown) => {
        const d = data as { content?: string };
        if (d?.content) {
          navigator.clipboard.writeText(d.content).catch(() => {});
          showToast(`tray scan complete — payload copied`, false);
        }
      },
    );
    return unsub;
  }, []);

  useEffect(() => {
    const unsub = Events.On(
      "pixalpeek:toast",
      (data: unknown) => {
        const d = data as { message?: string; error?: string };
        showToast(d?.message || d?.error || "notification", !!d?.error);
      },
    );
    return unsub;
  }, []);

  useEffect(() => {
    const handler = (e: Event) => {
      const detail = (e as CustomEvent<{ message?: string }>).detail;
      showToast(detail?.message || "done", false);
    };
    window.addEventListener("pixalpeek-toast", handler);
    return () => window.removeEventListener("pixalpeek-toast", handler);
  }, []);

  const showToast = (msg: string, err: boolean) => {
    setToast({ msg, err });
    window.clearTimeout(toastTimer.current);
    toastTimer.current = window.setTimeout(() => setToast(null), 3200);
  };

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (task) return;
      if (e.key === "F11") {
        e.preventDefault();
        try {
          if (document.fullscreenElement) {
            document.exitFullscreen();
          } else {
            document.documentElement.requestFullscreen();
          }
        } catch {}
        return;
      }
      if (e.ctrlKey && e.key === "w") {
        e.preventDefault();
        try { Window.Get().Close(); } catch {}
        return;
      }
      if (e.ctrlKey || e.altKey || e.metaKey) return;
      const target = e.target as HTMLElement;
      if (
        target.tagName === "INPUT" ||
        target.tagName === "TEXTAREA" ||
        target.tagName === "SELECT"
      )
        return;
      const found = ALL_TABS.find((t) => t.key === e.key);
      if (found && !found.wideOnly) setTab(found.id);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [task]);

  if (task) return <WorkerOverlay task={task} />;

  const pasteFromClipboard = async () => {
    try {
      const text = await navigator.clipboard.readText();
      const el = document.activeElement as HTMLInputElement | HTMLTextAreaElement | null;
      if (el && (el.tagName === "INPUT" || el.tagName === "TEXTAREA")) {
        const start = el.selectionStart ?? el.value.length;
        const end = el.selectionEnd ?? el.value.length;
        el.value = el.value.slice(0, start) + text + el.value.slice(end);
        el.selectionStart = el.selectionEnd = start + text.length;
        el.dispatchEvent(new Event("input", { bubbles: true }));
      } else {
        await navigator.clipboard.writeText(text);
      }
    } catch {}
  };

  const ctxItems = [
    { label: "Cut", shortcut: "Ctrl+X", action: () => document.execCommand("cut") },
    { label: "Copy", shortcut: "Ctrl+C", action: () => document.execCommand("copy") },
    { label: "Paste", shortcut: "Ctrl+V", action: pasteFromClipboard },
    { divider: true, label: "", action: () => {} },
    { label: "Select All", shortcut: "Ctrl+A", action: () => document.execCommand("selectAll") },
  ];

  const eccLabel = "REED_SOLOMON_2D";
  const now = new Date();
  const ts = `${String(now.getHours()).padStart(2, "0")}:${String(now.getMinutes()).padStart(2, "0")}`;

  return (
    <div className="app">
      <TitleBar tab={tab} onTabChange={setTab} viewport={viewport} />
      <main className="content">
        {tab === "workbench" && viewport >= 1200 ? (
          <div className="split-workbench">
            <ScannerView onToast={showToast} />
            <GeneratorView onToast={showToast} />
          </div>
        ) : (
          <>
            {tab === "scan" && <ScannerView onToast={showToast} />}
            {tab === "generate" && <GeneratorView onToast={showToast} />}
            {tab === "history" && <HistoryView onToast={showToast} />}
            {tab === "settings" && <SettingsView onToast={showToast} />}
          </>
        )}
      </main>
      <footer className="status-footer">
        <div className="status-left">
          <span className="status-online">● SYS ONLINE</span>
          <span>VIEWPORT {viewport}px</span>
        </div>
        <div className="status-right">
          <span>{eccLabel}</span>
          <span>{ts}</span>
        </div>
      </footer>
      {toast && (
        <div className={"toast" + (toast.err ? " error" : "")}>
          {toast.err ? "✖ " : "✔ "}
          {toast.msg}
        </div>
      )}
      <ContextMenu items={ctxItems} />
    </div>
  );
}
