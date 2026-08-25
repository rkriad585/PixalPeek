export function parseWifi(raw: string): {
  ssid: string;
  password: string;
  encryption: string;
  hidden: boolean;
} {
  const unesc = (s: string) => s.replace(/\\(.)/g, "$1");
  const t = raw.match(/T:((?:\\.|[^;])*)/)?.[1] || "WPA";
  const s = raw.match(/S:((?:\\.|[^;])*)/)?.[1] || "";
  const p = raw.match(/P:((?:\\.|[^;])*)/)?.[1] || "";
  const h = raw.match(/H:true/i)?.[1];
  return {
    ssid: unesc(s),
    password: unesc(p),
    encryption: unesc(t).toUpperCase(),
    hidden: !!h,
  };
}

export function vibrate(ms = 50): void {
  try {
    navigator.vibrate?.(ms);
  } catch {
    /* unsupported */
  }
}
