# BLE Direct page design

## Context

The ESP32 firmware (`mate-espidf-base`) exposes a local BLE GATT config
surface — settings, WiFi manager, log, and system info — intended for
provisioning and debugging a device without going through the backend or
even without the device having WiFi connectivity yet. There is currently no
UI for it; verifying it means using a generic BLE inspector app.

This adds a **BLE Direct** page to the `mate-things` frontend, under the
Operations nav group, that talks to the device directly from the browser
via the Web Bluetooth API — bypassing the `mate-things` backend entirely.
It also requires one small, coordinated firmware change (below) to remove a
redundant control path that would otherwise leak into the frontend as a
special case.

Two repos are touched:

- **`mate-espidf-base`**: fold `wifi_try_init` into the settings snapshot,
  remove the now-redundant `try_connect_on_init` BLE characteristic.
- **`mate-things`**: the new `/ble-direct` page (the bulk of the work).

## Firmware change (`mate-espidf-base`)

**Problem:** `wifi_try_init` is one of the 7 keys in the preloaded config
schema (`domain/models/preloaded.h`'s `DOMAIN_MODELS_PRELOADED_SCHEMA`),
but its live value is only reachable through a *separate* GATT
characteristic on the WiFi manager service
(`pres_ble_gatt_uuid_wifi_try_connect_on_init_chr`, R/W, immediate/no
restart) — not through the settings service's `data`/`update`
characteristics like every other schema key. A schema-driven Settings form
in the frontend would have to special-case this one key to read/write it
from a different service, defeating the "don't hardcode the schema" goal.

**Fix:** move it into the settings snapshot/update, and delete the
redundant characteristic — it's dead weight once settings covers it (see
call-graph analysis below).

**Why the wifi_manager characteristic can be removed outright:** grepping
every caller of `dom_usecases_internal_wifi_manager_t`'s
`get_try_connect_on_init`/`set_try_connect_on_init` shows the *only*
caller of either is the BLE characteristic's own access callback
(`presentation/ble/handler/wifi_manager/handler.c:324,344`). Everything
else that matters — the WiFi status JSON's `try_connect_on_init_enabled`
field (`application/internal/wifi_manager/impl.c:264`) and the actual
boot/reconnect decision logic (`impl.c:499`, inside `need_reconnect_impl`)
— reads `preloaded_repository->get_wifi_sta_try_connect_on_init()`
*directly*, bypassing the usecase methods entirely. So the value's storage,
its read-only visibility, and its behavioral effect are all unaffected by
removing the characteristic and its usecase wrapper methods.

### Changes

`domain/usecases/internal/settings.h`:
- `dom_usecases_internal_settings_snapshot_t` gains `bool
  wifi_sta_try_connect_on_init;`.
- `dom_usecases_internal_settings_preloaded_update_t` gains `bool
  wifi_try_init_set; bool wifi_try_init;`.

`application/internal/settings/impl_utils.c`:
- `app_internal_settings_impl_load_snapshot()`: populate the new field via
  the already-existing `preloaded_repository->get_wifi_sta_try_connect_on_init()`.
- `app_internal_settings_impl_has_preloaded_update()`: include
  `wifi_try_init_set` in the OR chain.

`application/internal/settings/impl.c`:
- `set_preloaded_impl()`: new branch — if `update->wifi_try_init_set`, call
  the already-existing `preloaded_repository->set_wifi_sta_try_connect_on_init()`
  and set `ctx->restart_required = true` (same convention as every other
  field in this function — confirmed via user decision, not a special
  case).

`presentation/ble/handler/settings/dto.c`:
- `pres_ble_handler_settings_dto_encode_snapshot()`: add `wifi_try_init`
  bool to the output JSON.
- `pres_ble_handler_settings_dto_decode_update()`: parse `wifi_try_init`
  bool from the update JSON, same pattern as the existing
  `system_restart_after_ms` number field.

`presentation/mqtt/handler/config/handler.c`:
- The `DOMAIN_MODELS_PRELOADED_VALUE_TYPE_BOOL` case currently has a
  comment explaining this exact gap and returns early with a warning log.
  Replace it with a real handler: populate `update.wifi_try_init_set` /
  `update.wifi_try_init` and fall through to the existing
  `ctx->settings->set_preloaded(...)` call at the bottom of the function,
  same as every other key. This fully closes a gap the firmware code
  already flagged, not just the BLE path.

**Removals** (the redundant characteristic and its exclusive callers):
- `presentation/ble/gatt/uuid.h` / `uuid.c`: remove
  `pres_ble_gatt_uuid_wifi_try_connect_on_init_chr`.
- `presentation/ble/handler/wifi_manager/handler.c`: remove
  `try_connect_on_init_access_callback` and its prototype; remove its
  `characteristic_defs[]` slot (array shrinks from 6→5 elements, i.e. 4
  characteristics + terminator instead of 5 + terminator); update the
  `deinit()` loop bound from `< 5` to `< 4`.
- `domain/usecases/internal/wifi_manager.h`: remove
  `get_try_connect_on_init`/`set_try_connect_on_init` from the contract
  struct.
- `application/internal/wifi_manager/impl.c`: remove
  `get_try_connect_on_init_impl`/`set_try_connect_on_init_impl` and their
  registration (`self->get_try_connect_on_init = ...` /
  `self->set_try_connect_on_init = ...`) and prototypes.
- `get_status_impl` (`impl.c:264`) and `need_reconnect_impl` (`impl.c:499`)
  are **untouched** — they already read the repository directly, not
  through the removed usecase methods.

**Not touched:** `docs/agent_test/v1.0.0-dev.1/scenario/10-ble-gatt-services.md`
is a historical test-run record (BLE-08 tested the now-removed
characteristic) — it stays as-is, it documents what was true at that pass,
not living API docs.

## Frontend: BLE Direct page (`mate-things`)

### Nav entry

`src/config/navigation.ts`, "Operations" group: add `{ label: "BLE
Direct", href: "/ble-direct", requiredPermissions: [] }`. No specific
permission — any authenticated user can see it, since the page never
touches the `mate-things` backend and there's nothing to gate server-side.

### Module layout

```
frontend/src/app/(authenticated)/ble-direct/
  page.tsx                        # server component: session check + PageHeader, renders the client root
  _components/
    BleDirectRoot.tsx             # client root: owns useBleConnection(), lays out banner + 4 blocks
    ConnectionBanner.tsx          # Connect/Connecting/Connected/reconnecting(reason)/Disconnect chrome
    blocks/
      SystemInfoBlock.tsx
      WifiManagerBlock.tsx
      SettingsBlock/
        SettingsBlock.tsx        # combines schema + snapshot into rows, owns the edit form + restart banner
        SettingsField.tsx        # one row: widget from {key,type}, handles the <key>_set secret convention
      LogBlock.tsx
  _lib/
    ble/
      uuid.ts                    # the 16 UUID strings (post firmware change), derived from gatt/uuid.c's formula
      protocol.ts                # characteristic UUIDs per service, wifi command opcodes, value-type strings
      codec.ts                   # pure encode/decode: JSON<->ArrayBuffer (UTF-8), uint8/uint32/bool framing
      BleClient.ts                # vanilla TS class: requestDevice, connect/discover, read/write/subscribe, reconnect loop
    useBleConnection.ts            # React hook: instantiates one BleClient per mount, exposes state via useSyncExternalStore
```

`BleClient` is deliberately React-free — the reconnect loop and
`gattserverdisconnected` handling are driven by browser events outside any
React render cycle, so keeping this logic in a plain class with its own
tiny subscribe/notify emitter (no external library needed) is the cleanest
seam. `useBleConnection` is the only file that touches `useState`/
`useSyncExternalStore`; every block component receives plain data + action
callbacks as props.

### UUID reference (derived from `gatt/uuid.c`'s
`4d415445-SSSS-4700-CCCC-000000000000` formula; verified against
`docs/agent_test/.../10-ble-gatt-services.md`'s protocol table)

| Service | UUID (SSSS) | Characteristics (CCCC) |
|---|---|---|
| Settings | `0001` | `0001` data (R), `0002` update (W, JSON), `0003` restart_required (R/notify), `0004` restart (W, uint32 LE delay_ms) |
| WiFi manager | `0002` | `0001` status (R/notify), `0002` connect (W, JSON `{ssid,password}`), `0003` command (W, uint8 opcode: 0=stop,1=start,2=connect_stored,3=disconnect,4=forget_stored), `0004` stored_credential (R) |
| Log | `0003` | `0001` message (R/notify), `0002` enabled (R/W, uint8 bool) |
| System info | `0004` | `0001` info (R, JSON), `0002` config_schema (R, JSON array of `{key,type}`) |

Full UUID string per entry: `4d415445-{SSSS}-4700-{CCCC}-000000000000`
(lowercase, as Web Bluetooth requires). WiFi manager's `0005`
(`try_connect_on_init`) is gone post firmware-change and is not in this
table.

### Device discovery

`navigator.bluetooth.requestDevice({ filters: [{ namePrefix: "matedev_" }],
optionalServices: [SETTINGS_SVC, WIFI_SVC, LOG_SVC, SYSTEM_INFO_SVC] })`.
Web Bluetooth has no API for a custom in-page scan list — this opens the
browser's own native chooser, already filtered to `matedev_*` names, which
satisfies "shows available devices matching `matedev_*` and ignores the
rest" directly. Must be called from a real user gesture (a button click).

### Secure-context / unsupported-browser handling

Web Bluetooth requires a secure context (HTTPS or `localhost`) and a
Chromium-based browser. On mount, feature-detect `"bluetooth" in
navigator`; if absent, render a `{ status: "unsupported" }` state with a
clear explanatory message instead of a broken Connect button. Production
HTTPS termination in front of this app is a deployment concern outside
this repo's `compose.yml` and out of scope here.

### Connection lifecycle / state machine

```ts
type DisconnectReason = "lost" | "restarting";

type BleClientState =
  | { status: "unsupported" }
  | { status: "disconnected" }
  | { status: "requesting" }
  | { status: "connecting" }
  | { status: "connected"; data: ConnectedData }
  | { status: "reconnecting"; data: ConnectedData; reason: DisconnectReason };
```

- **`disconnected` → `requesting`**: user clicks Connect → `requestDevice()`.
  Cancelled/no match (`NotFoundError`) → back to `disconnected`, toast "No
  device selected." Not a hard error state.
- **`requesting` → `connecting`**: `device.gatt.connect()`, then discover
  all 4 services + characteristics. Every operation in this phase races a
  **10s manual timeout** (Web Bluetooth has none built in). Reads:
  `system_info.info`, `system_info.config_schema`, `settings.data`,
  `settings.restart_required`, `wifi.status`, `wifi.stored_credential`.
  Subscribes notify: `settings.restart_required`, `wifi.status`,
  `log.message`. Writes `log.enabled = true` (auto-enable streaming).
  **Any missing service/characteristic** (name matched but wrong/older
  firmware) → hard error, `gatt.disconnect()`, back to `disconnected` with
  an inline error, not `reconnecting` — we were never actually up.
- **`connecting` → `connected`**: all of the above succeeded.
- **`connected`/`reconnecting` → `reconnecting("lost")`**:
  `gattserverdisconnected` fires and `expectingRestart` is not armed.
- **`connected`/`reconnecting` → `reconnecting("restarting")`**:
  `gattserverdisconnected` fires while `expectingRestart` is armed (see
  Restart flow below).
- **`reconnecting` loop**: every 3s, `device.gatt.connect()` again (same
  `BluetoothDevice` object — no new chooser prompt; permission persists for
  the page session). On success, re-run the full discovery/read/subscribe
  sequence (picks up anything that changed while disconnected, e.g. via
  MQTT), then → `connected`. A guard flag prevents overlapping attempts.
  Retries indefinitely (fixed 3s interval) until success or the user clicks
  Disconnect.
- **any state → `disconnected`**: user clicks Disconnect. Cancels the
  reconnect timer, `gatt.disconnect()` if connected, drops the
  `BluetoothDevice` reference entirely. The next Connect always goes
  through `requestDevice()` again — no attempt to silently resume a prior
  session.
- **Stale-data rendering**: during `reconnecting`, blocks keep showing
  last-known values at reduced opacity rather than clearing, avoiding
  flicker on a brief drop.

### Restart signaling

`BleClient.restartDevice(delayMs)`:

```ts
async restartDevice(delayMs: number): Promise<void> {
  await this.writeCharacteristic(SETTINGS_RESTART_CHR, encodeUint32LE(delayMs));
  this.expectingRestart = true;
  clearTimeout(this.expectingRestartTimer);
  // Safety net: if no disconnect follows within delay_ms + slack, the
  // write likely didn't trigger a reboot — don't let a later, unrelated
  // drop get mislabeled as "restarting".
  this.expectingRestartTimer = setTimeout(() => {
    this.expectingRestart = false;
  }, delayMs + 5000);
}
```

The `gattserverdisconnected` listener checks and clears `expectingRestart`
at the moment it fires, choosing the `reason` accordingly. Both reasons
drive the identical reconnect loop — only the label differs. This is the
single code path used both by the Settings block's restart-required banner
and any other future restart trigger.

`ConnectionBanner` renders each reason distinctly:
- `"lost"` → "Connection lost — reconnecting…" (warning tone).
- `"restarting"` → "Device is restarting — reconnecting…" (calm tone; early
  failed reconnect attempts are expected here since the ESP32 won't
  resume advertising for the first couple seconds of boot).
- On successful reconnect: one-shot toast — "Device restarted
  successfully" for `"restarting"`, "Reconnected" for `"lost"`.

### Where errors surface

- Connection-level (chooser cancelled, discovery failed, timeout) →
  `ConnectionBanner` inline + toast for transient ones.
- Per-action (a settings write rejected, malformed WiFi credential, etc.)
  → handled locally by the block that issued the write (try/catch around
  the `BleClient` call, inline error near the relevant control), matching
  this app's existing form-error convention. `BleClient` itself only
  rejects promises — it never touches toast or React state directly.

### Per-block designs

**System Info** — read-only, refreshed automatically on every (re)connect.
Two grouped `dt`/`dd` panels ("Project" and "Chip"), matching the existing
card list pattern (e.g. `NodeClassCard`'s `CountItem`): project
name/version/type/firmware version; chip model/revision/cores/MAC.

**WiFi Manager** — live status (subscribed notify) + actions:
- Status summary: Up/Down badge, connection status, SSID, IP/netmask/
  gateway, RSSI, and `try_connect_on_init_enabled` shown **read-only**
  (editing moved to Settings, per the firmware change above).
- Action buttons: Start / Stop / Connect to stored / Disconnect / Forget
  stored credential — single-byte opcode writes, contextually disabled
  (e.g. "Connect to stored" disabled when `stored_credential.available` is
  false). No confirmation dialogs — a direct hands-on debugging tool.
- "Connect to network" form: SSID (required) + password (optional, open
  networks exist) → `wifi.connect` JSON write.

**Settings — schema-driven, the core requirement:**
- One row per `config_schema` entry; widget chosen purely from
  `{key, type}`, never a hardcoded key list:
  - `"string"` → text input.
  - `"uint32"` → number input (integer, min 0).
  - `"bool"` → checkbox.
  - Any other type string → rendered disabled/read-only with the raw type
    name and a "not editable here" note, instead of crashing or silently
    dropping the field (schema can evolve independently of this frontend).
- Default value = `snapshot[key]` when present. If absent **and**
  `snapshot[\`${key}_set\`]` exists as a boolean, render the secret
  convention: masked input, placeholder "Already set — leave blank to
  keep" / "Not set". Otherwise blank.
- Change tracking: an explicit per-field "touched" flag (set on first
  edit), not a value-diff — required for correct secret-field semantics
  (blank-untouched ≠ blank-cleared) and keeps the update payload minimal.
  Save disabled until something is touched, or if a touched field fails
  basic validation (uint32 must parse as non-negative integer).
- Submit builds a partial JSON of touched fields only (typed per schema:
  numbers as numbers, checkbox as boolean) → `writeSettingsUpdate()`. On
  success: re-read the snapshot (picks up any server-side normalization,
  correctly re-blanks secret fields) and clear touched flags.
- Restart banner: driven by the already-subscribed `restart_required`
  notify — "Some changes require a restart to take effect" + "Restart
  now" button (`restartDevice(0)`), sharing the restart-flavored reconnect
  path above.

**Log** — the one asynchronously-updating block:
- Monospace scrolling console, capped at a fixed buffer (500 lines, drop
  oldest) to bound memory in a long session.
- Auto-scrolls to newest; pauses if the user scrolls up to read history,
  with a "Jump to latest" affordance to resume.
- "Streaming" toggle bound to `log.enabled` (defaults on; rewritten to
  `true` again after every reconnect as part of the connect sequence). A
  "Clear" button empties the local buffer only (no firmware effect).
- Note (not user-facing): per `docs/agent_test/.../10-ble-gatt-services.md`
  BLE-11, log notifications may drop during active WiFi
  reconnect/coexistence events on ESP32-C3 (single shared radio) — not a
  frontend bug if a gap appears during a WiFi state change.

### Page shell

`ConnectionBanner` + Connect/Disconnect control always visible at the top.
Before the first successful connect, the four blocks are replaced by a
single empty-state card explaining the page and prompting "Connect." Once
connected, all four blocks render in order (System Info → WiFi Manager →
Settings → Log), each independently dimmed during `reconnecting`.

## Verification

No formal automated test suite for this feature (explicit scope decision —
real verification is manual, against real hardware, by the requester).
Before handing off: `tsc --noEmit` and `eslint` on the new files, and a dev
server boot to confirm the page renders without runtime errors for the
pre-connect states (empty-state, nav entry) — the actual
device-connected flows require real BLE hardware not available in this
environment and are verified by the requester directly.
