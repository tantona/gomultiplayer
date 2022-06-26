import { useRef } from "react";

import { useEffect } from "react";

import { PixiApp } from "./PixiApp";

const app = new PixiApp();

export const Pixi = () => {
  const pixiContainerRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (pixiContainerRef?.current !== null) {
      for (const child of pixiContainerRef?.current.children) {
        pixiContainerRef?.current.removeChild(child);
      }
      pixiContainerRef?.current.appendChild(app.view);
      app.initialize();
    }
  }, [pixiContainerRef?.current]);

  return <div ref={pixiContainerRef} />;
};
