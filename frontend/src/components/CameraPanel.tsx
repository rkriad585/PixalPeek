import { useCallback, useEffect, useRef, useState } from "react";
import { QRService, type DecodeResult } from "../lib/api";
import { vibrate } from "../lib/formats";

type LogEntry = { content: string; type: string; time: string };

type Props = { onToast: (msg: string, err: boolean) => void };

export default function CameraPanel({ onToast }: Props) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const overlayRef = useRef<HTMLCanvasElement>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const timerRef = useRef<number | undefined>(undefined);
  const busyRef = useRef(false);
  const seenRef = useRef<Set<string>>(new Set());

  const [devices, setDevices] = useState<MediaDeviceInfo[]>([]);
  const [deviceId, setDeviceId] = useState("");
  const [active, setActive] = useState(false);
  const [status, setStatus] = useState("");
  const [log, setLog] = useState<LogEntry[]>([]);

  const refreshDevices = useCallback(async () => {
    try {
      const devs = await navigator.mediaDevices.enumerateDevices();
      const cams = devs.filter((d) => d.kind === "videoinput");
      setDevices(cams);
      if (!deviceId && cams.length > 0) setDeviceId(cams[0].deviceId);
    } catch {
      /* enumeration unavailable */
    }
  }, [deviceId]);

  useEffect(() => {
    refreshDevices();
  }, [refreshDevices]);

  const drawBoxes = useCallback((results: DecodeResult[]) => {
    const video = videoRef.current;
    const canvas = overlayRef.current;
    if (!video || !canvas || !video.videoWidth) return;
    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    ctx.lineWidth = Math.max(3, Math.round(video.videoWidth / 320));
    ctx.strokeStyle = getComputedStyle(document.documentElement).getPropertyValue("--accent").trim() || "#00ff99";
    ctx.font = "12px monospace";
    for (const r of results) {
      if (!r.bounding_box) continue;
      const tl = r.bounding_box.top_left;
      const tr = r.bounding_box.top_right;
      const bl = r.bounding_box.bottom_left;
      if (!tl || !tr || !bl) continue;
      const br = r.bounding_box.bottom_right ?? [
        tr[0] + bl[0] - tl[0],
        tr[1] + bl[1] - tl[1],
      ];
      ctx.beginPath();
      ctx.moveTo(tl[0], tl[1]);
      ctx.lineTo(tr[0], tr[1]);
      ctx.lineTo(br[0], br[1]);
      ctx.lineTo(bl[0], bl[1]);
      ctx.closePath();
      ctx.stroke();
    }
  }, []);

  const tick = useCallback(async () => {
    if (busyRef.current) return;
    const video = videoRef.current;
    if (!video || !video.videoWidth) return;
    busyRef.current = true;
    try {
      const canvas = document.createElement("canvas");
      canvas.width = video.videoWidth;
      canvas.height = video.videoHeight;
      const ctx = canvas.getContext("2d", { willReadFrequently: true });
      if (!ctx) return;
      ctx.drawImage(video, 0, 0);
      const dataURL = canvas.toDataURL("image/png");
      const res = await QRService.DecodeBase64(
        dataURL.slice(dataURL.indexOf(",") + 1),
        false,
      );
      if (res.success && res.results.length > 0) {
        drawBoxes(res.results);
        const fresh: LogEntry[] = [];
        for (const r of res.results) {
          if (seenRef.current.has(r.content)) continue;
          seenRef.current.add(r.content);
          fresh.push({
            content: r.content,
            type: String(r.content_type || "text"),
            time: new Date().toLocaleTimeString(),
          });
        }
        if (fresh.length > 0) {
          vibrate(50);
          setLog((prev) => [...fresh, ...prev].slice(0, 100));
          onToast(`decoded ${fresh.length} code(s)`, false);
        }
      }
    } catch {
      /* keep scanning */
    } finally {
      busyRef.current = false;
    }
  }, [drawBoxes, onToast]);

  const stopCamera = useCallback(() => {
    if (timerRef.current !== undefined) {
      window.clearInterval(timerRef.current);
      timerRef.current = undefined;
    }
    streamRef.current?.getTracks().forEach((t) => t.stop());
    streamRef.current = null;
    setActive(false);
    setStatus("CAMERA STANDBY");
  }, []);

  const startCamera = useCallback(async () => {
    stopCamera();
    setStatus("REQUESTING PERMISSION…");
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: deviceId
          ? { deviceId: { exact: deviceId }, width: { ideal: 1280 } }
          : { facingMode: "environment", width: { ideal: 1280 } },
        audio: false,
      });
      streamRef.current = stream;
      const video = videoRef.current;
      if (video) {
        video.srcObject = stream;
        video.setAttribute("playsinline", "true");
        await video.play();
      }
      setActive(true);
      setStatus("");
      refreshDevices();
      seenRef.current.clear();
      timerRef.current = window.setInterval(() => void tick(), 550);
    } catch (err) {
      setStatus(
        "CAMERA UNAVAILABLE — " +
          (err instanceof Error ? err.message : String(err)),
      );
    }
  }, [deviceId, refreshDevices, stopCamera, tick]);

  useEffect(() => () => stopCamera(), [stopCamera]);

  return (
    <div style={{ marginTop: 16 }}>
      <div className="section-label" style={{ margin: "0 0 8px" }}>
        LIVE SENSOR
      </div>
      <div className="row" style={{ marginBottom: 10 }}>
        <select
          value={deviceId}
          onChange={(e) => setDeviceId(e.target.value)}
          disabled={active}
          style={{ flex: 1 }}
        >
          {devices.length === 0 && <option value="">NO CAMERAS FOUND</option>}
          {devices.map((d, i) => (
            <option key={d.deviceId} value={d.deviceId}>
              {d.label || `CAMERA ${i + 1}`}
            </option>
          ))}
        </select>
        <button className={"btn" + (active ? "" : " primary")} onClick={active ? stopCamera : startCamera}>
          {active ? "STOP SENSOR" : "ACTIVATE"}
        </button>
      </div>

      <div
        className="camera-box"
        style={{
          position: "relative",
          display: active ? "block" : "none",
          border: "1px solid var(--accent)",
        }}
      >
        <video ref={videoRef} autoPlay muted playsInline style={{ width: "100%", display: "block" }} />
        <canvas
          ref={overlayRef}
          style={{
            position: "absolute",
            inset: 0,
            width: "100%",
            height: "100%",
            pointerEvents: "none",
          }}
        />
        <div
          style={{
            position: "absolute",
            top: 6,
            left: 6,
            width: 14,
            height: 14,
            borderTop: "2px solid rgba(255,255,255,0.65)",
            borderLeft: "2px solid rgba(255,255,255,0.65)",
            pointerEvents: "none",
          }}
        />
        <div
          style={{
            position: "absolute",
            top: 6,
            right: 6,
            width: 14,
            height: 14,
            borderTop: "2px solid rgba(255,255,255,0.65)",
            borderRight: "2px solid rgba(255,255,255,0.65)",
            pointerEvents: "none",
          }}
        />
        <div
          style={{
            position: "absolute",
            bottom: 6,
            left: 6,
            width: 14,
            height: 14,
            borderBottom: "2px solid rgba(255,255,255,0.65)",
            borderLeft: "2px solid rgba(255,255,255,0.65)",
            pointerEvents: "none",
          }}
        />
        <div
          style={{
            position: "absolute",
            bottom: 6,
            right: 6,
            width: 14,
            height: 14,
            borderBottom: "2px solid rgba(255,255,255,0.65)",
            borderRight: "2px solid rgba(255,255,255,0.65)",
            pointerEvents: "none",
          }}
        />
      </div>

      {!active && (
        <div className="hint" style={{ padding: "4px 0 8px" }}>
          {status || "WEBCAM DECODE — FRAMES ANALYSED LOCALLY VIA GOZXING"}
        </div>
      )}
      {active && status && <div className="hint" style={{ paddingBottom: 8 }}>{status}</div>}

      {log.length > 0 && (
        <>
          <div
            className="row"
            style={{ justifyContent: "space-between", margin: "10px 0 6px" }}
          >
            <span className="section-label" style={{ margin: 0 }}>
              LOGGED DECODES [{log.length}]
            </span>
            <button
              className="icon-btn"
              onClick={() => {
                setLog([]);
                seenRef.current.clear();
              }}
            >
              CLEAR
            </button>
          </div>
          <div style={{ maxHeight: 220, overflowY: "auto" }}>
            {log.map((entry, i) => (
              <div key={`${entry.time}-${i}`} className="result-card" style={{ marginTop: 6 }}>
                <div className="result-head">
                  <span className="badge">{entry.type}</span>
                  <span className="badge meta">{entry.time}</span>
                </div>
                <div className="payload-box">{entry.content}</div>
                <div className="row" style={{ marginTop: 6 }}>
                  <button
                    className="btn"
                    onClick={() => {
                      navigator.clipboard.writeText(entry.content);
                      onToast("payload copied", false);
                    }}
                  >
                    COPY
                  </button>
                  {entry.type === "url" && (
                    <a href={entry.content} target="_blank" rel="noreferrer" style={{ textDecoration: "none" }}>
                      <button className="btn">OPEN URL ↗</button>
                    </a>
                  )}
                </div>
              </div>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
