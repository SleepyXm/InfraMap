import type { ReactNode } from "react";
import { Badge } from "@/app/UI/primitives";
import styles from "@/app/UI/Landing.module.css";

type FeatureState = "Live" | "Watching" | "Ready";
const stateTone = { Live: "live", Watching: "watching", Ready: "ready" } as const;

export function FeatureCard({ id, title, detail, state, featured = false }: { id: string; title: string; detail: string; state: FeatureState; featured?: boolean }) {
  return <article className={`${styles.featureCard} ${featured ? styles.featureCardFeatured : ""}`}><div className={styles.featureMeta}><span>{id}</span><Badge tone={stateTone[state]}>{state}</Badge></div><h3>{title}</h3><p>{detail}</p></article>;
}

export function SignalPreview({ children }: { children?: ReactNode }) {
  return <div className={styles.signalPreview}><div className={styles.signalGrid} /><div className={styles.signalGlow} /><div className={styles.signalPathOne} /><div className={styles.signalPathTwo} /><div className={`${styles.signalNode} ${styles.signalNodeOne}`}><span /></div><div className={`${styles.signalNode} ${styles.signalNodeTwo}`}><span /></div><div className={`${styles.signalNode} ${styles.signalNodeThree}`}><span /></div><div className={styles.signalCaption}>{children}</div></div>;
}
