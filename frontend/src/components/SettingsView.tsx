import { useEffect, useState } from "react";
import { QRService, type Settings, type Preset } from "../lib/api";

type Theme = "dark" | "light";

function applyTheme(theme: Theme) {
  const root = document.documentElement;
  if (theme === "light") {
    root.style.setProperty("--bg", "#f5f5f5");
    root.style.setProperty("--bg2", "#ffffff");
    root.style.setProperty("--panel", "#ffffff");
    root.style.setProperty("--panel-2", "#f0f0f0");
    root.style.setProperty("--fg", "#1a1a1a");
    root.style.setProperty("--dim", "#666666");
    root.style.setProperty("--accent", "#00aa66");
    root.style.setProperty("--accent-dim", "rgba(0, 170, 102, 0.12)");
    root.style.setProperty("--accent-glow", "rgba(0, 170, 102, 0.18)");
    root.style.setProperty("--accent-overlay", "rgba(0, 170, 102, 0.06)");
    root.style.setProperty("--accent-shadow", "rgba(0, 170, 102, 0.20)");
    root.style.setProperty("--line", "#d0d0d0");
    root.style.setProperty("--payload-text", "#0a5c36");
    root.style.setProperty("--toast-bg", "#e6f7ef");
    root.style.setProperty("--toast-err-bg", "#fde8e8");
    root.style.setProperty("--btn-text", "#04140c");
    root.style.setProperty("--success", "#00aa66");
    root.style.setProperty("--success-dim", "rgba(0, 170, 102, 0.12)");
    root.style.setProperty("--grade-aaa-bg", "#e0f5ec");
    root.style.setProperty("--grade-aa-bg", "#fdf3dc");
    root.style.setProperty("--grade-fail-bg", "#fde8e8");
    root.style.setProperty("--danger", "#cc3344");
    root.style.setProperty("--danger-dim", "rgba(204, 51, 68, 0.08)");
    root.style.setProperty("--warn", "#cc8800");
    root.style.setProperty("--warn-dim", "rgba(204, 136, 0, 0.08)");
  } else {
    root.style.setProperty("--bg", "#0a0a0b");
    root.style.setProperty("--bg2", "#121214");
    root.style.setProperty("--panel", "#121214");
    root.style.setProperty("--panel-2", "#17171a");
    root.style.setProperty("--fg", "#e8e8ea");
    root.style.setProperty("--dim", "#8b8b93");
    root.style.setProperty("--accent", "#00ff99");
    root.style.setProperty("--accent-dim", "rgba(0, 255, 153, 0.14)");
    root.style.setProperty("--accent-glow", "rgba(0, 255, 153, 0.22)");
    root.style.setProperty("--accent-overlay", "rgba(0, 255, 153, 0.05)");
    root.style.setProperty("--accent-shadow", "rgba(0, 255, 153, 0.25)");
    root.style.setProperty("--line", "#2a2a30");
    root.style.setProperty("--payload-text", "#d5ffe9");
    root.style.setProperty("--toast-bg", "#0d1f16");
    root.style.setProperty("--toast-err-bg", "#210d10");
    root.style.setProperty("--btn-text", "#04140c");
    root.style.setProperty("--success", "#00ff88");
    root.style.setProperty("--success-dim", "rgba(0, 255, 136, 0.14)");
    root.style.setProperty("--grade-aaa-bg", "#104020");
    root.style.setProperty("--grade-aa-bg", "#403010");
    root.style.setProperty("--grade-fail-bg", "#401010");
    root.style.setProperty("--danger", "#ff5566");
    root.style.setProperty("--danger-dim", "rgba(255, 85, 102, 0.06)");
    root.style.setProperty("--warn", "#ffc857");
    root.style.setProperty("--warn-dim", "rgba(255, 200, 87, 0.07)");
  }
  localStorage.setItem("pixalpeek_theme", theme);
}

function loadTheme(): Theme {
  return (localStorage.getItem("pixalpeek_theme") as Theme) || "dark";
}

type Props = { onToast: (msg: string, err: boolean) => void };

export default function SettingsView({ onToast }: Props) {
  const [s, setS] = useState<Settings | null>(null);
  const [version, setVersion] = useState("");
  const [theme, setTheme] = useState<Theme>(loadTheme);

  useEffect(() => {
    QRService.GetSettings().then(setS).catch(() => {});
    QRService.AppVersion().then(setVersion).catch(() => {});
  }, []);

  useEffect(() => {
    applyTheme(theme);
  }, [theme]);

  if (!s)
    return (
      <div className="panel">
        <span className="panel-title">SETTINGS</span>
        <div className="hint" style={{ padding: 30 }}>LOADING_</div>
      </div>
    );

  const upd = (patch: Partial<Settings>) => setS({ ...s, ...patch });

  return (
    <div className="grid-2">
      <div className="panel">
        <span className="panel-title">GENERATOR DEFAULTS</span>
        <label className="fld">
          <span>Default size (px)</span>
          <input
            type="number"
            min={128}
            max={2048}
            value={s.size}
            onChange={(e) => upd({ size: Number(e.target.value) })}
          />
        </label>
        <div className="grid-2" style={{ gap: 8 }}>
          <label className="fld">
            <span>Quiet margin</span>
            <input
              type="number"
              min={0}
              max={8}
              value={s.margin}
              onChange={(e) => upd({ margin: Number(e.target.value) })}
            />
          </label>
          <label className="fld">
            <span>Error correction</span>
            <select
              value={s.default_ecc}
              onChange={(e) => upd({ default_ecc: e.target.value })}
            >
              {["L", "M", "Q", "H"].map((x) => (
                <option key={x} value={x}>{x}</option>
              ))}
            </select>
          </label>
          <label className="fld">
            <span>Shape</span>
            <select
              value={s.shape}
              onChange={(e) => upd({ shape: e.target.value })}
            >
              <option value="square">SQUARE</option>
              <option value="rounded">ROUNDED</option>
              <option value="dot">DOT</option>
            </select>
          </label>
          <label className="fld">
            <span>Format</span>
            <select
              value={s.default_format}
              onChange={(e) => upd({ default_format: e.target.value })}
            >
              <option value="png">PNG</option>
              <option value="jpg">JPEG</option>
              <option value="svg">SVG</option>
              <option value="pdf">PDF</option>
            </select>
          </label>
        </div>
      </div>

      <div className="panel">
        <span className="panel-title">SECURITY</span>
        <div className="security-toggle">
          <div className="security-toggle-info">
            <strong>OFFLINE URL SAFETY HEURISTICS</strong>
            <span className="hint">Flags shorteners, raw IPs, punycode, and suspicious TLDs without network calls</span>
          </div>
          <label className="switch">
            <input
              type="checkbox"
              checked={s.check_url_safety}
              onChange={(e) => upd({ check_url_safety: e.target.checked })}
            />
            <span className="slider"></span>
          </label>
        </div>

        <span className="panel-title" style={{ marginTop: 20 }}>APP</span>
        <div className="kv"><span className="k">VERSION</span><span>{version}</span></div>
        <div className="kv">
          <span className="k">SAVED PRESETS</span>
          <span>{(s.presets || []).map((p: Preset) => p.name).join(", ") || "—"}</span>
        </div>
        <div className="kv">
          <span className="k">THEME</span>
          <span className="row" style={{ gap: 6 }}>
            {(["dark", "light"] as Theme[]).map((t) => (
              <button
                key={t}
                className={"btn" + (theme === t ? " primary" : "")}
                style={{ padding: "3px 10px", fontSize: 10 }}
                onClick={() => setTheme(t)}
              >
                {t.toUpperCase()}
              </button>
            ))}
          </span>
        </div>

        <button
          className="btn primary"
          style={{ marginTop: 16 }}
          onClick={async () => {
            try {
              await QRService.SaveSettings(s);
              onToast("settings saved", false);
            } catch (err) {
              onToast(String(err), true);
            }
          }}
        >
          SAVE SETTINGS
        </button>

        <p className="hint" style={{ marginTop: 18 }}>
          CLI CHEATSHEET<br />
          ────────────────────────────────<br />
          pixalpeek -g "text" -o qr.png<br />
          pixalpeek -qr photo.png --multi<br />
          pixalpeek --batch codes.csv --zip<br />
          pixalpeek --scan-dir ./images<br />
          pixalpeek -qr img.png -o -
        </p>
      </div>
    </div>
  );
}
