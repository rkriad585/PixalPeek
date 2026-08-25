import { QRService } from "./api";

export type WorkerTask = "clipdecode" | "camera" | "clipcopy" | "";

export function taskFromURL(): WorkerTask {
  const p = new URLSearchParams(window.location.search).get("task") || "";
  if (p === "clipdecode" || p === "camera" || p === "clipcopy") return p;
  return "";
}

async function complete(result: string) {
  try {
    await QRService.CompleteTask(result);
  } catch {
    /* window is closing */
  }
}

function fail(err: unknown) {
  const msg = err instanceof Error ? err.message : String(err);
  return "ERR:" + msg;
}

export async function runClipDecode(): Promise<void> {
  try {
    const items = await navigator.clipboard.read();
    let b64 = "";
    for (const item of items) {
      const type = item.types.find((t) => t.startsWith("image/"));
      if (!type) continue;
      const blob = await item.getType(type);
      b64 = await new Promise<string>((resolve, reject) => {
        const r = new FileReader();
        r.onload = () => resolve(String(r.result));
        r.onerror = () => reject(new Error("failed to read clipboard image"));
        r.readAsDataURL(blob);
      });
      break;
    }
    if (!b64) {
      await complete("ERR:no image found on clipboard");
      return;
    }
    const res = await QRService.DecodeBase64(b64, false);
    await complete(JSON.stringify(res));
  } catch (err) {
    await complete(fail(err));
  }
}

export async function runClipCopy(): Promise<void> {
  try {
    const payload = await QRService.GetWorkerPayload();
    if (!payload.startsWith("data:image/png;base64,")) {
      await complete("ERR:invalid worker payload");
      return;
    }
    const blob = await (await fetch(payload)).blob();
    await navigator.clipboard.write([
      new ClipboardItem({ "image/png": blob }),
    ]);
    await complete("OK");
  } catch (err) {
    await complete(fail(err));
  }
}

export async function runCamera(
  video: HTMLVideoElement,
  onStatus: (s: string) => void,
): Promise<() => void> {
  let stopped = false;
  let stream: MediaStream | null = null;
  let timer: number | undefined;

  const finish = async (result: string) => {
    if (stopped) return;
    stopped = true;
    if (timer) window.clearInterval(timer);
    stream?.getTracks().forEach((t) => t.stop());
    await complete(result);
  };

  try {
    stream = await navigator.mediaDevices.getUserMedia({
      video: { facingMode: "environment", width: { ideal: 1280 } },
      audio: false,
    });
  } catch (err) {
    onStatus(String(err));
    await finish(fail(err));
    return () => {};
  }

  video.srcObject = stream;
  video.play();

  const canvas = document.createElement("canvas");
  const ctx = canvas.getContext("2d")!;

  timer = window.setInterval(async () => {
    if (stopped || !video.videoWidth) return;
    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;
    ctx.drawImage(video, 0, 0);
    const dataURL = canvas.toDataURL("image/png");
    try {
      const res = await QRService.DecodeBase64(
        dataURL.slice(dataURL.indexOf(",") + 1),
        false,
      );
      if (res.success && res.results.length > 0) {
        await finish(JSON.stringify(res));
      }
    } catch {
      /* keep scanning */
    }
  }, 450);

  window.setTimeout(() => {
    if (!stopped) {
      onStatus("camera scan timed out");
      finish("");
    }
  }, 110000);

  return () => {
    if (!stopped) finish("");
  };
}
