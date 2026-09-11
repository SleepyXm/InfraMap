"use client";

import type { HTMLAttributes, PointerEvent } from "react";
import { cx } from "@/app/UI/classnames";
import styles from "@/app/UI/InteractiveGrid.module.css";

export function InteractiveGrid({ className, onPointerMove, ...props }: HTMLAttributes<HTMLDivElement>) {
  const handlePointerMove = (event: PointerEvent<HTMLDivElement>) => {
    const bounds = event.currentTarget.getBoundingClientRect();
    event.currentTarget.style.setProperty("--grid-x", `${event.clientX - bounds.left}px`);
    event.currentTarget.style.setProperty("--grid-y", `${event.clientY - bounds.top}px`);
    onPointerMove?.(event);
  };

  return <div {...props} className={cx(styles.grid, className)} onPointerMove={handlePointerMove} />;
}
