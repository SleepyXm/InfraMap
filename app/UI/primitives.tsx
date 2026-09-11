import type {
  ButtonHTMLAttributes,
  CSSProperties,
  HTMLAttributes,
  InputHTMLAttributes,
  ReactNode,
  SelectHTMLAttributes,
  TextareaHTMLAttributes,
} from "react";
import Link from "next/link";
import { cx } from "@/app/UI/classnames";
import styles from "@/app/UI/UI.module.css";

type UIStyle = CSSProperties & Record<`--ui-${string}`, string | number>;
type SurfaceTone = "dark" | "light" | "neutral" | "deep" | "accent";
type Blur = "none" | "sm" | "md" | "lg";

const toneRGB: Record<SurfaceTone, string> = {
  dark: "0 0 0",
  light: "255 255 255",
  neutral: "23 23 23",
  deep: "10 10 10",
  accent: "94 234 212",
};

const blurSize: Record<Exclude<Blur, "none">, string> = {
  sm: "4px",
  md: "8px",
  lg: "16px",
};

export type SurfaceOptions = {
  tone?: SurfaceTone;
  opacity?: number;
  borderOpacity?: number;
  border?: "all" | "bottom" | "right";
  blur?: Blur;
  shadow?: boolean;
  width?: CSSProperties["width"];
  height?: CSSProperties["height"];
  maxWidth?: CSSProperties["maxWidth"];
  radius?: CSSProperties["borderRadius"];
  padding?: CSSProperties["padding"];
};

export function surfaceProps(
  {
    tone = "dark",
    opacity = 0.35,
    borderOpacity,
    border = "all",
    blur = "none",
    shadow = false,
    width,
    height,
    maxWidth,
    radius,
    padding,
  }: SurfaceOptions = {},
  className?: string,
  style?: CSSProperties,
) {
  const surfaceStyle: UIStyle = {
    "--ui-surface-rgb": toneRGB[tone],
    "--ui-surface-opacity": opacity,
    width,
    height,
    maxWidth,
    borderRadius: radius,
    padding,
    ...(borderOpacity === undefined
      ? {}
      : { "--ui-surface-border-opacity": borderOpacity }),
    ...(blur === "none" ? {} : { "--ui-surface-blur": blurSize[blur] }),
    ...style,
  };

  return {
    className: cx(
      styles.surface,
      borderOpacity !== undefined && border === "all" && styles.surfaceBorder,
      borderOpacity !== undefined && border === "bottom" && styles.surfaceBorderBottom,
      borderOpacity !== undefined && border === "right" && styles.surfaceBorderRight,
      blur !== "none" && styles.surfaceBlur,
      shadow && styles.surfaceShadow,
      className,
    ),
    style: surfaceStyle,
  };
}

type SurfaceProps = HTMLAttributes<HTMLElement> &
  SurfaceOptions & {
    as?: "div" | "section" | "aside" | "article" | "header" | "main" | "form" | "ul" | "p";
  };

export function Surface({
  as: Component = "div",
  tone,
  opacity,
  borderOpacity,
  border,
  blur,
  shadow,
  width,
  height,
  maxWidth,
  radius,
  padding,
  className,
  style,
  ...props
}: SurfaceProps) {
  const surface = surfaceProps(
    { tone, opacity, borderOpacity, border, blur, shadow, width, height, maxWidth, radius, padding },
    className,
    style,
  );
  return <Component {...props} {...surface} />;
}

type ButtonVariant = "action" | "primary" | "ghost" | "danger";
type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant;
  fullWidth?: boolean;
  width?: CSSProperties["width"];
  height?: CSSProperties["height"];
  padding?: CSSProperties["padding"];
  hover?: boolean;
};

export function Button({
  variant = "action",
  fullWidth = false,
  width,
  height,
  padding,
  hover = true,
  className,
  style,
  type = "button",
  ...props
}: ButtonProps) {
  const buttonStyle: UIStyle = {
    width: fullWidth ? "100%" : width,
    height,
    "--ui-button-padding": padding ?? "0",
    ...style,
  };
  return (
    <button
      {...props}
      type={type}
      className={cx(styles.button, styles[variant], !hover && styles.noHover, className)}
      style={buttonStyle}
    />
  );
}

type LinkButtonVariant = "brand" | "outline" | "dark";
type LinkButtonProps = {
  href: string;
  children: ReactNode;
  variant?: LinkButtonVariant;
  className?: string;
};

export function LinkButton({ href, children, variant = "brand", className }: LinkButtonProps) {
  return <Link href={href} className={cx(styles.linkButton, styles[`link${variant[0].toUpperCase()}${variant.slice(1)}`], className)}>{children}</Link>;
}

type BadgeTone = "live" | "watching" | "ready";
export function Badge({ children, tone = "ready", className }: { children: ReactNode; tone?: BadgeTone; className?: string }) {
  return <span className={cx(styles.badge, styles[`badge${tone[0].toUpperCase()}${tone.slice(1)}`], className)}>{children}</span>;
}

type ControlOptions = {
  tone?: SurfaceTone | "transparent";
  opacity?: number;
  borderOpacity?: number;
  blur?: Blur;
  focus?: "none" | "ring" | "border";
  focusWidth?: 1 | 2;
  focusOpacity?: number;
  width?: CSSProperties["width"];
  height?: CSSProperties["height"];
  radius?: CSSProperties["borderRadius"];
  padding?: CSSProperties["padding"];
  fullWidth?: boolean;
};

function controlProps(
  {
    tone = "light",
    opacity = 0.05,
    borderOpacity,
    blur = "none",
    focus = "none",
    focusWidth = 2,
    focusOpacity = 0.5,
    width,
    height,
    radius = "0.5rem",
    padding = "0.75rem",
    fullWidth = true,
  }: ControlOptions,
  className?: string,
  style?: CSSProperties,
) {
  const controlStyle: UIStyle = {
    "--ui-control-rgb": tone === "transparent" ? "0 0 0" : toneRGB[tone],
    "--ui-control-opacity": tone === "transparent" ? 0 : opacity,
    width,
    height,
    "--ui-control-radius": radius,
    "--ui-control-padding": padding,
    "--ui-control-focus-width": `${focusWidth}px`,
    "--ui-control-focus-opacity": focusOpacity,
    ...(borderOpacity === undefined
      ? {}
      : { "--ui-control-border-opacity": borderOpacity }),
    ...(blur === "none" ? {} : { "--ui-control-blur": blurSize[blur] }),
    ...style,
  };
  return {
    className: cx(
      styles.control,
      fullWidth && styles.controlFullWidth,
      borderOpacity !== undefined && styles.controlBorder,
      blur !== "none" && styles.controlBlur,
      focus === "ring" && styles.focusRing,
      focus === "border" && styles.focusBorder,
      className,
    ),
    style: controlStyle,
  };
}

type InputProps = InputHTMLAttributes<HTMLInputElement> & ControlOptions;
export function Input({ tone, opacity, borderOpacity, blur, focus, focusWidth, focusOpacity, width, height, radius, padding, fullWidth, className, style, ...props }: InputProps) {
  return <input {...props} {...controlProps({ tone, opacity, borderOpacity, blur, focus, focusWidth, focusOpacity, width, height, radius, padding, fullWidth }, className, style)} />;
}

type SelectProps = SelectHTMLAttributes<HTMLSelectElement> & ControlOptions;
export function Select({ tone, opacity, borderOpacity, blur, focus, focusWidth, focusOpacity, width, height, radius, padding, fullWidth, className, style, ...props }: SelectProps) {
  return <select {...props} {...controlProps({ tone, opacity, borderOpacity, blur, focus, focusWidth, focusOpacity, width, height, radius, padding, fullWidth }, className, style)} />;
}

type TextareaProps = TextareaHTMLAttributes<HTMLTextAreaElement> & ControlOptions;
export function Textarea({ tone, opacity, borderOpacity, blur, focus, focusWidth, focusOpacity, width, height, radius, padding, fullWidth, className, style, ...props }: TextareaProps) {
  return <textarea {...props} {...controlProps({ tone, opacity, borderOpacity, blur, focus, focusWidth, focusOpacity, width, height, radius, padding, fullWidth }, cx(styles.textarea, className), style)} />;
}

export function PropertyRow({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div {...props} className={cx(styles.propertyRow, className)} />;
}

export function MessageBubble({ fromUser = false, className, ...props }: HTMLAttributes<HTMLDivElement> & { fromUser?: boolean }) {
  return <div {...props} className={cx(styles.messageBubble, fromUser ? styles.userMessage : styles.assistantMessage, className)} />;
}
