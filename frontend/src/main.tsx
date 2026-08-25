import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import SplashScreen from "./components/SplashScreen";
import "./styles.css";

const params = new URLSearchParams(window.location.search);
const isSplash = params.get("splash") === "1";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    {isSplash ? <SplashScreen /> : <App />}
  </StrictMode>,
);
