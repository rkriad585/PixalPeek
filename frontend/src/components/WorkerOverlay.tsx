import { useEffect, useRef, useState } from "react";
import { runCamera, runClipCopy, runClipDecode } from "../lib/worker";

const LABELS: Record<string, string> = {
  clipdecode: "SCANNING CLIPBOARD IMAGE",
  camera: "CAMERA SCAN ACTIVE",
  clipcopy: "COPYING TO CLIPBOARD",
};

export default function WorkerOverlay({ task }: { task: string }) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const [status, setStatus] = useState("");

  useEffect(() => {
    if (task === "clipdecode") {
      runClipDecode();
    } else if (task === "clipcopy") {
      runClipCopy();
    } else if (task === "camera" && videoRef.current) {
      let cleanup = () => {};
      runCamera(videoRef.current, setStatus).then((f) => {
        cleanup = f;
      });
      return () => cleanup();
    }
  }, [task]);

  return (
    <div className="overlay">
      <h1>▚▞ PIXALPEEK</h1>
      {task === "camera" ? (
        <div className="camera-box">
          <video ref={videoRef} autoPlay muted playsInline />
        </div>
      ) : (
        <div className="spinner" />
      )}
      <div className="status-line">{LABELS[task]}</div>
      {status && <div className="error-text">{status}</div>}
      <div className="hint">
        point at a QR code · window closes automatically on detection
      </div>
    </div>
  );
}
