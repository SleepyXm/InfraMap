"use client";

import { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { completeOAuthCallback } from "@/app/components/handlers/auth";
import { useUser } from "@/app/components/provider/UserProvider";
import styles from "@/app/UI/Auth.module.css";

export default function OAuthCallbackPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { setUser } = useUser();
  const [message, setMessage] = useState("Finalising your secure sign-in…");

  useEffect(() => {
    const error = searchParams.get("error");
    const isConnection = searchParams.get("intent") === "connect";
    if (error) { router.replace(isConnection ? `/Profile?error=${encodeURIComponent(error)}` : `/login?error=${encodeURIComponent(error)}`); return; }
    completeOAuthCallback(isConnection).then(user => { setUser(user); router.replace(isConnection ? "/Profile" : "/"); }).catch(() => { setMessage("We could not complete that sign-in."); setTimeout(() => router.replace(isConnection ? "/Profile?error=oauth_failed" : "/login?error=oauth_failed"), 1200); });
  }, [router, searchParams, setUser]);

  return <main className={styles.callback}><p>{message}</p></main>;
}
