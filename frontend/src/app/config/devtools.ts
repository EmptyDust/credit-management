export function isApiTesterEnabled(env = import.meta.env) {
  return env.DEV || env.VITE_ENABLE_API_TESTER === "true";
}
