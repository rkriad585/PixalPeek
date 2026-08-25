function hexToRGB(hex: string): [number, number, number] {
  const h = hex.replace("#", "");
  return [
    parseInt(h.substring(0, 2), 16),
    parseInt(h.substring(2, 4), 16),
    parseInt(h.substring(4, 6), 16),
  ];
}

function luminance(r: number, g: number, b: number): number {
  const a = [r, g, b].map((v) => {
    v /= 255;
    return v <= 0.03928 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4);
  });
  return a[0] * 0.2126 + a[1] * 0.7152 + a[2] * 0.0722;
}

export function contrastRatio(fg: string, bg: string): number {
  const [r1, g1, b1] = hexToRGB(fg);
  const [r2, g2, b2] = hexToRGB(bg);
  const l1 = luminance(r1, g1, b1);
  const l2 = luminance(r2, g2, b2);
  const lighter = Math.max(l1, l2);
  const darker = Math.min(l1, l2);
  return (lighter + 0.05) / (darker + 0.05);
}

export type ContrastGrade = "AAA" | "AA" | "FAIL";

export function contrastGrade(fg: string, bg: string): ContrastGrade {
  const ratio = contrastRatio(fg, bg);
  if (ratio >= 7) return "AAA";
  if (ratio >= 4.5) return "AA";
  return "FAIL";
}

export function contrastBadgeClass(fg: string, bg: string): string {
  const grade = contrastGrade(fg, bg);
  if (grade === "AAA") return "grade-aaa";
  if (grade === "AA") return "grade-aa";
  return "grade-fail";
}
