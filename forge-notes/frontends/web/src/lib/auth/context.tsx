"use client";

// yours: scaffolded once, never touched again — forge will not overwrite this file
// The app's auth state, read from this app's API (GET /auth/session; the cookie is HttpOnly, so no
// token is reachable here). The Identity decides what to SHOW, never what is ALLOWED. MOCK_API set:
// the identity is the fixture user and no request is sent; in hybrid it is UI-only, never a
// credential. Depth: `forge skill load auth/frontend`.
// useAuth() lives in ./use-auth (so Fast Refresh can hot-swap this provider): import it from there.

import { identifyUser } from "@reliantlabs/forge-web-runtime";
import { useCallback, useContext, useEffect, useRef, useState, type ReactNode } from "react";

import { EventBusContext } from "../use-event";

import { fetchIdentity, signOut as apiSignOut, type Identity } from "./native-login";
import { createSessionAuthProvider, isMockMode } from "./session-provider";
import { AuthContext, type AuthContextValue } from "./use-auth";
// The runtime config's IDENTIFY_USER (config.js, window.__FORGE_CONFIG__):
// anything but "false" lets the signed-in user be named to telemetry.
function identifyUserAllowed(): boolean {
  if (typeof window === "undefined") return false;
  const cfg = (window as unknown as { __FORGE_CONFIG__?: Record<string, unknown> })
    .__FORGE_CONFIG__;
  return cfg?.IDENTIFY_USER !== "false";
}
export function AuthContextProvider({ children }: { children: ReactNode }) {
  const [identity, setIdentity] = useState<Identity>({ authenticated: false });
  const [isLoading, setIsLoading] = useState(true);
  // The fixture session, in mock mode only — see "Mock mode is the one
  // exception" above. Decided once, at mount: the mode is runtime config,
  // and it does not change under a running page. Held in state for the same
  // reason the React Native tree holds its provider there — a logout must
  // move the one session the callbacks below read.
  const [mockSession] = useState(() => (isMockMode() ? createSessionAuthProvider() : null));
  const refresh = useCallback(async () => {
    if (mockSession) {
      const user = mockSession.getUser();
      setIdentity({
        authenticated: user !== null,
        email: user?.email,
        name: user?.name,
        subject: user?.id,
      });
    } else {
      setIdentity(await fetchIdentity());
    }
    setIsLoading(false);
  }, [mockSession]);
  // ONE request on mount, to this app's own origin — none in mock mode.
  // Compare with what this replaced: a discovery fetch, an authorize
  // redirect, and a hidden iframe that sat on a five-second timeout whenever
  // there was no session to restore.
  //
  // Awaited inside an async IIFE rather than called bare. `refresh` sets
  // state, and an effect body that updates state synchronously forces a
  // second render pass before the browser paints. Awaiting first moves both
  // setState calls into a continuation — the effect body itself only starts
  // the work, which is what an effect is for.
  useEffect(() => {
    void (async () => {
      await refresh();
    })();
  }, [refresh]);
  // Tell telemetry who is signed in: identifyUser attaches the opaque
  // subject to every span and replay (the Sessions view, `forge obs sessions
  // --user`) and sends one daily `auth_signed_in {kind: "resumed"}`. The
  // server already counted the sign-in itself; this joins sessions to users.
  // Not in mock mode, whose user is a fixture, and not when the runtime
  // config says IDENTIFY_USER = "false" (telemetry.identify_user = False),
  // which turns off this path as it does otel_gen.ts's session watcher: then
  // any id attached earlier is cleared and no auth event is sent.
  useEffect(() => {
    if (mockSession || isLoading) return;
    if (!identifyUserAllowed()) {
      identifyUser(null);
      return;
    }
    identifyUser(identity.authenticated ? (identity.subject ?? null) : null);
  }, [mockSession, isLoading, identity.authenticated, identity.subject]);
  // A session can lapse mid-use: the cookie expires, or the server revokes it.
  // The transport emits `auth:expired` when an RPC fails Unauthenticated (HTTP
  // 401; see src/lib/connect.ts). Re-ask the server who is signed in rather
  // than trusting the one failure: when the session really is gone, identity
  // flips to signed-out and RouteGuard sends the user to sign-in with a
  // returnTo; when it is not (one rpc's own auth rule), nothing changes. A
  // burst of failing queries shares one check.
  const bus = useContext(EventBusContext);
  const expiryCheck = useRef<Promise<void> | null>(null);
  useEffect(() => {
    if (!bus) return;
    return bus.on("auth:expired", () => {
      expiryCheck.current ??= refresh().finally(() => {
        expiryCheck.current = null;
      });
    });
  }, [bus, refresh]);

  const logout = useCallback(async () => {
    if (mockSession) {
      // Ends the FIXTURE session, locally, so a signed-out state can be
      // built against. There is no server session to end in pure mock, and
      // in hybrid the fixture was never what the server saw. A reload
      // restores it.
      await mockSession.logout();
      setIdentity({ authenticated: false });
      return;
    }
    setIdentity(await apiSignOut());
  }, [mockSession]);

  const value: AuthContextValue = {
    identity,
    isAuthenticated: identity.authenticated,
    isLoading,
    refresh,
    logout,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
