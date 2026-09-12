import styles from "@/app/UI/Workspace.module.css";

type ResourceKey = string | number;

export type ResourceOption<Key extends ResourceKey> = {
  id: Key;
  title: string;
  description: string;
  tag: string;
};

type ResourceSelectorProps<Key extends ResourceKey> = {
  eyebrow: string;
  title: string;
  description: string;
  permissionNote: string;
  searchPlaceholder: string;
  emptyMessage: string;
  options: ResourceOption<Key>[];
  selected: Set<Key>;
  search: string;
  loading: boolean;
  saving: boolean;
  error: string;
  onSearchChange: (value: string) => void;
  onToggle: (id: Key) => void;
  onClose: () => void;
  onSave: () => void;
};

export function ResourceSelector<Key extends ResourceKey>({
  eyebrow,
  title,
  description,
  permissionNote,
  searchPlaceholder,
  emptyMessage,
  options,
  selected,
  search,
  loading,
  saving,
  error,
  onSearchChange,
  onToggle,
  onClose,
  onSave,
}: ResourceSelectorProps<Key>) {
  return (
    <section className={styles.repositoryPanel}>
      <div className={styles.repositoryHeader}>
        <div>
          <p className={styles.eyebrow}>{eyebrow}</p>
          <h2>{title}</h2>
          <p>{description}</p>
        </div>
        <button type="button" className={styles.closeButton} aria-label="Close selector" onClick={onClose}>×</button>
      </div>

      <p className={styles.permissionNote}>{permissionNote}</p>

      <div className={styles.repositoryToolbar}>
        <input
          type="search"
          value={search}
          onChange={(event) => onSearchChange(event.target.value)}
          placeholder={searchPlaceholder}
          aria-label={searchPlaceholder}
        />
        <span>{selected.size} selected</span>
      </div>

      {loading ? (
        <p className={styles.repositoryState}>Loading available resources…</p>
      ) : error ? (
        <p className={styles.error}>{error}</p>
      ) : options.length === 0 ? (
        <p className={styles.repositoryState}>{emptyMessage}</p>
      ) : (
        <div className={styles.repositoryList}>
          {options.map((option) => (
            <label key={option.id} className={styles.repositoryRow}>
              <input type="checkbox" checked={selected.has(option.id)} onChange={() => onToggle(option.id)} />
              <span>
                <strong>{option.title}</strong>
                <small>{option.description}</small>
              </span>
              <em>{option.tag}</em>
            </label>
          ))}
        </div>
      )}

      <div className={styles.repositoryFooter}>
        <button type="button" className={styles.cancelButton} onClick={onClose}>Cancel</button>
        <button type="button" className={styles.primaryAction} disabled={loading || saving} onClick={onSave}>
          {saving ? "Saving…" : "Save selection"}
        </button>
      </div>
    </section>
  );
}
