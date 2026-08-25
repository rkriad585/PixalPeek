import { useEffect, useRef, useState } from "react";
import {
  QRService,
  TYPES,
  ContentType,
  type EncodeOptions,
  type QRType,
  type Settings,
  type Preset,
} from "../lib/api";
import { fileToDataURL } from "../lib/api";
import {
  contrastRatio,
  contrastGrade,
  contrastBadgeClass,
} from "../lib/contrast";

type Props = { onToast: (msg: string, err: boolean) => void };

export default function GeneratorView({ onToast }: Props) {
  const [type, setType] = useState<QRType>("text");
  const [fields, setFields] = useState<Record<string, string>>({});
  const [size, setSize] = useState(512);
  const [ecc, setEcc] = useState("M");
  const [fg, setFg] = useState("#000000");
  const [bg, setBg] = useState("#FFFFFF");
  const [shape, setShape] = useState("square");
  const [quiet, setQuiet] = useState(4);
  const [format, setFormat] = useState("png");
  const [logoB64, setLogoB64] = useState("");
  const [preview, setPreview] = useState("");
  const [warnings, setWarnings] = useState<string[]>([]);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [presets, setPresets] = useState<Preset[]>([]);
  const [presetName, setPresetName] = useState("");
  const [selectedPreset, setSelectedPreset] = useState("");
  const [showBatch, setShowBatch] = useState(false);
  const [batchBusy, setBatchBusy] = useState(false);
  const batchRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    QRService.ListPresets()
      .then((p: Preset[]) => setPresets(p || []))
      .catch(() => {});
  }, []);

  useEffect(() => {
    QRService.GetSettings()
      .then((s: Settings) => {
        if (s.size >= 128) setSize(s.size);
        if (s.default_ecc) setEcc(s.default_ecc);
        if (s.shape) setShape(s.shape);
        if (s.default_format) setFormat(s.default_format);
      })
      .catch(() => {});
  }, []);

  const setF = (k: string, v: string) =>
    setFields((prev) => ({ ...prev, [k]: v }));

  const buildOpts = (content: string): EncodeOptions => ({
    content,
    type: type as ContentType,
    size,
    ecc,
    fg_color: fg,
    bg_color: bg,
    shape,
    logo_b64: logoB64,
    format,
    quiet_zone: quiet,
  });

  const generate = async () => {
    setBusy(true);
    setError("");
    setWarnings([]);
    try {
      const content = await QRService.BuildContent({
        type: type as ContentType,
        fields: Object.fromEntries(
          Object.entries(fields).filter(([, v]) => v !== ""),
        ),
      });
      const [dataURL, warns] = await QRService.GenerateWithWarnings(
        buildOpts(content),
      );
      setPreview(dataURL);
      setWarnings(warns || []);
      onToast("QR generated", false);
      QRService.AddHistory({
        id: "",
        kind: "generate" as any,
        content,
        content_type: type as any,
        timestamp: "",
        pinned: false,
        source: "generator",
        style: {
          format,
          ecc,
          fg_color: fg,
          bg_color: bg,
          shape,
          size,
          margin: quiet,
        },
      }).catch(() => {});
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  };

  const handleBatch = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setBatchBusy(true);
    try {
      const path = (file as unknown as { path?: string }).path || file.name;
      const [zipB64, count] = await QRService.ProcessBatchGUI(path, {
        content: "",
        type: type as ContentType,
        size,
        ecc,
        fg_color: fg,
        bg_color: bg,
        shape,
        format,
        quiet_zone: quiet,
      });
      const bytes = Uint8Array.from(atob(zipB64), (c) => c.charCodeAt(0));
      const blob = new Blob([bytes], { type: "application/zip" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = "pixalpeek_batch.zip";
      a.click();
      URL.revokeObjectURL(url);
      onToast(`batch complete: ${count} QR codes generated`, false);
      setShowBatch(false);
    } catch (err) {
      onToast(err instanceof Error ? err.message : String(err), true);
    } finally {
      setBatchBusy(false);
      if (batchRef.current) batchRef.current.value = "";
    }
  };

  return (
    <div className="grid-2">
      <div>
        <div className="panel">
          <span className="panel-title">CONTENT</span>
          <label className="fld">
            <span>Type</span>
            <select value={type} onChange={(e) => setType(e.target.value as QRType)}>
              {TYPES.map((t) => (
                <option key={t} value={t}>{t.toUpperCase()}</option>
              ))}
            </select>
          </label>
          {renderFields(type, fields, setF)}
        </div>

        <div className="panel">
          <span className="panel-title">STYLE</span>
          <div className="grid-2" style={{ gap: 8 }}>
            <label className="fld">
              <span>Size ({size}px)</span>
              <input
                type="range"
                min={128}
                max={2048}
                step={32}
                value={size}
                onChange={(e) => setSize(Number(e.target.value))}
                style={{ width: "100%" }}
              />
            </label>
            <label className="fld">
              <span>Error correction</span>
              <select value={ecc} onChange={(e) => setEcc(e.target.value)}>
                {["L", "M", "Q", "H"].map((x) => (
                  <option key={x} value={x}>{x}</option>
                ))}
              </select>
            </label>
            <label className="fld">
              <span>Foreground</span>
              <div className="row">
                <input type="color" value={fg} onChange={(e) => setFg(e.target.value)} />
                <input type="text" value={fg} onChange={(e) => setFg(e.target.value)} />
              </div>
            </label>
            <label className="fld">
              <span>Background</span>
              <div className="row">
                <input type="color" value={bg} onChange={(e) => setBg(e.target.value)} />
                <input type="text" value={bg} onChange={(e) => setBg(e.target.value)} />
              </div>
            </label>
            <div className="contrast-meter">
              <span style={{ color: "var(--dim)" }}>CONTRAST</span>
              <span className="contrast-ratio">{contrastRatio(fg, bg).toFixed(1)}:1</span>
              <span className={`contrast-grade ${contrastBadgeClass(fg, bg)}`}>
                {contrastGrade(fg, bg)}
              </span>
            </div>
            <label className="fld">
              <span>Module shape</span>
              <select value={shape} onChange={(e) => setShape(e.target.value)}>
                <option value="square">SQUARE</option>
                <option value="rounded">ROUNDED</option>
                <option value="dot">DOT</option>
              </select>
            </label>
            <label className="fld">
              <span>Format</span>
              <select value={format} onChange={(e) => setFormat(e.target.value)}>
                <option value="png">PNG</option>
                <option value="jpg">JPEG</option>
                <option value="svg">SVG</option>
                <option value="pdf">PDF</option>
              </select>
            </label>
            <label className="fld">
              <span>Quiet zone ({quiet})</span>
              <input
                type="range"
                min={0}
                max={8}
                value={quiet}
                onChange={(e) => setQuiet(Number(e.target.value))}
              />
            </label>
            <label className="fld">
              <span>Center logo (optional)</span>
              <input
                type="file"
                accept="image/*"
                onChange={async (e) => {
                  const f = e.target.files?.[0];
                  setLogoB64(f ? await fileToDataURL(f) : "");
                }}
                style={{ color: "var(--dim)" }}
              />
            </label>
          </div>
          {logoB64 && (
            <button className="btn" onClick={() => setLogoB64("")}>
              REMOVE LOGO
            </button>
          )}
        </div>

        <div className="panel">
          <span className="panel-title">STYLE PRESETS</span>
          <div className="row" style={{ marginBottom: 8 }}>
            <input
              type="text"
              value={presetName}
              placeholder="NEW PRESET NAME…"
              onChange={(e) => setPresetName(e.target.value)}
              style={{ flex: 1 }}
            />
            <button
              className="btn primary"
              disabled={!presetName.trim()}
              onClick={async () => {
                try {
                  const updated = await QRService.UpsertPreset({
                    name: presetName.trim(),
                    style: {
                      format,
                      ecc,
                      fg_color: fg,
                      bg_color: bg,
                      shape,
                      size,
                      margin: quiet,
                    },
                  });
                  setPresets(updated || []);
                  setSelectedPreset(presetName.trim());
                  onToast(`preset "${presetName.trim()}" saved`, false);
                } catch (err) {
                  onToast(String(err), true);
                }
              }}
            >
              SAVE
            </button>
          </div>
          <div className="row">
            <select
              value={selectedPreset}
              onChange={(e) => setSelectedPreset(e.target.value)}
              style={{ flex: 1 }}
            >
              <option value="">LOAD PRESET…</option>
              {presets.map((p) => (
                <option key={p.name} value={p.name}>
                  {p.name}
                </option>
              ))}
            </select>
            <button
              className="btn"
              disabled={!selectedPreset}
              onClick={() => {
                const p = presets.find((x) => x.name === selectedPreset);
                if (!p?.style) return;
                setSize(p.style.size || 512);
                setEcc(p.style.ecc || "M");
                setFg(p.style.fg_color || "#000000");
                setBg(p.style.bg_color || "#FFFFFF");
                setShape(p.style.shape || "square");
                setFormat(p.style.format || "png");
                if (p.style.format === "svg" || p.style.format === "pdf") {
                  setQuiet(4);
                } else {
                  setQuiet(p.style.margin ?? 4);
                }
                onToast(`preset "${p.name}" applied`, false);
              }}
            >
              APPLY
            </button>
            <button
              className="btn"
              disabled={!selectedPreset}
              onClick={async () => {
                try {
                  const updated = await QRService.DeletePreset(selectedPreset);
                  setPresets(updated || []);
                  setSelectedPreset("");
                  onToast("preset deleted", false);
                } catch (err) {
                  onToast(String(err), true);
                }
              }}
            >
              DEL
            </button>
          </div>
        </div>

        <button
          className="btn primary"
          style={{ width: "100%", padding: 13 }}
          disabled={busy}
          onClick={generate}
        >
          {busy ? "RENDERING…" : "▶ GENERATE"}
        </button>
        <button
          className="btn"
          style={{ width: "100%", marginTop: 6, padding: 10 }}
          onClick={() => setShowBatch(true)}
        >
          BATCH GENERATE
        </button>
        {error && <div className="error-text">✖ {error}</div>}
      </div>

      <div className="panel">
        <span className="panel-title">PREVIEW</span>
        {preview ? (
          <>
            <div className="qr-preview">
              <img src={preview} alt="generated QR code" />
            </div>
            <div className="row" style={{ marginTop: 12 }}>
              <button
                className="btn primary"
                onClick={async () => {
                  try {
                    const path = await QRService.SaveDataURL(preview, "qrcode.png");
                    onToast(`saved → ${path}`, false);
                  } catch (err) {
                    if (!(err instanceof Error && err.message.includes("cancel")))
                      onToast(String(err), true);
                  }
                }}
              >
                SAVE AS…
              </button>
              <button
                className="btn"
                onClick={async () => {
                  try {
                    await fetch(preview)
                      .then((r) => r.blob())
                      .then((b) =>
                        navigator.clipboard.write([
                          new ClipboardItem({ "image/png": b }),
                        ]),
                      );
                    onToast("image copied to clipboard", false);
                  } catch {
                    onToast("clipboard copy failed (use PNG format)", true);
                  }
                }}
              >
                COPY IMAGE
              </button>
            </div>
            <div className="warnlist">
              {warnings.map((w, i) => (
                <div className="warn" key={i}>⚠ {w}</div>
              ))}
            </div>
          </>
        ) : (
          <div className="hint" style={{ padding: 40, textAlign: "center" }}>
            NO CODE RENDERED_<br />
            <span style={{ fontSize: 10 }}>fill content → hit GENERATE</span>
          </div>
        )}
      </div>
      {showBatch && (
        <div className="modal-overlay" onClick={() => !batchBusy && setShowBatch(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <span className="panel-title">BATCH GENERATION</span>
              <button className="btn" onClick={() => !batchBusy && setShowBatch(false)}>✕</button>
            </div>
            <p className="hint" style={{ marginBottom: 12 }}>
              Upload a CSV or JSON file with QR entries. The generated ZIP will download automatically.
            </p>
            <label className="btn" style={{ display: "block", textAlign: "center", padding: 24, cursor: "pointer", borderStyle: "dashed" }}>
              {batchBusy ? "PROCESSING…" : "CLICK TO SELECT CSV / JSON"}
              <input
                ref={batchRef}
                type="file"
                accept=".csv,.json"
                style={{ display: "none" }}
                onChange={handleBatch}
                disabled={batchBusy}
              />
            </label>
          </div>
        </div>
      )}
    </div>
  );
}

function renderFields(
  type: QRType,
  fields: Record<string, string>,
  set: (k: string, v: string) => void,
) {
  const F = (
    k: string,
    label: string,
    ph = "",
    password = false,
  ) => (
    <label className="fld" key={k}>
      <span>{label}</span>
      <input
        type={password ? "password" : "text"}
        value={fields[k] || ""}
        placeholder={ph}
        onChange={(e) => set(k, e.target.value)}
      />
    </label>
  );

  switch (type) {
    case "url":
      return F("url", "URL", "https://example.com");
    case "wifi":
      return (
        <>
          {F("ssid", "SSID (network name)")}
          {F("password", "Password", "", true)}
          <label className="fld">
            <span>Encryption</span>
            <select
              value={fields.encryption || "WPA"}
              onChange={(e) => set("encryption", e.target.value)}
            >
              <option value="WPA">WPA/WPA2/WPA3</option>
              <option value="WEP">WEP</option>
              <option value="nopass">OPEN</option>
            </select>
          </label>
          <label className="row hint" style={{ marginBottom: 10 }}>
            <input
              type="checkbox"
              checked={fields.hidden === "true"}
              onChange={(e) => set("hidden", String(e.target.checked))}
            />
            HIDDEN NETWORK
          </label>
        </>
      );
    case "email":
      return (
        <>
          {F("to", "Recipient", "name@example.com")}
          {F("subject", "Subject")}
          {F("body", "Body")}
        </>
      );
    case "sms":
      return (
        <>
          {F("phone", "Phone number", "+15551234567")}
          {F("message", "Message")}
        </>
      );
    case "phone":
      return F("phone", "Phone number", "+15551234567");
    case "vcard":
      return (
        <>
          {F("first_name", "First name")}
          {F("last_name", "Last name")}
          {F("org", "Organization")}
          {F("title", "Title")}
          {F("phone", "Phone")}
          {F("email", "Email")}
          {F("url", "Website")}
          {F("address", "Address")}
          {F("note", "Note")}
        </>
      );
    case "geo":
      return (
        <>
          {F("latitude", "Latitude", "52.5200")}
          {F("longitude", "Longitude", "13.4050")}
        </>
      );
    case "event":
      return (
        <>
          {F("title", "Title")}
          {F("start", "Start (2026-08-24 09:30)", "2026-08-24 09:30")}
          {F("end", "End (optional)", "2026-08-24 11:00")}
          {F("location", "Location")}
          {F("description", "Description")}
        </>
      );
    case "social":
      return (
        <>
          <label className="fld">
            <span>Platform</span>
            <select
              className="social-select"
              value={fields.platform || "x"}
              onChange={(e) => set("platform", e.target.value)}
            >
              <option value="x">X (Twitter)</option>
              <option value="instagram">Instagram</option>
              <option value="github">GitHub</option>
              <option value="linkedin">LinkedIn</option>
              <option value="youtube">YouTube</option>
              <option value="tiktok">TikTok</option>
              <option value="telegram">Telegram</option>
              <option value="whatsapp">WhatsApp</option>
            </select>
          </label>
          <label className="fld">
            <span>{fields.platform === "whatsapp" ? "Phone number" : "Handle / username"}</span>
            <input
              type="text"
              value={fields.handle || ""}
              placeholder={fields.platform === "whatsapp" ? "+15551234567" : "yourhandle"}
              onChange={(e) => set("handle", e.target.value)}
            />
          </label>
        </>
      );
    default:
      return (
        <label className="fld">
          <span>Text content</span>
          <textarea
            rows={5}
            value={fields.text || ""}
            onChange={(e) => set("text", e.target.value)}
          />
        </label>
      );
  }
}
