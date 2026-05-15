"use client";

import { useRef } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { register, login } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";

export default function AuthPage() {
  const router = useRouter();
  const { refreshAuth } = useAuth();

  const registerEmailRef = useRef<HTMLInputElement>(null);
  const registerPasswordRef = useRef<HTMLInputElement>(null);
  const loginEmailRef = useRef<HTMLInputElement>(null);
  const loginPasswordRef = useRef<HTMLInputElement>(null);

  const handleRegister = async () => {
    const email = registerEmailRef.current?.value ?? "";
    const password = registerPasswordRef.current?.value ?? "";
    try {
      await register(email, password);
      await refreshAuth();
      router.push("/");
    } catch (err) {
      alert(
        "Registration failed: " +
          (err instanceof Error ? err.message : "Unknown error"),
      );
    }
  };

  const handleLogin = async () => {
    const email = loginEmailRef.current?.value ?? "";
    const password = loginPasswordRef.current?.value ?? "";
    try {
      await login(email, password);
      await refreshAuth();
      router.push("/");
    } catch (err) {
      alert(
        "Login failed: " +
          (err instanceof Error ? err.message : "Unknown error"),
      );
    }
  };

  const handleKeyDown =
    (action: () => void) => (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === "Enter") action();
    };

  return (
    <div className="min-h-screen bg-app-bg text-primary font-mono text-sm overflow-auto">
      <div className="grain" />

      {/* Header */}
      <header className="flex items-center justify-between px-10 py-5 border-b border-border">
        <Link
          href="/"
          className="font-display text-[1.3rem] text-accent tracking-[-0.02em] no-underline"
        >
          ⌁ AI Notes
        </Link>
      </header>

      {/* Auth panels */}
      <div className="grid grid-cols-2 gap-5 max-w-[800px] mx-auto px-10 py-10 max-[640px]:grid-cols-1">
        {/* Register */}
        <section className="bg-panel border border-border rounded-[10px] p-8">
          <div className="text-[0.7rem] tracking-[0.2em] uppercase text-accent mb-2">
            01 — New here?
          </div>
          <h2 className="font-display text-[1.8rem] font-normal text-primary mb-6 tracking-[-0.02em]">
            Create an account.
          </h2>
          <div className="flex flex-col gap-3">
            <input
              ref={registerEmailRef}
              type="email"
              placeholder="Email address"
              autoComplete="email"
              onKeyDown={handleKeyDown(handleRegister)}
              className="font-mono text-[0.8rem] bg-app-bg border border-border-light rounded-[6px] text-primary px-4 py-2.5 w-full outline-none transition-colors focus:border-accent placeholder:text-muted"
            />
            <input
              ref={registerPasswordRef}
              type="password"
              placeholder="Password"
              autoComplete="new-password"
              onKeyDown={handleKeyDown(handleRegister)}
              className="font-mono text-[0.8rem] bg-app-bg border border-border-light rounded-[6px] text-primary px-4 py-2.5 w-full outline-none transition-colors focus:border-accent placeholder:text-muted"
            />
            <button
              onClick={handleRegister}
              className="font-mono text-[0.75rem] tracking-[0.03em] px-[18px] py-[9px] rounded-[6px] border border-accent bg-accent text-primary font-medium cursor-pointer transition-all hover:bg-accent-dim hover:border-accent-dim w-full text-center"
            >
              Create Account
            </button>
          </div>
        </section>

        {/* Login */}
        <section className="bg-panel border border-border rounded-[10px] p-8">
          <div className="text-[0.7rem] tracking-[0.2em] uppercase text-accent mb-2">
            02 — Welcome back.
          </div>
          <h2 className="font-display text-[1.8rem] font-normal text-primary mb-6 tracking-[-0.02em]">
            Sign in.
          </h2>
          <div className="flex flex-col gap-3">
            <input
              ref={loginEmailRef}
              type="email"
              placeholder="Email address"
              autoComplete="email"
              onKeyDown={handleKeyDown(handleLogin)}
              className="font-mono text-[0.8rem] bg-app-bg border border-border-light rounded-[6px] text-primary px-4 py-2.5 w-full outline-none transition-colors focus:border-accent placeholder:text-muted"
            />
            <input
              ref={loginPasswordRef}
              type="password"
              placeholder="Password"
              autoComplete="current-password"
              onKeyDown={handleKeyDown(handleLogin)}
              className="font-mono text-[0.8rem] bg-app-bg border border-border-light rounded-[6px] text-primary px-4 py-2.5 w-full outline-none transition-colors focus:border-accent placeholder:text-muted"
            />
            <button
              onClick={handleLogin}
              className="font-mono text-[0.75rem] tracking-[0.03em] px-[18px] py-[9px] rounded-[6px] border border-accent bg-accent text-primary font-medium cursor-pointer transition-all hover:bg-accent-dim hover:border-accent-dim w-full text-center"
            >
              Sign In
            </button>
          </div>
        </section>
      </div>
    </div>
  );
}
