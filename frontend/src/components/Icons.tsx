interface IconProps {
  size?: number;
  className?: string;
}

export const ScanIcon = ({ size = 16, className }: IconProps) => (
  <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <path d="M1 5V3a2 2 0 0 1 2-2h2" />
    <path d="M11 1h2a2 2 0 0 1 2 2v2" />
    <path d="M15 11v2a2 2 0 0 1-2 2h-2" />
    <path d="M5 15H3a2 2 0 0 1-2-2v-2" />
    <line x1="8" y1="4" x2="8" y2="12" />
    <line x1="4" y1="8" x2="12" y2="8" />
  </svg>
);

export const GenerateIcon = ({ size = 16, className }: IconProps) => (
  <svg width={size} height={size} viewBox="0 0 16 16" fill="currentColor" className={className}>
    <rect x="1" y="1" width="4" height="4" />
    <rect x="11" y="1" width="4" height="4" />
    <rect x="1" y="11" width="4" height="4" />
    <rect x="6" y="6" width="4" height="4" />
  </svg>
);

export const HistoryIcon = ({ size = 16, className }: IconProps) => (
  <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <path d="M2.5 2v4h4" />
    <path d="M2.5 6a6 6 0 1 1 1 4" />
    <path d="M8 8l2.5-2.5" />
    <path d="M8 8l-1 2.5" />
  </svg>
);

export const SettingsIcon = ({ size = 16, className }: IconProps) => (
  <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <circle cx="8" cy="8" r="2" />
    <path d="M13.5 8a5.5 5.5 0 0 0-.1-1l1.1-1.3-1.5-2.6-1.3.6a5.5 5.5 0 0 0-1.7-1L10 1.5H7l-.5 1.2a5.5 5.5 0 0 0-1.7 1l-1.3-.6-1.5 2.6 1.1 1.3a5.5 5.5 0 0 0-.1 1 5.5 5.5 0 0 0 .1 1l-1.1 1.3 1.5 2.6 1.3-.6a5.5 5.5 0 0 0 1.7 1L7 14.5h3l.5-1.2a5.5 5.5 0 0 0 1.7-1l1.3.6 1.5-2.6-1.1-1.3a5.5 5.5 0 0 0 .1-1z" />
  </svg>
);

export const SplitIcon = ({ size = 16, className }: IconProps) => (
  <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <rect x="1.5" y="1.5" width="13" height="13" rx="1" />
    <line x1="8" y1="1.5" x2="8" y2="14.5" />
  </svg>
);

export const CloseIcon = ({ size = 16, className }: IconProps) => (
  <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <line x1="4" y1="4" x2="12" y2="12" />
    <line x1="12" y1="4" x2="4" y2="12" />
  </svg>
);

export const MinimizeIcon = ({ size = 16, className }: IconProps) => (
  <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <line x1="3" y1="8" x2="13" y2="8" />
  </svg>
);

export const MaximizeIcon = ({ size = 16, className }: IconProps) => (
  <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <rect x="2.5" y="2.5" width="11" height="11" rx="1" />
  </svg>
);

export const CopyIcon = ({ size = 16, className }: IconProps) => (
  <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <rect x="5" y="5" width="9" height="9" rx="1" />
    <path d="M5 11H3a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1h8a1 1 0 0 1 1 1v2" />
  </svg>
);

export const DownloadIcon = ({ size = 16, className }: IconProps) => (
  <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <path d="M8 2v9" />
    <path d="M5 8l3 3 3-3" />
    <path d="M2 12v2a1 1 0 0 0 1 1h10a1 1 0 0 0 1-1v-2" />
  </svg>
);

export const OpenUrlIcon = ({ size = 16, className }: IconProps) => (
  <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <path d="M10 1.5H4a1.5 1.5 0 0 0-1.5 1.5v10A1.5 1.5 0 0 0 4 14.5h8a1.5 1.5 0 0 0 1.5-1.5V7" />
    <path d="M8.5 8.5l4-4" />
    <path d="M9 1.5h4v4" />
  </svg>
);

export const CameraIcon = ({ size = 16, className }: IconProps) => (
  <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <path d="M2 5h12a1 1 0 0 1 1 1v7a1 1 0 0 1-1 1H2a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1z" />
    <path d="M5 5V3.5a.5.5 0 0 1 .5-.5h5a.5.5 0 0 1 .5.5V5" />
    <circle cx="8" cy="8.5" r="2" />
  </svg>
);

export const PasteIcon = ({ size = 16, className }: IconProps) => (
  <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <rect x="3" y="3" width="10" height="12" rx="1" />
    <path d="M6 1.5h4a1 1 0 0 1 1 1V4H5V2.5a1 1 0 0 1 1-1z" />
    <line x1="5.5" y1="7.5" x2="10.5" y2="7.5" />
    <line x1="5.5" y1="10.5" x2="9.5" y2="10.5" />
  </svg>
);

export const StarIcon = ({ size = 16, className }: IconProps) => (
  <svg width={size} height={size} viewBox="0 0 16 16" fill="currentColor" className={className}>
    <path d="M8 1.5l1.85 3.75L14 5.9l-3 2.92.71 4.13L8 10.9l-3.71 1.95.71-4.13-3-2.92 4.15-.65z" />
  </svg>
);

export const DeleteIcon = ({ size = 16, className }: IconProps) => (
  <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <path d="M2 4h12" />
    <path d="M5.5 4V2.5a1 1 0 0 1 1-1h3a1 1 0 0 1 1 1V4" />
    <path d="M3 4l.7 9.1a1 1 0 0 0 1 .9h6.6a1 1 0 0 0 1-.9L13 4" />
    <line x1="6.5" y1="7" x2="6.5" y2="11.5" />
    <line x1="9.5" y1="7" x2="9.5" y2="11.5" />
  </svg>
);

export const RefreshIcon = ({ size = 16, className }: IconProps) => (
  <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <path d="M1 7.5a6.5 6.5 0 0 1 11.8-3.5" />
    <path d="M15 8.5a6.5 6.5 0 0 1-11.8 3.5" />
    <path d="M12.8 1v3.5H9.3" />
    <path d="M3.2 15V11.5h3.5" />
  </svg>
);

export const LogoIcon = ({ size = 16, className }: IconProps) => (
  <svg width={size} height={size} viewBox="0 0 16 16" fill="currentColor" className={className}>
    <rect x="1" y="1" width="6" height="6" />
    <rect x="9" y="1" width="6" height="6" />
    <rect x="1" y="9" width="6" height="6" />
    <rect x="3" y="3" width="2" height="2" fill="var(--bg, #0a0a0b)" />
    <rect x="11" y="3" width="2" height="2" fill="var(--bg, #0a0a0b)" />
    <rect x="3" y="11" width="2" height="2" fill="var(--bg, #0a0a0b)" />
  </svg>
);
