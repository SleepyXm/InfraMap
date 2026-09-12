"use client";

import { useState } from "react";
import {
  beginOAuth,
  logout,
} from "@/app/components/handlers/auth";
import { useUser } from "@/app/components/provider/UserProvider";
import type { OAuthIdentityProvider, User } from "@/app/components/types/users";
import styles from "@/app/UI/Profile.module.css";

const tabs = ["Account", "Connections", "Security"] as const;
const providers: { id: OAuthIdentityProvider; label: string; detail: string }[] = [
  {
    id: "github",
    label: "GitHub",
    detail: "Repository and organisation context",
  },
  {
    id: "google",
    label: "Google",
    detail: "Google Cloud identity and project context",
  },
];

export default function ProfilePage() {
  const { user, resolved, setUser } = useUser();
  const [tab, setTab] = useState<(typeof tabs)[number]>("Account");
  if (!resolved)
    return (
      <main className={styles.state}>
        Loading your workspace…
      </main>
    );
  if (!user)
    return (
      <main className={styles.state}>
        Sign in to manage your account.
      </main>
    );
  return (
    <main className={styles.page}>
      <div className={styles.shell}>
        <header className={styles.header}>
          <div>
            <p className={styles.eyebrow}>InfraMap / profile</p>
            <h1>Profile settings</h1>
          </div>
          <div className={styles.user}>
            <span className={styles.avatar}>
              {user.username.slice(0, 2).toUpperCase()}
            </span>
            <div>
              <strong>{user.username}</strong>
              <small>{user.email}</small>
            </div>
          </div>
        </header>
        <div className={styles.layout}>
          <aside className={styles.sidebar}>
            <nav>
              {tabs.map((item) => (
                <button
                  type="button"
                  key={item}
                  className={tab === item ? styles.activeTab : ""}
                  onClick={() => setTab(item)}
                >
                  {item}
                </button>
              ))}
            </nav>
            <button
              type="button"
              className={styles.logout}
              onClick={() => {
                void logout();
                setUser(null);
                window.location.assign("/");
              }}
            >
              Sign out
            </button>
          </aside>
          <section className={styles.content}>
            {tab === "Account" && <Account user={user} />}
            {tab === "Connections" && <Connections user={user} />}
            {tab === "Security" && <Security user={user} />}
          </section>
        </div>
      </div>
    </main>
  );
}
function Account({ user }: { user: User }) {
  const accounts = user.accounts ?? [];
  return (
    <>
      <div className={styles.sectionHead}>
        <p className={styles.eyebrow}>Account</p>
        <h2>Your operator identity</h2>
        <p>
          Keep the account that owns your infrastructure clear and easy to
          verify.
        </p>
      </div>
      <div className={styles.rows}>
        <Row label="Username" value={user.username} />
        <Row label="Email" value={user.email} />
        <Row
          label="Verification"
          value={user.verified ? "Verified" : "Pending verification"}
        />
        <Row
          label="Workspace access"
          value={`${accounts.length} workspace${accounts.length === 1 ? "" : "s"}`}
        />
      </div>
    </>
  );
}
function Connections({ user }: { user: User }) {
  const identities = user.identities ?? [];
  return (
    <>
      <div className={styles.sectionHead}>
        <p className={styles.eyebrow}>Connections</p>
        <h2>Bring the right context in.</h2>
        <p>
          Connect a provider identity now. Infrastructure access remains scoped
          and managed by the backend.
        </p>
      </div>
      <div className={styles.connectionGrid}>
        {providers.map((provider) => {
          const identity = identities.find(
            (item) => item.provider === provider.id,
          );
          return (
            <article key={provider.id} className={styles.connection}>
              <div>
                <span className={styles.providerMark}>
                  {provider.label.slice(0, 1)}
                </span>
                <h3>{provider.label}</h3>
                <p>{identity ? identity.email : provider.detail}</p>
              </div>
              <button type="button" onClick={() => beginOAuth(provider.id, "connect")}>
                {identity ? "Reconnect" : "Connect"}
                <span aria-hidden="true">→</span>
              </button>
            </article>
          );
        })}
      </div>
    </>
  );
}
function Security({ user }: { user: User }) {
  const identities = user.identities ?? [];
  return (
    <>
      <div className={styles.sectionHead}>
        <p className={styles.eyebrow}>Security</p>
        <h2>Connected identities</h2>
        <p>Review the providers allowed to authenticate as you.</p>
      </div>
      <div className={styles.rows}>
        {identities.length ? (
          identities.map((identity) => (
            <Row
              key={identity.id}
              label={identity.provider}
              value={identity.email}
            />
          ))
        ) : (
          <div className={styles.empty}>
            No connected provider identities yet.
          </div>
        )}
      </div>
    </>
  );
}
function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className={styles.row}>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}
