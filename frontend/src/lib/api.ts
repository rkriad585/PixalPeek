import * as QRService from "../../bindings/github.com/rkriad585/PixalPeek/internal/service/qrservice";
import type {
  HistoryEntry,
  Preset,
  Settings,
} from "../../bindings/github.com/rkriad585/PixalPeek/internal/service/models";
import type {
  BoundingBox,
  BuildContentRequest,
  ContentType,
  DecodeResult,
  EncodeOptions,
  ScanResponse,
  SafetyAssessment,
  BatchScanResult,
} from "../../bindings/github.com/rkriad585/PixalPeek/internal/qrengine/models";

export { QRService };
export type {
  HistoryEntry,
  Preset,
  Settings,
  BoundingBox,
  BuildContentRequest,
  ContentType,
  DecodeResult,
  EncodeOptions,
  ScanResponse,
  SafetyAssessment,
  BatchScanResult,
};

export const TYPES = [
  "text",
  "url",
  "wifi",
  "email",
  "sms",
  "phone",
  "vcard",
  "geo",
  "event",
  "social",
] as const;

export type QRType = (typeof TYPES)[number];

export async function readClipboardImageB64(): Promise<string> {
  const items = await navigator.clipboard.read();
  for (const item of items) {
    const type = item.types.find((t) => t.startsWith("image/"));
    if (!type) continue;
    const blob = await item.getType(type);
    return blobToB64(blob);
  }
  throw new Error("no image found on clipboard");
}

export function blobToB64(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      const s = String(reader.result);
      resolve(s.slice(s.indexOf(",") + 1));
    };
    reader.onerror = () => reject(new Error("failed to read image data"));
    reader.readAsDataURL(blob);
  });
}

export function fileToDataURL(file: File | Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result));
    reader.onerror = () => reject(new Error("failed to read file"));
    reader.readAsDataURL(file);
  });
}

export function copyImageToSystemClipboard(dataURL: string): Promise<void> {
  return fetch(dataURL)
    .then((r) => r.blob())
    .then((blob) =>
      navigator.clipboard.write([new ClipboardItem({ "image/png": blob })]),
    );
}
