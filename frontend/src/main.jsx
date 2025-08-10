import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";

createRoot(document.getElementById("root")).render(
  <StrictMode>
    <div className="flex items-center h-screen justify-center text-primary ">
      <p>Hiiiiii</p>
    </div>
  </StrictMode>
);
