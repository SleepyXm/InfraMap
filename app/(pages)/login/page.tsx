"use client";

import { useState } from "react";
import {
  beginOAuth,
  login,
  OAuthProvider,
  signup,
} from "@/app/components/handlers/auth";
import { useUser } from "@/app/components/provider/UserProvider";
import styles from "@/app/UI/Auth.module.css";

const providers: { id: OAuthProvider; label: string }[] = [
  { id: "github", label: "Continue with GitHub" },
  { id: "google", label: "Continue with Google" },
  { id: "aws", label: "Continue with AWS" },
];

export default function AuthPage() {
  const { setUser } = useUser();
  const [signupMode, setSignupMode] = useState(false);
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [message, setMessage] = useState("");
  const [pending, setPending] = useState(false);
  const submit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setMessage("");
    setPending(true);
    try {
      if (signupMode) {
        if (password !== confirmPassword)
          throw new Error("Passwords do not match.");
        const result = await signup(
          username.trim(),
          email.trim().toLowerCase(),
          password,
        );
        setMessage(result.message ?? "Account created. You can sign in now.");
        setSignupMode(false);
        return;
      }
      const result = await login(email.trim().toLowerCase(), password);
      if (!result) throw new Error("Sign-in failed. Please try again.");
      setUser(result.user);
      window.location.assign("/");
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "Something went wrong. Please try again.",
      );
    } finally {
      setPending(false);
    }
  };
  return (
    <main className={styles.page}>
      <div className={styles.shell}>
        <section className={styles.intro}>
          <p className={styles.eyebrow}>InfraMap / access</p>
          <h1>Operate with context, not complexity.</h1>
          <p>
            Bring your infrastructure into focus with one operational view, then
            take the next action from the same place.
          </p>
          <div className={styles.note}>
            <span />
            Secure sessions are managed by the control plane.
          </div>
        </section>
        <section className={styles.panel}>
          <div className={styles.panelHead}>
            <p className={styles.eyebrow}>
              {signupMode ? "Create account" : "Welcome back"}
            </p>
            <h2>
              {signupMode ? "Start with a clear view." : "Sign in to InfraMap."}
            </h2>
            <p>
              {signupMode
                ? "Create a workspace for your infrastructure."
                : "Use your account or a connected provider."}
            </p>
          </div>
          <form onSubmit={submit} className={styles.form}>
            {signupMode && (
              <label>
                Username
                <input
                  required
                  value={username}
                  onChange={(event) => setUsername(event.target.value)}
                  placeholder="Your name"
                />
              </label>
            )}
            <label>
              Email
              <input
                required
                type="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                placeholder="you@company.com"
              />
            </label>
            <label>
              Password
              <input
                required
                type="password"
                minLength={8}
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                placeholder="At least 8 characters"
              />
            </label>
            {signupMode && (
              <label>
                Confirm password
                <input
                  required
                  type="password"
                  minLength={8}
                  value={confirmPassword}
                  onChange={(event) => setConfirmPassword(event.target.value)}
                  placeholder="Repeat your password"
                />
              </label>
            )}
            <button
              className={styles.primaryButton}
              disabled={pending}
              type="submit"
            >
              {pending ? "Working…" : signupMode ? "Create account" : "Sign in"}
              <span aria-hidden="true">→</span>
            </button>
          </form>
          {message && <p className={styles.message}>{message}</p>}
          <div className={styles.divider}>
            <span>or continue with</span>
          </div>
          <div className={styles.providers}>
            {providers.map((provider) => (
              <button
                key={provider.id}
                type="button"
                className={styles.providerButton}
                onClick={() => beginOAuth(provider.id)}
              >
                {provider.label}
                <span aria-hidden="true">↗</span>
              </button>
            ))}
          </div>
          <p className={styles.switch}>
            {signupMode ? "Already have an account?" : "New to InfraMap?"}{" "}
            <button
              type="button"
              onClick={() => {
                setSignupMode(!signupMode);
                setMessage("");
              }}
            >
              {signupMode ? "Sign in" : "Create one"}
            </button>
          </p>
        </section>
      </div>
    </main>
  );
}
