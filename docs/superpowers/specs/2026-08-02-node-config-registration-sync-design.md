# Node config registration sync — design

## Context

`node_config_values` (per-node config key/value rows) and `firmware_config_parameters`
(per-firmware config schema: key + value_type) already exist, along with a
working set/push path: the admin UI or API can call `SetByNodeId`, which
upserts a row and publishes the new value to the device over MQTT
(`/sub/{device_id}/config`, consumed by the firmware's
`pres_mqtt_handler_config`).

What's missing is the reverse direction: nothing currently populates
`node_config_values` from what the device itself is actually running. A
device's registration message (sent over MQTT on every boot/reconnect,
handled by `MessagingCallback.Register`) carries only
`{device_id, device_info, firmware_name}` — no config. So today,
`node_config_values` only ever reflects what an admin explicitly set
through the UI/API; if a device's real NVS-stored config drifts from that
(or a node has never had any values set at all), the backend has no way to
know or reconcile it.

This is scoped narrowly, per explicit direction: **no schema migration, no
default-value concept anywhere.** The device already knows its own current
config (it's running with it); it just needs to say so at registration, and
the backend needs to listen and upsert accordingly. Every registration
(every reconnect, not just first-time device creation) re-syncs the DB to
match whatever the device reports, which is what keeps BLE-set and
DB-stored config from drifting apart over time — the device is the single
source of truth for its own values.

Two repos, two directions:

- **`mate-espidf-base`**: registration payload gains a `config` object
  (device → backend).
- **`mate-things`**: registration handler upserts `node_config_values` from
  it (validated against the node's current firmware schema), and the
  already-existing set→push path (backend → device) gets a careful
  correctness pass, fixing anything found broken.

## Firmware (`mate-espidf-base`): registration payload gains `config`

**Current payload** (`inf_messaging_def_pub_mqtt_impl_build_registration_json`,
published to `/pub/registration` from `publish_registration_impl` in
`application/internal/messaging_callbacks/impl.c`):

```json
{"device_id": "...", "device_info": "...", "firmware_name": "..."}
```

**New payload:**

```json
{
  "device_id": "...", "device_info": "...", "firmware_name": "...",
  "config": {
    "mqtt_proto": "mqtt", "mqtt_host": "192.168.1.1", "mqtt_port": "1883",
    "mqtt_user": "", "mqtt_pass": "", "sys_rst_aft_ms": "4294967295",
    "wifi_try_init": "false"
  }
}
```

Every value is a **string**, regardless of the field's real type — this
matches the convention every other config-value transport in this codebase
already uses (`node_config_values.value` is `TEXT`; the MQTT config-set
payload's `value` field is a string; `SetByNodeId`'s type validation parses
strings like `"true"`/`"1883"`). Keeping registration consistent with that
avoids introducing a second representation.

### Changes

- **`domain/models/preloaded.h`**: add a small pair type,
  `dom_models_preloaded_kv_t { const char* key; const char* value; }`.
- **`domain/contracts/messaging/def_pub.h`**: `registration(...)` gains two
  parameters, `const dom_models_preloaded_kv_t* config, size_t config_count`
  (include `domain/models/preloaded.h` for the new type).
- **`application/internal/messaging_callbacks/impl.c`**'s
  `publish_registration_impl`: after building `device_info`/`firmware_name`
  as it already does, read all 7 current preloaded values through
  `ctx->cfg.preloaded_repository`'s existing getters (the same getters
  `settings`'s `get_snapshot` uses — this is the live, NVS-backed
  "currently configured" value, not a boot-time snapshot), stringify the
  two non-string ones (`system_restart_after_ms` via `snprintf` with
  `%lu`; `wifi_sta_try_connect_on_init` as `"true"`/`"false"`), assemble a
  fixed 7-entry `dom_models_preloaded_kv_t config[]` array keyed by the
  existing `DOMAIN_MODELS_PRELOADED_*_KEY` macros, and pass it to
  `def_pub->registration(...)`. This mirrors the explicit
  per-key/per-type style `impl_utils.c`'s `load_snapshot` and the MQTT
  config handler already use in this codebase — no new generic
  X-macro-driven serialization introduced.
- **`infrastructure/messaging/def_pub/mqtt_impl.c`** +
  **`mqtt_impl_utils.c`/`.h`**: thread `config`/`config_count` through
  `registration_impl` into
  `inf_messaging_def_pub_mqtt_impl_build_registration_json`, which adds a
  nested `"config"` object (one string field per entry) to the existing
  JSON.
- **`infrastructure/messaging/def_pub/stub_impl.c`**: update the stub's
  `registration_impl` signature to match (accepts and ignores the new
  params — nothing currently reads recorded registration state from the
  test-seed stub, so no new stub fields are added for it).

## Backend (`mate-things`): registration upserts `node_config_values`

- **`presentation/mqtt/dto/registration.go`**: `Registration` gains
  `Config map[string]string \`json:"config"\`` (a plain JSON object decodes
  natively into this; no new parsing logic needed beyond the existing
  `json.Unmarshal`).
- **`domain/usecases/node/messaging_callback.go`**:
  `RegisterNodeMessageRequest` gains `Config map[string]string`.
- **`presentation/mqtt/dto/registration.go`**'s `DecodeRegistration`:
  passes `dto.Config` through unchanged (an absent/missing `config` key
  unmarshal to a nil map — handled as "nothing to sync", not an error;
  older/incompatible firmware without this field must not break
  registration).
- **`application/node/messaging_callback/usecase.go`**'s `Register`: right
  after `UpsertRegistration` resolves `node` (which carries
  `node.FirmwareId`), and before the existing `if created { subscribeNode }`
  block, sync config:
  1. Read `firmware_config_parameters` for `node.FirmwareId` (needs a new
     `parameterRepository domaincontractsrepository.FirmwareConfigParameter`
     dependency on this usecase).
  2. For each `(key, value)` in `request.Config`: if `key` isn't in the
     firmware's schema, log a warning and skip (mirrors the firmware's own
     MQTT config handler's "unknown key" behavior — never a hard failure,
     since a device can legitimately report keys from a firmware version
     the backend's current schema doesn't recognize yet). If it matches, validate `value` against
     the schema entry's `value_type` using the **shared** validator (see
     below); on a validation failure, log a warning and skip that one key
     (a single malformed value must not abort syncing the rest, or block
     the ack).
  3. Upsert every key that passed validation via `NodeConfigValue.Upsert`
     (needs a new `nodeConfigValueRepository domaincontractsrepository.NodeConfigValue`
     dependency — same repository `config_value`'s usecase already uses).
  4. A repository-level error (not a validation/unknown-key case — an
     actual read/write failure) logs and returns the error, same
     early-return strictness the rest of `Register` already uses for its
     other steps.
  - This runs on **every** registration, not gated on `created` — an
    existing device reconnecting with unchanged config produces harmless
    no-op upserts (same value written back); one that reports a changed
    value (e.g. set via BLE since its last reconnect) gets the DB brought
    back in sync, which is the actual point of this change.
- **Shared validator extraction**: `config_value/usecase.go`'s
  private `validateConfigValue(value, valueType string) error` moves to
  `application/shared/validation.go` as exported
  `ValidateConfigValue(value string, valueType string) error`, so both
  `config_value` and `messaging_callback` usecases call the same one
  instead of duplicating the type-check switch. `config_value/usecase.go`'s
  call site updates to `applicationshared.ValidateConfigValue(...)`; its
  existing table test (`usecase_test.go`,
  `TestValidateConfigValuePreservesStringDomainAndRejectsInvalidTypedValues`)
  updates its call site the same way so it keeps passing — this is a
  mechanical fix to keep the existing test green through the move, not a
  new test.
- **`composition/main/application.go`**: `nodeMessagingCallback`'s
  `NewUsecaseImpl(...)` call gains two new arguments,
  `l.infra.firmwareConfigParameterRepository` and
  `l.infra.nodeConfigValueRepository` (both already constructed earlier in
  the same file for other usecases — no new repository wiring needed, just
  passing existing instances to one more consumer).

## Verification pass: the existing set→push path (backend → device)

Traced end-to-end against the current code, no bug found so far, but this
gets re-confirmed once the registration-sync changes land (they touch nodes
and configs are exercised the same session, on the same running
containers) — the full chain is:

`SetByNodeId` (validates via the now-shared `ValidateConfigValue`, upserts
`node_config_values`, then calls) → `Publish.Config` (MQTT publish to
`/sub/{device_id}/config`, payload `{"key": "...", "value": "..."}`) →
firmware's `on_message.c` (topic-matches `ctx->config_topic`, dispatches
to) → `pres_mqtt_handler_config` → looks up the key in
`dom_models_preloaded_schema`, type-switches, and for the one matching key
this session's earlier work fixed (`wifi_try_init`) → calls
`ctx->settings->set_preloaded(...)`. Topic naming
(`infrastructurenodeshared.NodeSubTopic(deviceId, "config")` on the Go side
vs. `/sub/%s/config` built in `presentation/mqtt/context.c` on the firmware
side) matches exactly. No changes anticipated here; this is a
confirm-and-fix-if-broken pass, not new functionality.

## Verification

No new automated tests are being planned (explicit scope decision) — the
bar is that everything **builds and runs**:

- **`mate-espidf-base`**: `idf.py reconfigure && idf.py build` (or, if the
  compiler toolchain isn't installed in the working environment, careful
  manual re-reads of every touched file plus a repo-wide grep for dangling
  references, same fallback used earlier this session for the BLE work).
- **`mate-things` backend**: `go build ./...` and `go vet ./...` clean;
  `go test ./...` passes (covers the one existing test whose call site
  moves, `TestValidateConfigValuePreservesStringDomainAndRejectsInvalidTypedValues`).
- **Live check**: rebuild the Docker containers, flash/run the firmware (or
  exercise it however this session's environment allows), confirm a real
  registration message on `/pub/registration` now carries a `config` object,
  and that `node_config_values` gets populated/updated for that node in
  Postgres. Re-exercise `SetByNodeId` (via the node config UI or a direct
  API call) and confirm the device still receives and applies the pushed
  value.
