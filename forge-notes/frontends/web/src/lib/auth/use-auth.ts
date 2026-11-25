// yours: scaffolded once, never touched again — forge will not overwrite this file
//
// useAuth() — read the app's auth state anywhere under <AuthContextProvider>
// (./context.tsx, which owns how the identity is fetched).
//
// The context object and its hook live here, apart from the provider
// component, so ./context.tsx exports nothing but a component. React Fast
// Refresh can only hot-swap a module whose exports are all components; a
// hook exported beside the provider turns every edit to the provider into a
// full page reload and loses the state the reload was meant to keep
// (eslint-plugin-react-refresh's only-export-components).
import { createContext, useContext } from "react";

import type { Identity } from "./native-login";

export interface AuthContextValue {
  /** Who is signed in. Presentation only — see the note in ./context.tsx. */
  identity: Identity;
  isAuthenticated: boolean;
  /** True until the first identity check completes. */
  isLoading: boolean;
  /** Re-read the session from the server. Call after signing in. */
  refresh: () => Promise<void>;
  logout: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextValue | null>(null);

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within an AuthContextProvider");
  }
  return ctx;
}
