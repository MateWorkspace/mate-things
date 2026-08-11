/**
 * The full record-new-device journey, from the pre-session form through
 * the live wizard. Shared by both `record/new` (steps 0-1) and
 * `record/[id]` (all 8) so the stepper shows the whole journey from the
 * very first screen, not just "step N of 2" before a session even exists.
 */
export const RECORD_WIZARD_STEPS = [
  "Device details",
  "State values",
  "Building cases",
  "Review & set default",
  "Recording",
  "Analyzing",
  "Testing",
  "Finishing",
] as const;
