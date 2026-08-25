import { useState, useCallback } from "react";
// @ts-ignore
import { Window } from "/wails/runtime.js";
import {
  LogoIcon,
  CloseIcon,
  MinimizeIcon,
  MaximizeIcon,
  ScanIcon,
  GenerateIcon,
  HistoryIcon,
  SettingsIcon,
  SplitIcon,
} from "./Icons";

type Tab = "scan" | "generate" | "history" | "settings" | "workbench";

interface TabDef {
  id: Tab;
  label: string;
  key: string;
  icon: React.ReactNode;
  wideOnly?: boolean;
}

export const ALL_TABS: TabDef[] = [
  { id: "scan", label: "SCAN", key: "1", icon: <ScanIcon size={14} /> },
  { id: "generate", label: "GENERATE", key: "2", icon: <GenerateIcon size={14} /> },
  { id: "history", label: "HISTORY", key: "3", icon: <HistoryIcon size={14} /> },
  { id: "settings", label: "SETTINGS", key: "4", icon: <SettingsIcon size={14} /> },
  { id: "workbench", label: "SPLIT", key: "5", icon: <SplitIcon size={14} />, wideOnly: true },
];

interface Props {
  tab: Tab;
  onTabChange: (tab: Tab) => void;
  viewport: number;
}

export default function TitleBar({ tab, onTabChange, viewport }: Props) {
  const [maximized, setMaximized] = useState(false);

  const handleDoubleClick = useCallback(async () => {
    try {
      const w = Window.Get();
      const isMax = await w.IsMaximised();
      if (isMax) await w.UnMaximise();
      else await w.Maximise();
      setMaximized(!isMax);
    } catch {}
  }, []);

  const handleMinimize = useCallback(async () => {
    try { await Window.Get().Minimise(); } catch {}
  }, []);

  const handleMaximize = useCallback(async () => {
    try {
      const w = Window.Get();
      const isMax = await w.IsMaximised();
      if (isMax) await w.UnMaximise();
      else await w.Maximise();
      setMaximized(!isMax);
    } catch {}
  }, []);

  const handleClose = useCallback(async () => {
    try { await Window.Get().Close(); } catch {}
  }, []);

  const tabs = ALL_TABS.filter((t) => !t.wideOnly || viewport >= 1200);

  return (
    <header className="titlebar" onDoubleClick={handleDoubleClick}>
      <div className="titlebar-left">
        <div className="window-controls">
          <button className="window-btn close" onClick={handleClose} title="Close">
            <CloseIcon size={10} />
          </button>
          <button className="window-btn minimize" onClick={handleMinimize} title="Minimize">
            <MinimizeIcon size={10} />
          </button>
          <button className="window-btn maximize" onClick={handleMaximize} title={maximized ? "Restore" : "Maximize"}>
            <MaximizeIcon size={10} />
          </button>
        </div>
      </div>

      <div className="titlebar-brand">
        <LogoIcon size={18} className="titlebar-logo" />
        <span className="titlebar-title">PIXALPEEK</span>
      </div>

      <div className="titlebar-spacer" />

      <nav className="titlebar-tabs">
        {tabs.map((t) => (
          <button
            key={t.id}
            className={"titlebar-tab" + (tab === t.id ? " active" : "")}
            onClick={() => onTabChange(t.id)}
            title={t.label}
          >
            {t.icon}
            <span className="tab-text">{t.label}</span>
          </button>
        ))}
      </nav>
    </header>
  );
}
