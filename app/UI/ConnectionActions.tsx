import styles from "./ConnectionActions.module.css";

type ConnectionActionsProps = {
  primaryLabel: string;
  reconnectLabel: string;
  onPrimary: () => void;
  onReconnect: () => void;
};

export function ConnectionActions({ primaryLabel, reconnectLabel, onPrimary, onReconnect }: ConnectionActionsProps) {
  return (
    <div className={styles.actions}>
      <button type="button" className={styles.primary} onClick={onPrimary}>
        <span>{primaryLabel}</span><span>→</span>
      </button>
      <button type="button" className={styles.secondary} onClick={onReconnect}>{reconnectLabel}</button>
    </div>
  );
}
