import { useCallback, useEffect, useRef, useState } from "react";
import { QRService, type ScanResponse, type SafetyAssessment, type HistoryEntry } from "../lib/api";
import { readClipboardImageB64 } from "../lib/api";
import { parseWifi, vibrate } from "../lib/formats";
import CameraPanel from "./CameraPanel";
import { TechReadout } from "./TechReadout";

type Props = { onToast: (msg: string, err: boolean) => void };

function downloadBlob(content: string, filename: string, mime: string) {
  const blob = new Blob([content], { type: mime });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

export default function ScannerView({ onToast }: Props) {
  const [multi, setMulti] = useState(false);
  const [busy, setBusy] = useState(false);
  const [result, setResult] = useState<ScanResponse | null>(null);
  const [error, setError] = useState("");
  const [over, setOver] = useState(false);
  const [previewUrl, setPreviewUrl] = useState("");
  const [inputName, setInputName] = useState("");
  const fileRef = useRef<HTMLInputElement>(null);

  const decodeFile = useCallback(
    async (file: File) => {
      setBusy(true);
      setError("");
      setInputName(file.name || "clipboard");
      try {
        const path = (file as File & { path?: string }).path;
        let res: ScanResponse;
        if (path) {
          const reader = new FileReader();
          const dataURL = await new Promise<string>((resolve, reject) => {
            reader.onload = () => resolve(String(reader.result));
            reader.onerror = () => reject(new Error("failed to read file"));
            reader.readAsDataURL(file);
          });
          setPreviewUrl(dataURL);
          res = await QRService.DecodeFile(path, multi);
          if (!res.source_file) res = { ...res, source_file: file.name };
          else {
            const parts = res.source_file.replace(/\\/g, "/").split("/");
            res = { ...res, source_file: parts[parts.length - 1] || file.name };
          }
        } else {
          const dataURL = await new Promise<string>((resolve, reject) => {
            const r = new FileReader();
            r.onload = () => resolve(String(r.result));
            r.onerror = () => reject(new Error("failed to read file"));
            r.readAsDataURL(file);
          });
          setPreviewUrl(dataURL);
          res = await QRService.DecodeBase64(
            dataURL.slice(dataURL.indexOf(",") + 1),
            multi,
          );
          res = { ...res, source_file: file.name };
        }
        setResult(res);
        if (!res.success) setError(res.error || "no QR code detected");
        else {
          vibrate(50);
          onToast(`${res.results.length} code(s) decoded`, false);
          for (const r of res.results) {
            QRService.AddHistory({
              id: "",
              kind: "scan" as any,
              content: r.content,
              content_type: r.content_type,
              timestamp: "",
              pinned: false,
              source: path || file.name,
              error_level: r.error_correction_level || "",
            }).catch(() => {});
          }
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : String(err));
      } finally {
        setBusy(false);
      }
    },
    [multi, onToast],
  );

  const decodeClipboard = useCallback(async () => {
    setBusy(true);
    setError("");
    setInputName("clipboard");
    setPreviewUrl("");
    try {
      const b64 = await readClipboardImageB64();
      setPreviewUrl("data:image/png;base64," + b64);
      const decoded = await QRService.DecodeBase64(b64, multi);
      const res = { ...decoded, source_file: "clipboard" };
      setResult(res);
      if (!res.success) setError(res.error || "no QR found in clipboard image");
      else {
        vibrate(50);
        onToast("clipboard decoded", false);
          for (const r of res.results) {
            QRService.AddHistory({
              id: "",
              kind: "scan" as any,
              content: r.content,
              content_type: r.content_type,
              timestamp: "",
              pinned: false,
              source: "clipboard",
              error_level: r.error_correction_level || "",
            }).catch(() => {});
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  }, [multi, onToast]);

  useEffect(() => {
    const onPaste = (e: ClipboardEvent) => {
      const items = e.clipboardData?.items;
      if (!items) return;
      for (const it of Array.from(items)) {
        if (it.type.startsWith("image/")) {
          const f = it.getAsFile();
          if (f) void decodeFile(f);
          break;
        }
      }
    };
    window.addEventListener("paste", onPaste);
    return () => window.removeEventListener("paste", onPaste);
  }, [decodeFile]);

  return (
    <div className="grid-2">
      <div className="panel">
        <span className="panel-title">INPUT</span>
        <div
          className={"dropzone" + (over ? " over" : "")}
          onClick={() => fileRef.current?.click()}
          onDragOver={(e) => {
            e.preventDefault();
            setOver(true);
          }}
          onDragLeave={() => setOver(false)}
          onDrop={(e) => {
            e.preventDefault();
            setOver(false);
            const f = e.dataTransfer.files?.[0];
            if (f) decodeFile(f);
          }}
        >
          {previewUrl ? (
            <img src={previewUrl} alt="input preview" className="input-preview" />
          ) : (
            <>
              <div className="big">▚▞</div>
              <div>PASTE (CTRL+V) · DROP IMAGE · CLICK TO BROWSE</div>
              <div className="hint">PNG · JPG · WEBP · BMP · GIF</div>
            </>
          )}
        </div>
        <input
          ref={fileRef}
          type="file"
          accept="image/*"
          hidden
          onChange={(e) => {
            const f = e.target.files?.[0];
            if (f) decodeFile(f);
            e.target.value = "";
          }}
        />
        <div className="row" style={{ marginTop: 12 }}>
          <button className="btn primary" disabled={busy} onClick={decodeClipboard}>
            {busy ? "SCANNING…" : "PASTE FROM CLIPBOARD"}
          </button>
          <label className="row hint" style={{ gap: 6, cursor: "pointer" }}>
            <input
              type="checkbox"
              checked={multi}
              onChange={(e) => setMulti(e.target.checked)}
            />
            MULTI-SCAN
          </label>
        </div>
        <CameraPanel onToast={onToast} />
      </div>

      <div className="panel">
        <span className="panel-title">RESULTS{inputName ? ` :: ${inputName}` : ""}</span>
        {!result && !error && (
          <div className="hint" style={{ padding: 20, textAlign: "center" }}>
            AWAITING INPUT_
          </div>
        )}
        {error && <div className="error-text">✖ {error}</div>}
        {result?.success &&
          result.results.map((r, i) => (
            <>
            <div className="result-card" key={i}>
              <div className="result-head">
                <span className="badge">{r.content_type}</span>
                <span className="badge meta">{r.format}</span>
                {r.error_correction_level && (
                  <span className="badge meta">ECC {r.error_correction_level}</span>
                )}
                {r.bounding_box && (
                  <span className="badge meta">
                    [{Math.round(r.bounding_box.top_left[0])},
                    {Math.round(r.bounding_box.top_left[1])}]
                  </span>
                )}
              </div>
              <div className="payload-box">{r.content}</div>
              {r.security && r.security.is_suspicious && (
                <div className="security-warn">
                  <span className="security-icon">&#9888;</span>
                  <div className="security-details">
                    <strong>SECURITY WARNING</strong>
                    <span className="security-score">Risk Score: {r.security.score}/100{r.security.is_shortener ? " · URL SHORTENER" : ""}</span>
                    {r.security.reasons.map((reason, j) => (
                      <span key={j} className="security-reason">{reason}</span>
                    ))}
                  </div>
                </div>
              )}
              {String(r.content_type) === "wifi" && (() => {
                const wifi = parseWifi(r.content);
                return (
                  <div className="hint" style={{ marginTop: 6 }}>
                    SSID: {wifi.ssid || "—"} · PASS: {wifi.password || "—"} ·{" "}
                    {wifi.encryption}
                    {wifi.hidden ? " · HIDDEN" : ""}
                  </div>
                );
              })()}
              <div className="row" style={{ marginTop: 8 }}>
                <button
                  className="btn"
                  onClick={() => {
                    navigator.clipboard.writeText(r.content);
                    onToast("payload copied", false);
                  }}
                >
                  COPY PAYLOAD
                </button>
                {r.content_type === "url" && (
                  <button className="btn" onClick={async () => {
                    try {
                      await QRService.OpenURL(r.content);
                    } catch {
                      window.open(r.content, "_blank");
                    }
                  }}>
                    OPEN URL
                  </button>
                )}
                {r.content_type === "vcard" && (
                  <button
                    className="btn"
                    onClick={() => downloadBlob(r.content, "contact.vcf", "text/vcard")}
                  >
                    SAVE .VCF
                  </button>
                )}
                {r.content_type === "event" && (
                  <button
                    className="btn"
                    onClick={() => downloadBlob(r.content, "event.ics", "text/calendar")}
                  >
                    SAVE .ICS
                  </button>
                )}
                {r.content_type === "geo" && (
                  <a
                    href={`https://www.openstreetmap.org/search?query=${r.content.replace("geo:", "")}`}
                    target="_blank"
                    rel="noreferrer"
                    style={{ textDecoration: "none" }}
                  >
                    <button className="btn">VIEW MAP</button>
                  </a>
                )}
                {r.content_type === "wifi" && (
                  <button
                    className="btn"
                    onClick={() => {
                      navigator.clipboard.writeText(parseWifi(r.content).password || r.content);
                      onToast("wifi credentials copied", false);
                    }}
                  >
                    COPY PASS
                  </button>
                )}
                {(r.content_type === "email" || r.content_type === "sms" || r.content_type === "phone") && (
                  <a
                    href={r.content_type === "email" ? r.content : r.content_type === "phone" ? r.content : `sms:${r.content.replace("SMSTO:", "").replace(":", "?body=")}`}
                    style={{ textDecoration: "none" }}
                  >
                    <button className="btn">
                      {r.content_type === "email" ? "OPEN MAIL" : r.content_type === "phone" ? "DIAL" : "OPEN SMS"}
                    </button>
                  </a>
                )}
              </div>
            </div>
            <TechReadout result={r} index={i} />
            </>
          ))}
      </div>
    </div>
  );
}
