"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useUser } from "@/app/components/provider/UserProvider";
import { logout } from "@/app/components/handlers/auth";
import { Surface } from "@/app/UI/primitives";
import styles from "@/app/UI/Nav.module.css";
import { cx } from "@/app/UI/classnames";

export default function Navbar() {
  const { user, setUser, resolved } = useUser();
  const [mobileOpen, setMobileOpen] = useState(false);
  const router = useRouter();
  const pathname = usePathname();

  const handleLogout = async () => {
    await logout();
    setUser(null);
    setMobileOpen(false);
    router.push("/");
  };

  const links = [
    { label: "Home", url: "/" },
    { label: "Workspaces", url: "/workspaces" },

    // Prevent hydration mismatch by rendering the logged-out state
    // until the client-side auth check has finished.
    ...(!resolved
      ? [{ label: "Sign in", url: "/login" }]
      : user
        ? [
            { label: "Profile", url: "/Profile" },
            { label: "Sign out", onClick: handleLogout },
          ]
        : [{ label: "Sign in", url: "/login" }]),
  ];

  return (
    <Surface as="header" tone="light" opacity={0.92} borderOpacity={0.1} border="bottom" blur="lg" className={styles.header}>
      <div className={styles.inner}>
        {/* Logo */}
        <Link href="/" className={styles.brand}><span className={styles.brandMark} aria-hidden="true" /><span>InfraMap</span></Link>

        {/* Desktop Links */}
        <nav>
          <ul className={styles.links}>
            {links.map((link) => (
              <li key={link.label}>
                {link.url ? (
                  <Link
                    href={link.url}
                    className={cx(styles.link, (link.url === "/" ? pathname === "/" : pathname.startsWith(link.url)) && styles.linkActive)}
                  >
                    {link.label}
                  </Link>
                ) : (
                  <button
                    onClick={link.onClick}
                    className={styles.link}
                  >
                    {link.label}
                  </button>
                )}
              </li>
            ))}
          </ul>
        </nav>

        {/* Mobile Toggle */}
        <button
          className={styles.menuButton}
          onClick={() => setMobileOpen(!mobileOpen)}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            className={styles.menuIcon}
          >
            <path d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>

        {/* Mobile Menu */}
        {mobileOpen && (
          <Surface as="ul" tone="light" opacity={0.98} borderOpacity={0.1} blur="md" padding="1rem" width="100%" className={styles.mobileLinks}>
            {links.map((link) => (
              <li key={link.label}>
                {link.url ? (
                  <Link
                    href={link.url}
                    className={styles.mobileLink}
                  >
                    {link.label}
                  </Link>
                ) : (
                  <button
                    onClick={link.onClick}
                    className={styles.mobileLink}
                  >
                    {link.label}
                  </button>
                )}
              </li>
            ))}
          </Surface>
        )}
      </div>
    </Surface>
  );
}
