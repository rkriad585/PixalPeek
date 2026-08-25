import { useEffect, useRef, useState } from "react";

interface MenuItem {
  label: string;
  shortcut?: string;
  action: () => void;
  divider?: boolean;
  disabled?: boolean;
}

interface Props {
  items: MenuItem[];
}

export default function ContextMenu({ items }: Props) {
  const [visible, setVisible] = useState(false);
  const [pos, setPos] = useState({ x: 0, y: 0 });
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      e.preventDefault();
      const x = e.clientX;
      const y = e.clientY;
      setPos({ x, y });
      setVisible(true);
    };

    const close = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setVisible(false);
      }
    };

    window.addEventListener("contextmenu", handler);
    window.addEventListener("mousedown", close);
    return () => {
      window.removeEventListener("contextmenu", handler);
      window.removeEventListener("mousedown", close);
    };
  }, []);

  useEffect(() => {
    if (!visible) return;
    const close = () => setVisible(false);
    window.addEventListener("keydown", close);
    return () => window.removeEventListener("keydown", close);
  }, [visible]);

  if (!visible) return null;

  const vw = window.innerWidth;
  const vh = window.innerHeight;
  const menuW = 220;
  const menuH = items.length * 32;
  const x = pos.x + menuW > vw ? vw - menuW - 8 : pos.x;
  const y = pos.y + menuH > vh ? vh - menuH - 8 : pos.y;

  return (
    <div
      ref={ref}
      className="ctx-menu"
      style={{ left: x, top: y }}
    >
      {items.map((item, i) =>
        item.divider ? (
          <div key={i} className="ctx-divider" />
        ) : (
          <button
            key={i}
            className={"ctx-item" + (item.disabled ? " disabled" : "")}
            disabled={item.disabled}
            onClick={() => {
              item.action();
              setVisible(false);
            }}
          >
            <span>{item.label}</span>
            {item.shortcut && <span className="ctx-shortcut">{item.shortcut}</span>}
          </button>
        )
      )}
    </div>
  );
}
