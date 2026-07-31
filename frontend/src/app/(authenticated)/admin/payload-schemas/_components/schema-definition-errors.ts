const SUFFIXES: Array<{ suffix: string; message: string }> = [
  { suffix: " has an invalid or missing type", message: "Choose a valid type." },
  {
    suffix: " enum type requires at least one option",
    message: "Add at least one option.",
  },
];

const PREFIX = "definition";

export function parseDefinitionFieldError(
  message: string,
): { path: string[]; message: string } | null {
  for (const { suffix, message: friendly } of SUFFIXES) {
    if (!message.endsWith(suffix)) continue;
    const head = message.slice(0, -suffix.length);
    if (head === PREFIX) return { path: [], message: friendly };
    if (head.startsWith(`${PREFIX}.`)) {
      return { path: head.slice(PREFIX.length + 1).split("."), message: friendly };
    }
  }
  return null;
}
