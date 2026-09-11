import { cx } from "@/app/UI/classnames";
import styles from "@/app/UI/UI.module.css";

type JitterLoaderProps = {
  message: string;
  className?: string;
};

export default function JitterLoader({ message, className = "" }: JitterLoaderProps) {
  return (
    <div
      role="status"
      aria-live="polite"
      aria-busy="true"
      className={cx(styles.loader, className)}
    >
      <span aria-hidden="true" className={styles.loaderDots}>
        {[0, 1, 2].map((index) => (
          <span
            key={index}
            className={styles.loaderDot}
            style={{ animationDelay: `${index * 100}ms` }}
          />
        ))}
      </span>
      <span className={styles.loaderMessage}>{message}</span>
    </div>
  );
}
