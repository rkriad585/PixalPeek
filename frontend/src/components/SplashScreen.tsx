import { LogoIcon } from "./Icons";

export default function SplashScreen() {
  return (
    <div className="splash">
      <div className="splash-content">
        <LogoIcon size={64} className="splash-logo" />
        <div className="splash-name">PIXALPEEK</div>
        <div className="splash-bar">
          <div className="splash-bar-fill" />
        </div>
      </div>
    </div>
  );
}
