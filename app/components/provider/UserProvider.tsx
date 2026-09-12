"use client";
import { createContext, useContext, useState, useEffect, useRef, ReactNode } from "react";
import { User } from "@/app/components/types/users";
import { validateUser } from "@/app/components/handlers/auth";

export interface UserContextType {
  user: User | null;
  setUser: React.Dispatch<React.SetStateAction<User | null>>;
  resolved: boolean;
}

export const UserContext = createContext<UserContextType | undefined>(undefined);

function getCachedUser(): User | null {
  try {
    const cached = localStorage.getItem("user");
    return cached ? JSON.parse(cached) : null;
  } catch {
    return null;
  }
}

export const UserProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<User | null>(null);
  const [resolved, setResolved] = useState(false);
  const didRun = useRef(false);

  const initAuth = async () => {
    try {
      const result = await validateUser();
      setUser(result?.user ?? null);
      if (!result) {
        localStorage.removeItem("user");
      }
    } catch {
      setUser(null);
      localStorage.removeItem("user");
    } finally {
      setResolved(true);
    }
  };

  useEffect(() => {
    if (didRun.current) return;
    didRun.current = true;

    const cachedUser = getCachedUser();

    if (cachedUser) {
      setUser(cachedUser);
      setResolved(true);
    }

    initAuth();

    const handlePageShow = (e: PageTransitionEvent) => {
      if (e.persisted) initAuth();
    };

    window.addEventListener("pageshow", handlePageShow);
    return () => window.removeEventListener("pageshow", handlePageShow);
  }, []);

  return (
    <UserContext.Provider value={{ user, setUser, resolved }}>
      {children}
    </UserContext.Provider>
  );
};

export const useUser = () => {
  const ctx = useContext(UserContext);
  if (!ctx) throw new Error("useUser must be inside UserProvider");
  return ctx;
};
