export const API_BASE_URL: string = envString(
  "API_BASE_URL",
  "http://127.0.0.1:80",
);
export const IS_PRODUCTION: boolean = process.env.NODE_ENV === "production";
export const COOKIE_SECURE: boolean = process.env.COOKIE_SECURE === "true";

function envString(key: string, fallback: string): string {
  const value = process.env[key];
  return value && value.trim() !== "" ? value : fallback;
}
