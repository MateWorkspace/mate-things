# BLE Direct Page Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a client-only "BLE Direct" page to the `mate-things` frontend that talks straight to an ESP32's BLE GATT config surface via the Web Bluetooth API, plus a small coordinated firmware change in `mate-espidf-base` that folds a stray config key into the settings snapshot and removes a now-redundant characteristic.

**Architecture:** Two repos. `mate-espidf-base` gets a mechanical struct/DTO extension (Task 1) and a characteristic/usecase removal (Task 2). `mate-things` gets a new `/ble-direct` route built from a React-free `BleClient` class (owns GATT discovery, reads/writes/notify subscriptions, and a reconnect state machine) wrapped by one `useSyncExternalStore`-based hook, with four presentational block components (System Info, WiFi Manager, Settings, Log) driven entirely by props.

**Tech Stack:** ESP-IDF v6.0.2 / NimBLE (firmware), Next.js App Router / React 19 / TypeScript / Web Bluetooth API (frontend).

## Global Constraints

- Web Bluetooth requires a Chromium-based browser and a secure context (HTTPS or `localhost`) — feature-detect via `"bluetooth" in navigator`, never assume it exists.
- Device discovery is filtered to `namePrefix: "matedev_"` only — this is the browser's native device chooser, not a custom in-page list.
- The BLE Direct page makes zero calls to the `mate-things` backend API. `requiredPermissions: []` in nav config — any authenticated user can see it.
- UUID formula (from `mate-espidf-base/main/include/presentation/ble/gatt/uuid.h`): `4d415445-{SSSS:04x}-4700-{CCCC:04x}-000000000000`, all lowercase, where SSSS = service id, CCCC = characteristic id (`0000` for the service's own UUID).
- No automated test suite for this feature (explicit product decision) — verification is `tsc --noEmit` + `eslint` + a dev-server render check only. Real functional verification happens against real hardware, outside this plan.
- Settings service field-set convention: every field `settings->set_preloaded` writes marks `restart_required = true` — the new `wifi_try_init` field follows this, no exception.
- Firmware toolchain: ESP-IDF v6.0.2, `idf.py reconfigure && idf.py build` per `mate-espidf-base/AGENTS.md`. There is no host/unit-test framework for the firmware.
- Frontend toolchain quirk (this environment only): `frontend/node_modules` has `next`/`typescript`/`eslint` properly installed, but there is no system `node`/`npm` on `PATH` — every frontend command in this plan must be run with `PATH=/tmp/mate-node-v24.18.1/bin:$PATH` prepended (a working Node 24 binary already present in this sandbox). Use `node_modules/.bin/tsc`, `node_modules/.bin/eslint`, `node_modules/.bin/next` directly rather than `npm run`.
- When smoke-testing the dev server, always request `http://localhost:<port>/...`, never `http://127.0.0.1:<port>/...` — this frontend silently fails to hydrate over `127.0.0.1` in this environment.

---

## Task 1: Firmware — fold `wifi_try_init` into the settings snapshot/update

**Repo:** `mate-espidf-base`

**Files:**
- Modify: `main/include/domain/usecases/internal/settings.h`
- Modify: `main/src/application/internal/settings/impl_utils.c`
- Modify: `main/src/application/internal/settings/impl.c`
- Modify: `main/src/presentation/ble/handler/settings/dto.c`
- Modify: `main/src/presentation/mqtt/handler/config/handler.c`

**Interfaces:**
- Produces: `dom_usecases_internal_settings_snapshot_t.wifi_sta_try_connect_on_init` (bool), `dom_usecases_internal_settings_preloaded_update_t.wifi_try_init_set` / `.wifi_try_init` (bool), settings `data` JSON gains a `"wifi_try_init"` boolean key, settings `update` JSON accepts a `"wifi_try_init"` boolean key. Task 2 depends on none of this — it only removes code, no shared symbols.

- [ ] **Step 1: Extend the settings usecase structs**

In `main/include/domain/usecases/internal/settings.h`, add a field to `dom_usecases_internal_settings_snapshot_t`:

```c
struct dom_usecases_internal_settings_snapshot_t {
    uint64_t device_id;
    char     device_id_str[32];

    char mqtt_proto[16];
    char mqtt_host[128];
    char mqtt_port[8];
    char mqtt_user[64];
    char mqtt_pass[128];

    uint32_t system_restart_after_ms;
    bool     wifi_sta_try_connect_on_init;
};
```

And a field pair to `dom_usecases_internal_settings_preloaded_update_t`:

```c
struct dom_usecases_internal_settings_preloaded_update_t {
    bool mqtt_proto_set;
    char mqtt_proto[16];

    bool mqtt_host_set;
    char mqtt_host[128];

    bool mqtt_port_set;
    char mqtt_port[8];

    bool mqtt_user_set;
    char mqtt_user[64];

    bool mqtt_pass_set;
    char mqtt_pass[128];

    bool     system_restart_after_ms_set;
    uint32_t system_restart_after_ms;

    bool wifi_try_init_set;
    bool wifi_try_init;
};
```

(`#include <stdbool.h>` is already present in this header — no new include needed.)

- [ ] **Step 2: Populate the new snapshot field and update the "has update" check**

In `main/src/application/internal/settings/impl_utils.c`, at the end of `app_internal_settings_impl_load_snapshot`, right before its final `return DOMAIN_MODELS_ERROR_OK;`, add:

```c
    err = ctx->cfg.preloaded_repository->get_wifi_sta_try_connect_on_init(ctx->cfg.preloaded_repository, &out->wifi_sta_try_connect_on_init);
    if (err != DOMAIN_MODELS_ERROR_OK) {
        return err;
    }

    return DOMAIN_MODELS_ERROR_OK;
```

(replacing the existing final `return DOMAIN_MODELS_ERROR_OK;` — i.e. insert the new read before it, keep one final return.)

In the same file, update `app_internal_settings_impl_has_preloaded_update`:

```c
bool app_internal_settings_impl_has_preloaded_update(const dom_usecases_internal_settings_preloaded_update_t* update) {
    return update &&
           (update->mqtt_proto_set ||
            update->mqtt_host_set ||
            update->mqtt_port_set ||
            update->mqtt_user_set ||
            update->mqtt_pass_set ||
            update->system_restart_after_ms_set ||
            update->wifi_try_init_set);
}
```

Also add `preloaded_repository->get_wifi_sta_try_connect_on_init` and `->set_wifi_sta_try_connect_on_init` to the existing `has_preloaded_repository_functions` validation helper in the same file:

```c
static bool has_preloaded_repository_functions(dom_contracts_repository_preloaded_t* preloaded_repository) {
    return preloaded_repository &&
           preloaded_repository->get_device_id &&
           preloaded_repository->get_device_id_str &&
           preloaded_repository->get_mqtt_proto &&
           preloaded_repository->set_mqtt_proto &&
           preloaded_repository->get_mqtt_host &&
           preloaded_repository->set_mqtt_host &&
           preloaded_repository->get_mqtt_port &&
           preloaded_repository->set_mqtt_port &&
           preloaded_repository->get_mqtt_user &&
           preloaded_repository->set_mqtt_user &&
           preloaded_repository->get_mqtt_pass &&
           preloaded_repository->set_mqtt_pass &&
           preloaded_repository->get_system_restart_after_ms &&
           preloaded_repository->set_system_restart_after_ms &&
           preloaded_repository->get_wifi_sta_try_connect_on_init &&
           preloaded_repository->set_wifi_sta_try_connect_on_init;
}
```

(`preloaded_repository` already exposes both of these — confirmed via `application/internal/wifi_manager/impl.c`'s existing use of the same two functions. This just makes the settings usecase declare the dependency it now also relies on.)

- [ ] **Step 3: Write the new field in `set_preloaded_impl`**

In `main/src/application/internal/settings/impl.c`, in `set_preloaded_impl`, add a new branch right before the existing `if (restart_required_out) { ... }` block at the end of the function:

```c
    if (update->wifi_try_init_set) {
        err = ctx->cfg.preloaded_repository->set_wifi_sta_try_connect_on_init(ctx->cfg.preloaded_repository, update->wifi_try_init);
        if (err != DOMAIN_MODELS_ERROR_OK) {
            ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to set WiFi try-connect-on-init: %s (%d)", dom_models_error_str(err), (int)err);
            return err;
        }
        ctx->restart_required = true;
    }

    if (restart_required_out) {
        *restart_required_out = ctx->restart_required;
    }
```

- [ ] **Step 4: Encode/decode the new field over BLE**

In `main/src/presentation/ble/handler/settings/dto.c`, in `pres_ble_handler_settings_dto_encode_snapshot`, add one line right before `bool ok = cJSON_PrintPreallocated(...)`:

```c
    cJSON_AddBoolToObject(root, "wifi_try_init", snapshot->wifi_sta_try_connect_on_init);
```

In the same file, in `pres_ble_handler_settings_dto_decode_update`, add right before `cJSON_Delete(root);`:

```c
    cJSON* wifi_try_init_item = cJSON_GetObjectItemCaseSensitive(root, "wifi_try_init");
    if (cJSON_IsBool(wifi_try_init_item)) {
        out->wifi_try_init     = cJSON_IsTrue(wifi_try_init_item);
        out->wifi_try_init_set = true;
    }
```

- [ ] **Step 5: Wire the MQTT config handler's known gap**

In `main/src/presentation/mqtt/handler/config/handler.c`, replace the entire `case DOMAIN_MODELS_PRELOADED_VALUE_TYPE_BOOL:` block:

```c
        case DOMAIN_MODELS_PRELOADED_VALUE_TYPE_BOOL: {
            if (strcmp(request.value, "true") != 0 && strcmp(request.value, "false") != 0) {
                ctx->logger->warn(ctx->logger, tag, "Invalid config value for %s: %s", request.key, request.value);
                return;
            }

            if (strcmp(request.key, DOMAIN_MODELS_PRELOADED_WIFI_STA_TRY_CONNECT_ON_INIT_KEY) == 0) {
                /* The underlying setting IS writable - dom_usecases_internal_wifi_manager_t
                   .set_try_connect_on_init exists and BLE's wifi_manager handler already
                   calls it. The gap is wiring only: pres_mqtt_context_t holds no
                   wifi_manager reference, so this handler cannot reach it, and
                   dom_usecases_internal_settings_preloaded_update_t (the only path it
                   does have) carries no wifi_sta_try_connect_on_init field.
                   Log and skip rather than silently no-op. */
                ctx->logger->warn(ctx->logger, tag, "Config key %s is not writable via MQTT (MQTT context has no wifi_manager reference; use BLE)", request.key);
                return;
            }
            break;
        }
```

with:

```c
        case DOMAIN_MODELS_PRELOADED_VALUE_TYPE_BOOL: {
            if (strcmp(request.value, "true") != 0 && strcmp(request.value, "false") != 0) {
                ctx->logger->warn(ctx->logger, tag, "Invalid config value for %s: %s", request.key, request.value);
                return;
            }

            if (strcmp(request.key, DOMAIN_MODELS_PRELOADED_WIFI_STA_TRY_CONNECT_ON_INIT_KEY) == 0) {
                update.wifi_try_init     = strcmp(request.value, "true") == 0;
                update.wifi_try_init_set = true;
                handled                  = true;
            }
            break;
        }
```

- [ ] **Step 6: Verify**

Try to build:

```bash
source /home/dodol/.espressif/v6.0.2/esp-idf/export.sh
cd /home/dodol/Repositories/mate/mate-espidf-base
idf.py reconfigure && idf.py build
```

If the compiler toolchain isn't installed in this environment (it wasn't, as of this plan being written — `xtensa-esp-elf`/`riscv32-esp-elf` etc. reported "no installed versions"), `idf.py build` cannot run here. In that case, verify by careful manual inspection instead — re-read every file touched in this task end-to-end and confirm: every new struct field has a matching read/write site, every `err != DOMAIN_MODELS_ERROR_OK` early-return path is preserved, brace/paren balance is correct, and no existing field's behavior changed. State plainly in your final report whether `idf.py build` actually ran, and if not, that verification was manual-only — do not claim a build succeeded without having run it.

- [ ] **Step 7: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-espidf-base
git add main/include/domain/usecases/internal/settings.h \
        main/src/application/internal/settings/impl_utils.c \
        main/src/application/internal/settings/impl.c \
        main/src/presentation/ble/handler/settings/dto.c \
        main/src/presentation/mqtt/handler/config/handler.c
git commit -m "feat(settings): fold wifi_try_init into the settings snapshot/update"
```

---

## Task 2: Firmware — remove the redundant `try_connect_on_init` characteristic

**Repo:** `mate-espidf-base`

**Files:**
- Modify: `main/include/presentation/ble/gatt/uuid.h`
- Modify: `main/src/presentation/ble/gatt/uuid.c`
- Modify: `main/include/domain/usecases/internal/wifi_manager.h`
- Modify: `main/src/application/internal/wifi_manager/impl.c`
- Modify: `main/src/presentation/ble/handler/wifi_manager/handler.c`

**Interfaces:**
- Consumes: nothing from Task 1.
- Produces: nothing new — this task only removes dead surface. After this task, `dom_usecases_internal_wifi_manager_t` no longer has `get_try_connect_on_init`/`set_try_connect_on_init`, and the WiFi manager BLE service exposes 4 characteristics instead of 5. `get_status_impl` and `need_reconnect_impl` are untouched (they already read `preloaded_repository` directly, confirmed by grep before this plan was written).

- [ ] **Step 1: Remove the UUID**

In `main/include/presentation/ble/gatt/uuid.h`, remove this line from the "WiFi manager service" block:

```c
extern const ble_uuid128_t pres_ble_gatt_uuid_wifi_try_connect_on_init_chr;
```

In `main/src/presentation/ble/gatt/uuid.c`, remove this line from the "WiFi manager service" block:

```c
const ble_uuid128_t pres_ble_gatt_uuid_wifi_try_connect_on_init_chr = PRES_BLE_GATT_UUID128(0x00, 0x02, 0x00, 0x05);
```

- [ ] **Step 2: Remove the usecase contract methods**

In `main/include/domain/usecases/internal/wifi_manager.h`, remove these two members from the `dom_usecases_internal_wifi_manager_t` struct:

```c
    dom_models_error_t (*get_try_connect_on_init)(
        dom_usecases_internal_wifi_manager_t* self,
        bool*                                 out
    );
    dom_models_error_t (*set_try_connect_on_init)(
        dom_usecases_internal_wifi_manager_t* self,
        bool                                  enabled
    );
```

(Leave `try_connect_on_init_enabled` in `dom_usecases_internal_wifi_manager_status_t` untouched — that field is populated by `get_status_impl` reading the repository directly, not via the removed methods, and stays as the read-only signal the WiFi status characteristic still reports.)

- [ ] **Step 3: Remove the usecase implementations**

In `main/src/application/internal/wifi_manager/impl.c`:

Remove these two prototypes:

```c
static dom_models_error_t get_try_connect_on_init_impl(
    dom_usecases_internal_wifi_manager_t* self,
    bool*                                 out
);
static dom_models_error_t set_try_connect_on_init_impl(
    dom_usecases_internal_wifi_manager_t* self,
    bool                                  enabled
);
```

Remove these two registration lines from the constructor:

```c
    self->get_try_connect_on_init  = get_try_connect_on_init_impl;
    self->set_try_connect_on_init  = set_try_connect_on_init_impl;
```

Remove the two function bodies:

```c
static dom_models_error_t get_try_connect_on_init_impl(
    dom_usecases_internal_wifi_manager_t* self,
    bool*                                 out
) {
    const char* tag = BASE_TAG "/get_try_connect_on_init";

    app_internal_wifi_manager_impl_ctx_t* ctx = NULL;
    dom_models_error_t                    err = get_ctx(self, &ctx);
    if (err != DOMAIN_MODELS_ERROR_OK) {
        return err;
    }

    if (!out) {
        err = DOMAIN_MODELS_ERROR_BAD_ARGUMENT;
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Missing try-connect-on-init output: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    err = ctx->cfg.preloaded_repository->get_wifi_sta_try_connect_on_init(ctx->cfg.preloaded_repository, out);
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to get try-connect-on-init flag: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    ctx->cfg.logger->info(ctx->cfg.logger, tag, "Try-connect-on-init flag retrieved successfully");

    return DOMAIN_MODELS_ERROR_OK;
}

static dom_models_error_t set_try_connect_on_init_impl(
    dom_usecases_internal_wifi_manager_t* self,
    bool                                  enabled
) {
    const char* tag = BASE_TAG "/set_try_connect_on_init";

    app_internal_wifi_manager_impl_ctx_t* ctx = NULL;
    dom_models_error_t                    err = get_ctx(self, &ctx);
    if (err != DOMAIN_MODELS_ERROR_OK) {
        return err;
    }

    err = ctx->cfg.preloaded_repository->set_wifi_sta_try_connect_on_init(ctx->cfg.preloaded_repository, enabled);
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to set try-connect-on-init flag: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    ctx->cfg.logger->info(ctx->cfg.logger, tag, "Try-connect-on-init flag set successfully");

    return DOMAIN_MODELS_ERROR_OK;
}
```

Leave `get_status_impl` (which reads `try_connect_on_init_enabled` via `preloaded_repository->get_wifi_sta_try_connect_on_init` directly) and `need_reconnect_impl` (same direct-read pattern) completely untouched.

- [ ] **Step 4: Remove the BLE characteristic**

In `main/src/presentation/ble/handler/wifi_manager/handler.c`:

Change the characteristic array size declaration:

```c
static struct ble_gatt_chr_def characteristic_defs[6];
```
→
```c
static struct ble_gatt_chr_def characteristic_defs[5];
```

Remove the prototype:

```c
static int try_connect_on_init_access_callback(uint16_t conn_handle, uint16_t attr_handle, struct ble_gatt_access_ctxt* ctxt, void* arg);
```

Remove the characteristic def slot (index 4) and renumber the terminator to index 4:

```c
    characteristic_defs[4] = (struct ble_gatt_chr_def){
        .uuid      = &pres_ble_gatt_uuid_wifi_try_connect_on_init_chr.u,
        .access_cb = try_connect_on_init_access_callback,
        .arg       = self,
        .flags     = BLE_GATT_CHR_F_READ | BLE_GATT_CHR_F_WRITE,
    };
    characteristic_defs[5] = (struct ble_gatt_chr_def){0};
```
→
```c
    characteristic_defs[4] = (struct ble_gatt_chr_def){0};
```

Change the deinit loop bound:

```c
    for (size_t i = 0; i < 5; i++) {
        characteristic_defs[i].access_cb = pres_ble_gatt_util_disabled_access_callback;
        characteristic_defs[i].arg       = NULL;
    }
```
→
```c
    for (size_t i = 0; i < 4; i++) {
        characteristic_defs[i].access_cb = pres_ble_gatt_util_disabled_access_callback;
        characteristic_defs[i].arg       = NULL;
    }
```

Remove the entire function body:

```c
static int try_connect_on_init_access_callback(
    uint16_t                     conn_handle,
    uint16_t                     attr_handle,
    struct ble_gatt_access_ctxt* ctxt,
    void*                        arg
) {
    (void)conn_handle;
    (void)attr_handle;

    pres_ble_handler_wifi_manager_t* self = arg;
    if (!self || !ctxt) {
        return BLE_ATT_ERR_REQ_NOT_SUPPORTED;
    }

    if (ctxt->op == BLE_GATT_ACCESS_OP_READ_CHR) {
        bool               enabled = false;
        dom_models_error_t err     = self->cfg.wifi_manager->get_try_connect_on_init(self->cfg.wifi_manager, &enabled);
        if (err != DOMAIN_MODELS_ERROR_OK) {
            return BLE_ATT_ERR_UNLIKELY;
        }

        uint8_t value = enabled ? 1 : 0;
        return pres_ble_gatt_util_write_read_response(ctxt, (const char*)&value, sizeof(value));
    }

    if (ctxt->op == BLE_GATT_ACCESS_OP_WRITE_CHR) {
        uint8_t value = 0;
        if (OS_MBUF_PKTLEN(ctxt->om) < sizeof(value)) {
            return BLE_ATT_ERR_INVALID_ATTR_VALUE_LEN;
        }

        int rc = ble_hs_mbuf_to_flat(ctxt->om, &value, sizeof(value), NULL);
        if (rc != 0) {
            return BLE_ATT_ERR_UNLIKELY;
        }

        dom_models_error_t err = self->cfg.wifi_manager->set_try_connect_on_init(self->cfg.wifi_manager, value != 0);
        return err == DOMAIN_MODELS_ERROR_OK ? 0 : BLE_ATT_ERR_UNLIKELY;
    }

    return BLE_ATT_ERR_REQ_NOT_SUPPORTED;
}
```

- [ ] **Step 5: Verify**

Same approach as Task 1, Step 6 — attempt `idf.py reconfigure && idf.py build`; if the toolchain isn't available, verify manually: confirm `characteristic_defs` has exactly 5 entries (4 real + terminator) with correct indices, confirm every remaining callback still references the right `characteristic_defs[i]` index, confirm no remaining reference anywhere in the repo to `pres_ble_gatt_uuid_wifi_try_connect_on_init_chr`, `get_try_connect_on_init`, `set_try_connect_on_init`, or `try_connect_on_init_access_callback` (grep the whole `main/` tree). Report honestly whether the real build ran.

- [ ] **Step 6: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-espidf-base
git add main/include/presentation/ble/gatt/uuid.h \
        main/src/presentation/ble/gatt/uuid.c \
        main/include/domain/usecases/internal/wifi_manager.h \
        main/src/application/internal/wifi_manager/impl.c \
        main/src/presentation/ble/handler/wifi_manager/handler.c
git commit -m "refactor(wifi_manager): remove try_connect_on_init BLE characteristic

Superseded by wifi_try_init now being part of the settings snapshot/update
(see previous commit). The removed usecase methods were dead wrappers with
no other caller - get_status_impl and need_reconnect_impl already read the
preloaded repository directly and are unaffected."
```

---

## Task 3: Frontend — Web Bluetooth ambient types + pure protocol/uuid/codec modules

**Repo:** `mate-things`

**Files:**
- Create: `frontend/src/types/web-bluetooth.d.ts`
- Create: `frontend/src/app/(authenticated)/ble-direct/_lib/ble/uuid.ts`
- Create: `frontend/src/app/(authenticated)/ble-direct/_lib/ble/protocol.ts`
- Create: `frontend/src/app/(authenticated)/ble-direct/_lib/ble/codec.ts`

**Interfaces:**
- Produces (consumed by Task 4): from `uuid.ts` — `SETTINGS_SERVICE`, `SETTINGS_DATA_CHR`, `SETTINGS_UPDATE_CHR`, `SETTINGS_RESTART_REQUIRED_CHR`, `SETTINGS_RESTART_CHR`, `WIFI_SERVICE`, `WIFI_STATUS_CHR`, `WIFI_CONNECT_CHR`, `WIFI_COMMAND_CHR`, `WIFI_STORED_CREDENTIAL_CHR`, `LOG_SERVICE`, `LOG_MESSAGE_CHR`, `LOG_ENABLED_CHR`, `SYSTEM_INFO_SERVICE`, `SYSTEM_INFO_INFO_CHR`, `SYSTEM_INFO_CONFIG_SCHEMA_CHR` (all `string`), `ALL_SERVICE_UUIDS: string[]`, `DEVICE_NAME_PREFIX: string`. From `protocol.ts` — types `WifiCommand`, `ConfigSchemaEntry`, `SystemInfoProject`, `SystemInfoChip`, `SystemInfo`, `SettingsSnapshot` (= `Record<string, unknown>`), `WifiStatus`, `WifiStoredCredential`, and `WIFI_COMMAND_OPCODES: Record<WifiCommand, number>`. From `codec.ts` — `encodeUtf8Json(value: unknown): Uint8Array`, `decodeUtf8Json<T>(view: DataView): T`, `decodeUtf8Text(view: DataView): string`, `encodeUint32LE(value: number): Uint8Array`, `decodeUint32LE(view: DataView): number`, `encodeBoolByte(value: boolean): Uint8Array`, `decodeBoolByte(view: DataView): boolean`.

- [ ] **Step 1: Write the Web Bluetooth ambient type declarations**

The installed TypeScript/`lib.dom.d.ts` in this project does not include the Web Bluetooth API (it's not a finished W3C spec). Rather than adding a new npm dependency, declare exactly the surface this feature uses.

Create `frontend/src/types/web-bluetooth.d.ts`:

```ts
export {};

declare global {
  interface BluetoothLEScanFilter {
    services?: (string | number)[];
    name?: string;
    namePrefix?: string;
  }

  interface RequestDeviceOptions {
    filters?: BluetoothLEScanFilter[];
    optionalServices?: (string | number)[];
    acceptAllDevices?: boolean;
  }

  interface BluetoothRemoteGATTCharacteristic extends EventTarget {
    readonly service: BluetoothRemoteGATTService;
    readonly uuid: string;
    readonly value?: DataView;
    readValue(): Promise<DataView>;
    writeValueWithResponse(value: BufferSource): Promise<void>;
    writeValueWithoutResponse(value: BufferSource): Promise<void>;
    startNotifications(): Promise<BluetoothRemoteGATTCharacteristic>;
    stopNotifications(): Promise<BluetoothRemoteGATTCharacteristic>;
    addEventListener(
      type: "characteristicvaluechanged",
      listener: (event: Event) => void,
    ): void;
    removeEventListener(
      type: "characteristicvaluechanged",
      listener: (event: Event) => void,
    ): void;
  }

  interface BluetoothRemoteGATTService extends EventTarget {
    readonly device: BluetoothDevice;
    readonly uuid: string;
    getCharacteristic(
      characteristic: string | number,
    ): Promise<BluetoothRemoteGATTCharacteristic>;
  }

  interface BluetoothRemoteGATTServer {
    readonly device: BluetoothDevice;
    readonly connected: boolean;
    connect(): Promise<BluetoothRemoteGATTServer>;
    disconnect(): void;
    getPrimaryService(
      service: string | number,
    ): Promise<BluetoothRemoteGATTService>;
  }

  interface BluetoothDevice extends EventTarget {
    readonly id: string;
    readonly name?: string;
    readonly gatt?: BluetoothRemoteGATTServer;
    addEventListener(
      type: "gattserverdisconnected",
      listener: (event: Event) => void,
    ): void;
    removeEventListener(
      type: "gattserverdisconnected",
      listener: (event: Event) => void,
    ): void;
  }

  interface Bluetooth extends EventTarget {
    requestDevice(options: RequestDeviceOptions): Promise<BluetoothDevice>;
  }

  interface Navigator {
    readonly bluetooth?: Bluetooth;
  }
}
```

- [ ] **Step 2: Write the UUID module**

Create `frontend/src/app/(authenticated)/ble-direct/_lib/ble/uuid.ts`:

```ts
/**
 * Every custom UUID this device exposes follows
 * 4d415445-SSSS-4700-CCCC-000000000000, where SSSS identifies the service
 * and CCCC the characteristic within it (0000 for the service UUID
 * itself). Mirrors mate-espidf-base's presentation/ble/gatt/uuid.h/.c —
 * see that file's comment for the byte-order derivation.
 */
function uuid(service: number, characteristic: number): string {
  const svc = service.toString(16).padStart(4, "0");
  const chr = characteristic.toString(16).padStart(4, "0");
  return `4d415445-${svc}-4700-${chr}-000000000000`;
}

export const SETTINGS_SERVICE = uuid(0x0001, 0x0000);
export const SETTINGS_DATA_CHR = uuid(0x0001, 0x0001);
export const SETTINGS_UPDATE_CHR = uuid(0x0001, 0x0002);
export const SETTINGS_RESTART_REQUIRED_CHR = uuid(0x0001, 0x0003);
export const SETTINGS_RESTART_CHR = uuid(0x0001, 0x0004);

export const WIFI_SERVICE = uuid(0x0002, 0x0000);
export const WIFI_STATUS_CHR = uuid(0x0002, 0x0001);
export const WIFI_CONNECT_CHR = uuid(0x0002, 0x0002);
export const WIFI_COMMAND_CHR = uuid(0x0002, 0x0003);
export const WIFI_STORED_CREDENTIAL_CHR = uuid(0x0002, 0x0004);

export const LOG_SERVICE = uuid(0x0003, 0x0000);
export const LOG_MESSAGE_CHR = uuid(0x0003, 0x0001);
export const LOG_ENABLED_CHR = uuid(0x0003, 0x0002);

export const SYSTEM_INFO_SERVICE = uuid(0x0004, 0x0000);
export const SYSTEM_INFO_INFO_CHR = uuid(0x0004, 0x0001);
export const SYSTEM_INFO_CONFIG_SCHEMA_CHR = uuid(0x0004, 0x0002);

export const ALL_SERVICE_UUIDS = [
  SETTINGS_SERVICE,
  WIFI_SERVICE,
  LOG_SERVICE,
  SYSTEM_INFO_SERVICE,
];

export const DEVICE_NAME_PREFIX = "matedev_";
```

- [ ] **Step 3: Write the protocol types module**

Create `frontend/src/app/(authenticated)/ble-direct/_lib/ble/protocol.ts`:

```ts
export type WifiCommand =
  | "stop"
  | "start"
  | "connect_stored"
  | "disconnect"
  | "forget_stored";

export const WIFI_COMMAND_OPCODES: Record<WifiCommand, number> = {
  stop: 0,
  start: 1,
  connect_stored: 2,
  disconnect: 3,
  forget_stored: 4,
};

/**
 * One entry from the system_info config_schema characteristic. `type` is
 * intentionally a plain string, not a narrowed union — the firmware schema
 * can introduce new types this frontend doesn't know how to render yet,
 * and the Settings block must degrade gracefully rather than assume a
 * fixed set.
 */
export interface ConfigSchemaEntry {
  key: string;
  type: string;
}

export interface SystemInfoProject {
  project_name: string;
  project_version: string;
  name: string;
  type: string;
  firmware_version: string;
}

export interface SystemInfoChip {
  hardware_mac: string;
  model: string;
  revision: number;
  cores: number;
}

export interface SystemInfo {
  project: SystemInfoProject;
  chip: SystemInfoChip;
}

/**
 * The settings `data` characteristic's JSON. Deliberately untyped beyond
 * "a JSON object" - which keys exist is driven entirely by the
 * config_schema characteristic (see ConfigSchemaEntry), not hardcoded
 * here. Some keys (e.g. mqtt_pass) never appear directly; only a
 * `<key>_set` boolean companion does, for secrets that are write-only over
 * BLE.
 */
export type SettingsSnapshot = Record<string, unknown>;

export interface WifiStatus {
  is_up: boolean;
  sta_connection_status: "connected" | "disconnected";
  ssid: string;
  ip: string;
  netmask: string;
  gateway: string;
  rssi: number;
  try_connect_on_init_enabled: boolean;
  connect_attempted: boolean;
}

export interface WifiStoredCredential {
  available: boolean;
  ssid: string;
}
```

- [ ] **Step 4: Write the codec module**

Create `frontend/src/app/(authenticated)/ble-direct/_lib/ble/codec.ts`:

```ts
const textEncoder = new TextEncoder();
const textDecoder = new TextDecoder();

export function encodeUtf8Json(value: unknown): Uint8Array {
  return textEncoder.encode(JSON.stringify(value));
}

export function decodeUtf8Text(view: DataView): string {
  return textDecoder.decode(view);
}

export function decodeUtf8Json<T>(view: DataView): T {
  return JSON.parse(decodeUtf8Text(view)) as T;
}

export function encodeUint32LE(value: number): Uint8Array {
  const bytes = new Uint8Array(4);
  new DataView(bytes.buffer).setUint32(0, value, true);
  return bytes;
}

export function decodeUint32LE(view: DataView): number {
  return view.getUint32(0, true);
}

export function encodeBoolByte(value: boolean): Uint8Array {
  return new Uint8Array([value ? 1 : 0]);
}

export function decodeBoolByte(view: DataView): boolean {
  return view.getUint8(0) !== 0;
}
```

- [ ] **Step 5: Verify it compiles and lints**

```bash
cd /home/dodol/Repositories/mate/mate-things/frontend
PATH=/tmp/mate-node-v24.18.1/bin:$PATH node_modules/.bin/tsc --noEmit
PATH=/tmp/mate-node-v24.18.1/bin:$PATH node_modules/.bin/eslint \
  src/types/web-bluetooth.d.ts \
  "src/app/(authenticated)/ble-direct/_lib/ble/uuid.ts" \
  "src/app/(authenticated)/ble-direct/_lib/ble/protocol.ts" \
  "src/app/(authenticated)/ble-direct/_lib/ble/codec.ts"
```

Expected: no new errors from these files (pre-existing unrelated errors in other files, e.g. missing test-only packages, are expected and not this task's concern).

Also sanity-check the UUID derivation directly, since a byte-order mistake here would silently fail to match anything on a real device with no useful error:

```bash
cd /home/dodol/Repositories/mate/mate-things/frontend
PATH=/tmp/mate-node-v24.18.1/bin:$PATH node -e '
function uuid(s, c) {
  return `4d415445-${s.toString(16).padStart(4,"0")}-4700-${c.toString(16).padStart(4,"0")}-000000000000`;
}
console.log(uuid(1, 0));
console.log(uuid(4, 2));
'
```

Expected output:
```
4d415445-0001-4700-0000-000000000000
4d415445-0004-4700-0002-000000000000
```

The first must match `SETTINGS_SERVICE`, the second `SYSTEM_INFO_CONFIG_SCHEMA_CHR` — both are also independently confirmed by `mate-espidf-base/docs/agent_test/v1.0.0-dev.1/scenario/10-ble-gatt-services.md`'s protocol table (Settings=`0001`, System info=`0004`).

- [ ] **Step 6: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-things
git add frontend/src/types/web-bluetooth.d.ts \
        "frontend/src/app/(authenticated)/ble-direct/_lib/ble/uuid.ts" \
        "frontend/src/app/(authenticated)/ble-direct/_lib/ble/protocol.ts" \
        "frontend/src/app/(authenticated)/ble-direct/_lib/ble/codec.ts"
git commit -m "feat(ble-direct): add Web Bluetooth types, UUID/protocol constants, and codec helpers"
```

---

## Task 4: Frontend — `BleClient` class

**Repo:** `mate-things`

**Files:**
- Create: `frontend/src/app/(authenticated)/ble-direct/_lib/ble/BleClient.ts`

**Interfaces:**
- Consumes: everything from Task 3 (`uuid.ts`, `protocol.ts`, `codec.ts`).
- Produces (consumed by Task 5): `export type DisconnectReason = "lost" | "restarting"`, `export interface ConnectedData { systemInfo: SystemInfo; configSchema: ConfigSchemaEntry[]; settingsSnapshot: SettingsSnapshot; restartRequired: boolean; wifiStatus: WifiStatus; wifiStoredCredential: WifiStoredCredential; logLines: string[]; logEnabled: boolean; }`, `export type BleClientState = { status: "unsupported" } | { status: "disconnected"; error?: string } | { status: "requesting" } | { status: "connecting" } | { status: "connected"; data: ConnectedData } | { status: "reconnecting"; data: ConnectedData; reason: DisconnectReason }`, and `export class BleClient` with methods `getState(): BleClientState`, `subscribe(listener: () => void): () => void`, `connect(): Promise<void>`, `disconnect(): void`, `writeSettingsUpdate(partial: Record<string, unknown>): Promise<void>`, `restartDevice(delayMs: number): Promise<void>`, `writeWifiConnect(ssid: string, password: string): Promise<void>`, `sendWifiCommand(command: WifiCommand): Promise<void>`, `setLogEnabled(enabled: boolean): Promise<void>`, `clearLog(): void`.

- [ ] **Step 1: Write the class**

Create `frontend/src/app/(authenticated)/ble-direct/_lib/ble/BleClient.ts`:

```ts
import {
  decodeBoolByte,
  decodeUtf8Json,
  decodeUtf8Text,
  encodeBoolByte,
  encodeUint32LE,
  encodeUtf8Json,
} from "./codec";
import {
  WIFI_COMMAND_OPCODES,
  type ConfigSchemaEntry,
  type SettingsSnapshot,
  type SystemInfo,
  type WifiCommand,
  type WifiStatus,
  type WifiStoredCredential,
} from "./protocol";
import {
  ALL_SERVICE_UUIDS,
  DEVICE_NAME_PREFIX,
  LOG_ENABLED_CHR,
  LOG_MESSAGE_CHR,
  LOG_SERVICE,
  SETTINGS_DATA_CHR,
  SETTINGS_RESTART_CHR,
  SETTINGS_RESTART_REQUIRED_CHR,
  SETTINGS_SERVICE,
  SETTINGS_UPDATE_CHR,
  SYSTEM_INFO_CONFIG_SCHEMA_CHR,
  SYSTEM_INFO_INFO_CHR,
  SYSTEM_INFO_SERVICE,
  WIFI_COMMAND_CHR,
  WIFI_CONNECT_CHR,
  WIFI_SERVICE,
  WIFI_STATUS_CHR,
  WIFI_STORED_CREDENTIAL_CHR,
} from "./uuid";

const RECONNECT_INTERVAL_MS = 3000;
const OPERATION_TIMEOUT_MS = 10000;
const RESTART_GRACE_MS = 5000;
const LOG_BUFFER_MAX = 500;

export type DisconnectReason = "lost" | "restarting";

export interface ConnectedData {
  systemInfo: SystemInfo;
  configSchema: ConfigSchemaEntry[];
  settingsSnapshot: SettingsSnapshot;
  restartRequired: boolean;
  wifiStatus: WifiStatus;
  wifiStoredCredential: WifiStoredCredential;
  logLines: string[];
  logEnabled: boolean;
}

export type BleClientState =
  | { status: "unsupported" }
  | { status: "disconnected"; error?: string }
  | { status: "requesting" }
  | { status: "connecting" }
  | { status: "connected"; data: ConnectedData }
  | { status: "reconnecting"; data: ConnectedData; reason: DisconnectReason };

interface CharacteristicRefs {
  settingsData: BluetoothRemoteGATTCharacteristic;
  settingsUpdate: BluetoothRemoteGATTCharacteristic;
  settingsRestartRequired: BluetoothRemoteGATTCharacteristic;
  settingsRestart: BluetoothRemoteGATTCharacteristic;
  wifiStatus: BluetoothRemoteGATTCharacteristic;
  wifiConnect: BluetoothRemoteGATTCharacteristic;
  wifiCommand: BluetoothRemoteGATTCharacteristic;
  wifiStoredCredential: BluetoothRemoteGATTCharacteristic;
  logMessage: BluetoothRemoteGATTCharacteristic;
  logEnabled: BluetoothRemoteGATTCharacteristic;
  systemInfoInfo: BluetoothRemoteGATTCharacteristic;
  systemInfoConfigSchema: BluetoothRemoteGATTCharacteristic;
}

function withTimeout<T>(
  promise: Promise<T>,
  ms: number,
  message: string,
): Promise<T> {
  return new Promise<T>((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error(message)), ms);
    promise.then(
      (value) => {
        clearTimeout(timer);
        resolve(value);
      },
      (error: unknown) => {
        clearTimeout(timer);
        reject(error instanceof Error ? error : new Error(String(error)));
      },
    );
  });
}

async function readJson<T>(
  characteristic: BluetoothRemoteGATTCharacteristic,
): Promise<T> {
  const view = await characteristic.readValue();
  return decodeUtf8Json<T>(view);
}

async function readBool(
  characteristic: BluetoothRemoteGATTCharacteristic,
): Promise<boolean> {
  const view = await characteristic.readValue();
  return decodeBoolByte(view);
}

async function writeJson(
  characteristic: BluetoothRemoteGATTCharacteristic,
  value: unknown,
): Promise<void> {
  // writeValueWithResponse (never writeValueWithoutResponse) is required:
  // it maps to the ATT "Write Request", which the platform Bluetooth stack
  // automatically splits into the "Write Long Characteristic Values"
  // prepare/execute sub-procedure when the payload exceeds the negotiated
  // MTU. writeValueWithoutResponse ("Write Command") has no such fallback
  // and would silently truncate a settings-update payload over MTU.
  await characteristic.writeValueWithResponse(encodeUtf8Json(value));
}

async function writeBool(
  characteristic: BluetoothRemoteGATTCharacteristic,
  value: boolean,
): Promise<void> {
  await characteristic.writeValueWithResponse(encodeBoolByte(value));
}

export class BleClient {
  private state: BleClientState =
    typeof navigator !== "undefined" && "bluetooth" in navigator
      ? { status: "disconnected" }
      : { status: "unsupported" };
  private listeners = new Set<() => void>();
  private device: BluetoothDevice | null = null;
  private chars: CharacteristicRefs | null = null;
  private reconnectTimer: ReturnType<typeof setInterval> | null = null;
  private reconnecting = false;
  private expectingRestart = false;
  private expectingRestartTimer: ReturnType<typeof setTimeout> | null = null;

  getState(): BleClientState {
    return this.state;
  }

  subscribe(listener: () => void): () => void {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  }

  async connect(): Promise<void> {
    if (this.state.status === "unsupported") return;

    this.setState({ status: "requesting" });

    let device: BluetoothDevice;
    try {
      device = await navigator.bluetooth!.requestDevice({
        filters: [{ namePrefix: DEVICE_NAME_PREFIX }],
        optionalServices: ALL_SERVICE_UUIDS,
      });
    } catch {
      this.setState({ status: "disconnected" });
      return;
    }

    this.device = device;
    device.addEventListener("gattserverdisconnected", this.handleDisconnected);

    await this.performConnect(false);
  }

  disconnect(): void {
    this.stopReconnectLoop();
    this.clearRestartExpectation();
    if (this.device) {
      this.device.removeEventListener(
        "gattserverdisconnected",
        this.handleDisconnected,
      );
      if (this.device.gatt?.connected) {
        this.device.gatt.disconnect();
      }
    }
    this.device = null;
    this.chars = null;
    this.setState({ status: "disconnected" });
  }

  async writeSettingsUpdate(partial: Record<string, unknown>): Promise<void> {
    if (!this.chars) throw new Error("Not connected.");
    await writeJson(this.chars.settingsUpdate, partial);
    const settingsSnapshot = await readJson<SettingsSnapshot>(
      this.chars.settingsData,
    );
    this.updateData((data) => ({ ...data, settingsSnapshot }));
  }

  async restartDevice(delayMs: number): Promise<void> {
    if (!this.chars) throw new Error("Not connected.");
    await this.chars.settingsRestart.writeValueWithResponse(
      encodeUint32LE(delayMs),
    );
    this.expectingRestart = true;
    if (this.expectingRestartTimer) {
      clearTimeout(this.expectingRestartTimer);
    }
    // Safety net: if no disconnect follows within delay_ms + slack, the
    // write likely didn't trigger a reboot - don't let a later, unrelated
    // drop get mislabeled as "restarting".
    this.expectingRestartTimer = setTimeout(() => {
      this.expectingRestart = false;
      this.expectingRestartTimer = null;
    }, delayMs + RESTART_GRACE_MS);
  }

  async writeWifiConnect(ssid: string, password: string): Promise<void> {
    if (!this.chars) throw new Error("Not connected.");
    await writeJson(this.chars.wifiConnect, { ssid, password });
  }

  async sendWifiCommand(command: WifiCommand): Promise<void> {
    if (!this.chars) throw new Error("Not connected.");
    await this.chars.wifiCommand.writeValueWithResponse(
      new Uint8Array([WIFI_COMMAND_OPCODES[command]]),
    );
  }

  async setLogEnabled(enabled: boolean): Promise<void> {
    if (!this.chars) throw new Error("Not connected.");
    await writeBool(this.chars.logEnabled, enabled);
    this.updateData((data) => ({ ...data, logEnabled: enabled }));
  }

  clearLog(): void {
    this.updateData((data) => ({ ...data, logLines: [] }));
  }

  private setState(next: BleClientState): void {
    this.state = next;
    for (const listener of this.listeners) {
      listener();
    }
  }

  private updateData(updater: (data: ConnectedData) => ConnectedData): void {
    if (this.state.status === "connected") {
      this.setState({ status: "connected", data: updater(this.state.data) });
    } else if (this.state.status === "reconnecting") {
      this.setState({
        status: "reconnecting",
        data: updater(this.state.data),
        reason: this.state.reason,
      });
    }
  }

  private clearRestartExpectation(): void {
    this.expectingRestart = false;
    if (this.expectingRestartTimer) {
      clearTimeout(this.expectingRestartTimer);
      this.expectingRestartTimer = null;
    }
  }

  private async performConnect(isReconnect: boolean): Promise<void> {
    const device = this.device;
    if (!device) return;

    if (!isReconnect) {
      this.setState({ status: "connecting" });
    }

    try {
      const server = await withTimeout(
        device.gatt!.connect(),
        OPERATION_TIMEOUT_MS,
        "Timed out connecting to device.",
      );

      const settingsService = await withTimeout(
        server.getPrimaryService(SETTINGS_SERVICE),
        OPERATION_TIMEOUT_MS,
        "Settings service not found on this device.",
      );
      const wifiService = await withTimeout(
        server.getPrimaryService(WIFI_SERVICE),
        OPERATION_TIMEOUT_MS,
        "WiFi manager service not found on this device.",
      );
      const logService = await withTimeout(
        server.getPrimaryService(LOG_SERVICE),
        OPERATION_TIMEOUT_MS,
        "Log service not found on this device.",
      );
      const systemInfoService = await withTimeout(
        server.getPrimaryService(SYSTEM_INFO_SERVICE),
        OPERATION_TIMEOUT_MS,
        "System info service not found on this device.",
      );

      const chars: CharacteristicRefs = {
        settingsData: await settingsService.getCharacteristic(SETTINGS_DATA_CHR),
        settingsUpdate: await settingsService.getCharacteristic(SETTINGS_UPDATE_CHR),
        settingsRestartRequired: await settingsService.getCharacteristic(SETTINGS_RESTART_REQUIRED_CHR),
        settingsRestart: await settingsService.getCharacteristic(SETTINGS_RESTART_CHR),
        wifiStatus: await wifiService.getCharacteristic(WIFI_STATUS_CHR),
        wifiConnect: await wifiService.getCharacteristic(WIFI_CONNECT_CHR),
        wifiCommand: await wifiService.getCharacteristic(WIFI_COMMAND_CHR),
        wifiStoredCredential: await wifiService.getCharacteristic(WIFI_STORED_CREDENTIAL_CHR),
        logMessage: await logService.getCharacteristic(LOG_MESSAGE_CHR),
        logEnabled: await logService.getCharacteristic(LOG_ENABLED_CHR),
        systemInfoInfo: await systemInfoService.getCharacteristic(SYSTEM_INFO_INFO_CHR),
        systemInfoConfigSchema: await systemInfoService.getCharacteristic(SYSTEM_INFO_CONFIG_SCHEMA_CHR),
      };

      const [
        systemInfo,
        configSchema,
        settingsSnapshot,
        restartRequired,
        wifiStatus,
        wifiStoredCredential,
      ] = await withTimeout(
        Promise.all([
          readJson<SystemInfo>(chars.systemInfoInfo),
          readJson<ConfigSchemaEntry[]>(chars.systemInfoConfigSchema),
          readJson<SettingsSnapshot>(chars.settingsData),
          readBool(chars.settingsRestartRequired),
          readJson<WifiStatus>(chars.wifiStatus),
          readJson<WifiStoredCredential>(chars.wifiStoredCredential),
        ]),
        OPERATION_TIMEOUT_MS,
        "Timed out reading device data.",
      );

      await writeBool(chars.logEnabled, true);

      chars.settingsRestartRequired.addEventListener(
        "characteristicvaluechanged",
        this.handleRestartRequiredChanged,
      );
      chars.wifiStatus.addEventListener(
        "characteristicvaluechanged",
        this.handleWifiStatusChanged,
      );
      chars.logMessage.addEventListener(
        "characteristicvaluechanged",
        this.handleLogMessageChanged,
      );
      await chars.settingsRestartRequired.startNotifications();
      await chars.wifiStatus.startNotifications();
      await chars.logMessage.startNotifications();

      this.chars = chars;

      const data: ConnectedData = {
        systemInfo,
        configSchema,
        settingsSnapshot,
        restartRequired,
        wifiStatus,
        wifiStoredCredential,
        logLines: [],
        logEnabled: true,
      };
      this.setState({ status: "connected", data });
    } catch (error) {
      if (isReconnect) {
        // Swallow - the reconnect loop retries on its next tick. Keep
        // whatever "reconnecting" state is already showing (stale data +
        // reason); don't touch this.device so the same paired
        // BluetoothDevice keeps being retried without a new chooser
        // prompt.
        if (device.gatt?.connected) {
          device.gatt.disconnect();
        }
        return;
      }

      device.removeEventListener(
        "gattserverdisconnected",
        this.handleDisconnected,
      );
      if (device.gatt?.connected) {
        device.gatt.disconnect();
      }
      this.device = null;
      this.chars = null;
      const message =
        error instanceof Error ? error.message : "Failed to connect to device.";
      this.setState({ status: "disconnected", error: message });
    }
  }

  private startReconnectLoop(): void {
    if (this.reconnectTimer) return;
    this.reconnectTimer = setInterval(() => {
      void this.attemptReconnect();
    }, RECONNECT_INTERVAL_MS);
  }

  private stopReconnectLoop(): void {
    if (this.reconnectTimer) {
      clearInterval(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  private async attemptReconnect(): Promise<void> {
    if (this.reconnecting || !this.device) return;
    this.reconnecting = true;
    try {
      await this.performConnect(true);
      if (this.state.status === "connected") {
        this.stopReconnectLoop();
      }
    } finally {
      this.reconnecting = false;
    }
  }

  private readonly handleDisconnected = (): void => {
    this.stopReconnectLoop();
    const previous = this.state;
    const data =
      previous.status === "connected" || previous.status === "reconnecting"
        ? previous.data
        : null;
    if (!data) {
      // Disconnected before we ever reached "connected" - the in-flight
      // performConnect()'s own try/catch will handle this via a failed
      // GATT operation.
      return;
    }

    const reason: DisconnectReason = this.expectingRestart
      ? "restarting"
      : "lost";
    this.clearRestartExpectation();

    this.chars = null;
    this.setState({ status: "reconnecting", data, reason });
    this.startReconnectLoop();
  };

  private readonly handleRestartRequiredChanged = (event: Event): void => {
    const characteristic = event.target as BluetoothRemoteGATTCharacteristic;
    if (!characteristic.value) return;
    const restartRequired = decodeBoolByte(characteristic.value);
    this.updateData((data) => ({ ...data, restartRequired }));
  };

  private readonly handleWifiStatusChanged = (event: Event): void => {
    const characteristic = event.target as BluetoothRemoteGATTCharacteristic;
    if (!characteristic.value) return;
    const wifiStatus = decodeUtf8Json<WifiStatus>(characteristic.value);
    this.updateData((data) => ({ ...data, wifiStatus }));
  };

  private readonly handleLogMessageChanged = (event: Event): void => {
    const characteristic = event.target as BluetoothRemoteGATTCharacteristic;
    if (!characteristic.value) return;
    const line = decodeUtf8Text(characteristic.value);
    if (!line) return;
    this.updateData((data) => {
      const logLines = [...data.logLines, line];
      if (logLines.length > LOG_BUFFER_MAX) {
        logLines.splice(0, logLines.length - LOG_BUFFER_MAX);
      }
      return { ...data, logLines };
    });
  };
}
```

- [ ] **Step 2: Verify it compiles and lints**

```bash
cd /home/dodol/Repositories/mate/mate-things/frontend
PATH=/tmp/mate-node-v24.18.1/bin:$PATH node_modules/.bin/tsc --noEmit
PATH=/tmp/mate-node-v24.18.1/bin:$PATH node_modules/.bin/eslint \
  "src/app/(authenticated)/ble-direct/_lib/ble/BleClient.ts"
```

Expected: no new errors introduced by this file.

- [ ] **Step 3: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-things
git add "frontend/src/app/(authenticated)/ble-direct/_lib/ble/BleClient.ts"
git commit -m "feat(ble-direct): add BleClient — GATT discovery, read/write/notify, reconnect state machine"
```

---

## Task 5: Frontend — hook, connection banner, page shell, nav entry

**Repo:** `mate-things`

**Files:**
- Create: `frontend/src/app/(authenticated)/ble-direct/_lib/useBleConnection.ts`
- Create: `frontend/src/app/(authenticated)/ble-direct/_components/ConnectionBanner.tsx`
- Create: `frontend/src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx`
- Create: `frontend/src/app/(authenticated)/ble-direct/page.tsx`
- Modify: `frontend/src/config/navigation.ts`

**Interfaces:**
- Consumes: `BleClient`/`BleClientState` from Task 4; `useToast` from `@/hooks/use-toast` (already exists in this codebase).
- Produces (consumed by Tasks 6-9): `useBleConnection(): { state: BleClientState; connect: () => void; disconnect: () => void; writeSettingsUpdate: (partial: Record<string, unknown>) => Promise<void>; restartDevice: (delayMs: number) => Promise<void>; writeWifiConnect: (ssid: string, password: string) => Promise<void>; sendWifiCommand: (command: WifiCommand) => Promise<void>; setLogEnabled: (enabled: boolean) => Promise<void>; clearLog: () => void; }`. `BleDirectRoot` is the component Tasks 6-9 will wire their blocks into (a placeholder empty-state milestone is built now; blocks are added in later tasks).

- [ ] **Step 1: Write the hook**

Create `frontend/src/app/(authenticated)/ble-direct/_lib/useBleConnection.ts`. This also owns the one-shot "reconnected" toast from the design spec: it tracks the previous render's state via a ref, and fires a success toast exactly on a `reconnecting` → `connected` transition, worded differently depending on the `reconnecting` state's `reason`.

```ts
"use client";

import { useCallback, useEffect, useRef, useState, useSyncExternalStore } from "react";

import { useToast } from "@/hooks/use-toast";

import { BleClient, type BleClientState } from "./ble/BleClient";
import type { WifiCommand } from "./ble/protocol";

export function useBleConnection(): {
  state: BleClientState;
  connect: () => void;
  disconnect: () => void;
  writeSettingsUpdate: (partial: Record<string, unknown>) => Promise<void>;
  restartDevice: (delayMs: number) => Promise<void>;
  writeWifiConnect: (ssid: string, password: string) => Promise<void>;
  sendWifiCommand: (command: WifiCommand) => Promise<void>;
  setLogEnabled: (enabled: boolean) => Promise<void>;
  clearLog: () => void;
} {
  const [client] = useState(() => new BleClient());
  const toast = useToast();

  useEffect(() => {
    return () => {
      client.disconnect();
    };
  }, [client]);

  const state = useSyncExternalStore(
    useCallback((listener) => client.subscribe(listener), [client]),
    useCallback(() => client.getState(), [client]),
    useCallback(() => client.getState(), [client]),
  );

  const previousStateRef = useRef(state);
  useEffect(() => {
    const previous = previousStateRef.current;
    if (previous.status === "reconnecting" && state.status === "connected") {
      if (previous.reason === "restarting") {
        toast.success("Device restarted", "Device restarted successfully.");
      } else {
        toast.success("Reconnected", "The device connection was restored.");
      }
    }
    previousStateRef.current = state;
  }, [state, toast]);

  return {
    state,
    connect: useCallback(() => {
      void client.connect();
    }, [client]),
    disconnect: useCallback(() => {
      client.disconnect();
    }, [client]),
    writeSettingsUpdate: useCallback(
      (partial: Record<string, unknown>) => client.writeSettingsUpdate(partial),
      [client],
    ),
    restartDevice: useCallback(
      (delayMs: number) => client.restartDevice(delayMs),
      [client],
    ),
    writeWifiConnect: useCallback(
      (ssid: string, password: string) => client.writeWifiConnect(ssid, password),
      [client],
    ),
    sendWifiCommand: useCallback(
      (command: WifiCommand) => client.sendWifiCommand(command),
      [client],
    ),
    setLogEnabled: useCallback(
      (enabled: boolean) => client.setLogEnabled(enabled),
      [client],
    ),
    clearLog: useCallback(() => client.clearLog(), [client]),
  };
}
```

- [ ] **Step 2: Write the connection banner**

Create `frontend/src/app/(authenticated)/ble-direct/_components/ConnectionBanner.tsx`:

```tsx
"use client";

import Button from "@/components/ui/button";

import type { BleClientState } from "../_lib/ble/BleClient";

export default function ConnectionBanner({
  state,
  onConnect,
  onDisconnect,
}: {
  state: BleClientState;
  onConnect: () => void;
  onDisconnect: () => void;
}) {
  if (state.status === "unsupported") {
    return (
      <div className="border-critical bg-critical/10 text-critical rounded-xl border p-4 text-sm">
        Web Bluetooth isn&apos;t available in this browser or context. Use a
        Chromium-based browser (Chrome, Edge) over HTTPS or localhost.
      </div>
    );
  }

  if (state.status === "disconnected") {
    return (
      <div className="border-control-border bg-muted flex flex-wrap items-center justify-between gap-3 rounded-xl border p-4">
        <div>
          <p className="font-semibold">Not connected</p>
          {state.error ? (
            <p className="text-critical mt-1 text-sm">{state.error}</p>
          ) : (
            <p className="text-muted-foreground mt-1 text-sm">
              Connect to a device advertising as matedev_* to get started.
            </p>
          )}
        </div>
        <Button onClick={onConnect}>Connect</Button>
      </div>
    );
  }

  if (state.status === "requesting" || state.status === "connecting") {
    return (
      <div className="border-control-border bg-muted rounded-xl border p-4">
        <p className="font-semibold">
          {state.status === "requesting" ? "Choose a device…" : "Connecting…"}
        </p>
      </div>
    );
  }

  if (state.status === "reconnecting") {
    const label =
      state.reason === "restarting"
        ? "Device is restarting — reconnecting…"
        : "Connection lost — reconnecting…";
    const tone =
      state.reason === "restarting"
        ? "border-control-border bg-muted"
        : "border-warning bg-warning/10";
    return (
      <div
        className={`flex flex-wrap items-center justify-between gap-3 rounded-xl border p-4 ${tone}`}
      >
        <p className="font-semibold">{label}</p>
        <Button variant="secondary" onClick={onDisconnect}>
          Disconnect
        </Button>
      </div>
    );
  }

  return (
    <div className="border-success bg-success/10 flex flex-wrap items-center justify-between gap-3 rounded-xl border p-4">
      <p className="font-semibold">Connected</p>
      <Button variant="secondary" onClick={onDisconnect}>
        Disconnect
      </Button>
    </div>
  );
}
```

- [ ] **Step 3: Write the root component (empty-state milestone — blocks added in later tasks)**

Create `frontend/src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx`:

```tsx
"use client";

import ConnectionBanner from "./ConnectionBanner";
import { useBleConnection } from "../_lib/useBleConnection";

export default function BleDirectRoot() {
  const { state, connect, disconnect } = useBleConnection();

  return (
    <div className="space-y-6">
      <ConnectionBanner state={state} onConnect={connect} onDisconnect={disconnect} />

      {state.status === "disconnected" ? (
        <div className="border-control-border bg-muted rounded-2xl border border-dashed p-8 text-center">
          <p className="font-semibold">No device connected</p>
          <p className="text-muted-foreground mt-1 text-sm">
            Press Connect above and choose a device advertising as
            matedev_* to view its system info, WiFi status, settings, and
            live log.
          </p>
        </div>
      ) : null}
    </div>
  );
}
```

(Tasks 6-9 will add the four blocks and expand this component's connected-state branch. This intermediate version is a deliberate, independently-renderable milestone.)

- [ ] **Step 4: Write the page**

Create `frontend/src/app/(authenticated)/ble-direct/page.tsx`:

```tsx
import type { Metadata } from "next";

import PageHeader from "@/components/ui/page-header";
import { requireSessionContext } from "@/lib/session";

import BleDirectRoot from "./_components/BleDirectRoot";

export const metadata: Metadata = { title: "BLE Direct — Mate Things" };

export default async function BleDirectPage() {
  await requireSessionContext();

  return (
    <main className="mx-auto w-full max-w-5xl space-y-8 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="BLE Direct"
        description="Connect directly to a device over Bluetooth to inspect its system info, manage WiFi, edit settings, and view a live log — no backend involved."
      />
      <BleDirectRoot />
    </main>
  );
}
```

- [ ] **Step 5: Add the nav entry**

In `frontend/src/config/navigation.ts`, in the `"Operations"` group's `items` array, add a new entry after `"Action History"`:

```ts
  {
    label: "Operations",
    items: [
      {
        label: "Actions",
        href: "/actions",
        requiredPermissions: ["action:get"],
      },
      {
        label: "Action History",
        href: "/action-history",
        requiredPermissions: ["action_log:get"],
      },
      {
        label: "BLE Direct",
        href: "/ble-direct",
        requiredPermissions: [],
      },
    ],
  },
```

- [ ] **Step 6: Verify it compiles, lints, and renders**

```bash
cd /home/dodol/Repositories/mate/mate-things/frontend
PATH=/tmp/mate-node-v24.18.1/bin:$PATH node_modules/.bin/tsc --noEmit
PATH=/tmp/mate-node-v24.18.1/bin:$PATH node_modules/.bin/eslint \
  "src/app/(authenticated)/ble-direct/_lib/useBleConnection.ts" \
  "src/app/(authenticated)/ble-direct/_components/ConnectionBanner.tsx" \
  "src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx" \
  "src/app/(authenticated)/ble-direct/page.tsx" \
  src/config/navigation.ts
```

Then boot the dev server and confirm the page renders (this is the "quick test to ensure it compiles and rendered successfully" the requester asked for — log in, navigate to `/ble-direct`, and confirm the nav entry and empty-state both appear):

```bash
cd /home/dodol/Repositories/mate/mate-things/frontend
PATH=/tmp/mate-node-v24.18.1/bin:$PATH node_modules/.bin/next dev -p 3100 &
sleep 5
curl -sf http://localhost:3100/ble-direct -o /dev/null -w "%{http_code}\n"
```

Expected: a 200 or a redirect to `/login` (this page is behind auth, per `requireSessionContext`) — either confirms the route compiles and the server doesn't crash rendering it. If a real login flow is available, actually log in and visually confirm the nav entry ("BLE Direct" under Operations) and the empty-state card render correctly. Stop the dev server afterward (`kill %1` or equivalent).

- [ ] **Step 7: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-things
git add "frontend/src/app/(authenticated)/ble-direct/_lib/useBleConnection.ts" \
        "frontend/src/app/(authenticated)/ble-direct/_components/ConnectionBanner.tsx" \
        "frontend/src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx" \
        "frontend/src/app/(authenticated)/ble-direct/page.tsx" \
        frontend/src/config/navigation.ts
git commit -m "feat(ble-direct): add page shell, connection banner, useBleConnection hook, and nav entry"
```

---

## Task 6: Frontend — System Info block

**Repo:** `mate-things`

**Files:**
- Create: `frontend/src/app/(authenticated)/ble-direct/_components/blocks/SystemInfoBlock.tsx`
- Modify: `frontend/src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx`

**Interfaces:**
- Consumes: `SystemInfo` from Task 3's `protocol.ts`; `Card` from `@/components/ui/card`.
- Produces: `export default function SystemInfoBlock({ info }: { info: SystemInfo })`.

- [ ] **Step 1: Write the block**

Create `frontend/src/app/(authenticated)/ble-direct/_components/blocks/SystemInfoBlock.tsx`:

```tsx
import Card from "@/components/ui/card";

import type { SystemInfo } from "../../_lib/ble/protocol";

function InfoRow({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="flex items-start justify-between gap-4">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="text-right font-medium">{value}</dd>
    </div>
  );
}

export default function SystemInfoBlock({ info }: { info: SystemInfo }) {
  return (
    <Card>
      <h2 className="font-display text-primary text-xl tracking-wide">
        System info
      </h2>
      <div className="mt-4 grid gap-6 sm:grid-cols-2">
        <div>
          <h3 className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
            Project
          </h3>
          <dl className="mt-2 space-y-2">
            <InfoRow label="Name" value={info.project.name} />
            <InfoRow label="Project" value={info.project.project_name} />
            <InfoRow label="Version" value={info.project.project_version} />
            <InfoRow label="Type" value={info.project.type} />
            <InfoRow label="Firmware" value={info.project.firmware_version} />
          </dl>
        </div>
        <div>
          <h3 className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
            Chip
          </h3>
          <dl className="mt-2 space-y-2">
            <InfoRow label="Model" value={info.chip.model} />
            <InfoRow label="Revision" value={info.chip.revision} />
            <InfoRow label="Cores" value={info.chip.cores} />
            <InfoRow label="MAC" value={info.chip.hardware_mac} />
          </dl>
        </div>
      </div>
    </Card>
  );
}
```

- [ ] **Step 2: Wire it into `BleDirectRoot`**

Replace the full contents of `frontend/src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx` with:

```tsx
"use client";

import ConnectionBanner from "./ConnectionBanner";
import SystemInfoBlock from "./blocks/SystemInfoBlock";
import { useBleConnection } from "../_lib/useBleConnection";

export default function BleDirectRoot() {
  const { state, connect, disconnect } = useBleConnection();

  const data =
    state.status === "connected" || state.status === "reconnecting"
      ? state.data
      : null;

  return (
    <div className="space-y-6">
      <ConnectionBanner state={state} onConnect={connect} onDisconnect={disconnect} />

      {data ? (
        <div
          className={
            state.status === "reconnecting" ? "space-y-6 opacity-60" : "space-y-6"
          }
        >
          <SystemInfoBlock info={data.systemInfo} />
        </div>
      ) : state.status === "disconnected" ? (
        <div className="border-control-border bg-muted rounded-2xl border border-dashed p-8 text-center">
          <p className="font-semibold">No device connected</p>
          <p className="text-muted-foreground mt-1 text-sm">
            Press Connect above and choose a device advertising as
            matedev_* to view its system info, WiFi status, settings, and
            live log.
          </p>
        </div>
      ) : null}
    </div>
  );
}
```

(Tasks 7-9 will add `WifiManagerBlock`, `SettingsBlock`, and `LogBlock` into the same connected-state `<div>`, in that order.)

- [ ] **Step 3: Verify**

```bash
cd /home/dodol/Repositories/mate/mate-things/frontend
PATH=/tmp/mate-node-v24.18.1/bin:$PATH node_modules/.bin/tsc --noEmit
PATH=/tmp/mate-node-v24.18.1/bin:$PATH node_modules/.bin/eslint \
  "src/app/(authenticated)/ble-direct/_components/blocks/SystemInfoBlock.tsx" \
  "src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx"
```

- [ ] **Step 4: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-things
git add "frontend/src/app/(authenticated)/ble-direct/_components/blocks/SystemInfoBlock.tsx" \
        "frontend/src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx"
git commit -m "feat(ble-direct): add System Info block"
```

---

## Task 7: Frontend — WiFi Manager block

**Repo:** `mate-things`

**Files:**
- Create: `frontend/src/app/(authenticated)/ble-direct/_components/blocks/WifiManagerBlock.tsx`
- Modify: `frontend/src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx`

**Interfaces:**
- Consumes: `WifiStatus`, `WifiStoredCredential`, `WifiCommand` from Task 3's `protocol.ts`; `writeWifiConnect`/`sendWifiCommand` from Task 5's `useBleConnection`.
- Produces: `export default function WifiManagerBlock({ status, stored, onConnect, onCommand }: { status: WifiStatus; stored: WifiStoredCredential; onConnect: (ssid: string, password: string) => Promise<void>; onCommand: (command: WifiCommand) => Promise<void>; })`.

- [ ] **Step 1: Write the block**

Create `frontend/src/app/(authenticated)/ble-direct/_components/blocks/WifiManagerBlock.tsx`:

```tsx
"use client";

import { useState, type FormEvent } from "react";

import Button from "@/components/ui/button";
import Card from "@/components/ui/card";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import StatusBadge from "@/components/ui/status-badge";

import type { WifiCommand, WifiStatus, WifiStoredCredential } from "../../_lib/ble/protocol";

function InfoRow({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="flex items-start justify-between gap-4">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="text-right font-medium">{value}</dd>
    </div>
  );
}

export default function WifiManagerBlock({
  status,
  stored,
  onConnect,
  onCommand,
}: {
  status: WifiStatus;
  stored: WifiStoredCredential;
  onConnect: (ssid: string, password: string) => Promise<void>;
  onCommand: (command: WifiCommand) => Promise<void>;
}) {
  const [ssid, setSsid] = useState("");
  const [password, setPassword] = useState("");
  const [connectError, setConnectError] = useState<string>();
  const [connectPending, setConnectPending] = useState(false);
  const [commandError, setCommandError] = useState<string>();
  const [pendingCommand, setPendingCommand] = useState<WifiCommand>();

  async function handleConnect(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setConnectError(undefined);
    setConnectPending(true);
    try {
      await onConnect(ssid, password);
      setSsid("");
      setPassword("");
    } catch (error) {
      setConnectError(
        error instanceof Error ? error.message : "Failed to connect.",
      );
    } finally {
      setConnectPending(false);
    }
  }

  async function handleCommand(command: WifiCommand) {
    setCommandError(undefined);
    setPendingCommand(command);
    try {
      await onCommand(command);
    } catch (error) {
      setCommandError(
        error instanceof Error ? error.message : "Command failed.",
      );
    } finally {
      setPendingCommand(undefined);
    }
  }

  return (
    <Card>
      <h2 className="font-display text-primary text-xl tracking-wide">
        WiFi manager
      </h2>

      <div className="mt-4 flex flex-wrap items-center gap-2">
        <StatusBadge variant={status.is_up ? "success" : "neutral"}>
          {status.is_up ? "Up" : "Down"}
        </StatusBadge>
        <StatusBadge
          variant={
            status.sta_connection_status === "connected" ? "success" : "warning"
          }
        >
          {status.sta_connection_status === "connected"
            ? "Connected"
            : "Disconnected"}
        </StatusBadge>
      </div>

      <dl className="mt-4 space-y-2 text-sm">
        <InfoRow label="SSID" value={status.ssid || "—"} />
        <InfoRow label="IP" value={status.ip} />
        <InfoRow label="Netmask" value={status.netmask} />
        <InfoRow label="Gateway" value={status.gateway} />
        <InfoRow label="RSSI" value={`${status.rssi} dBm`} />
        <InfoRow
          label="Try connect on init"
          value={status.try_connect_on_init_enabled ? "Enabled" : "Disabled"}
        />
      </dl>
      <p className="text-muted-foreground mt-1 text-xs">
        Edit &quot;Try connect on init&quot; in the Settings block below.
      </p>

      <div className="mt-4 flex flex-wrap gap-2">
        <Button
          variant="secondary"
          disabled={pendingCommand !== undefined}
          onClick={() => void handleCommand("start")}
        >
          Start
        </Button>
        <Button
          variant="secondary"
          disabled={pendingCommand !== undefined}
          onClick={() => void handleCommand("stop")}
        >
          Stop
        </Button>
        <Button
          variant="secondary"
          disabled={pendingCommand !== undefined || !stored.available}
          onClick={() => void handleCommand("connect_stored")}
        >
          Connect to stored
        </Button>
        <Button
          variant="secondary"
          disabled={pendingCommand !== undefined || !status.is_up}
          onClick={() => void handleCommand("disconnect")}
        >
          Disconnect
        </Button>
        <Button
          variant="critical"
          disabled={pendingCommand !== undefined || !stored.available}
          onClick={() => void handleCommand("forget_stored")}
        >
          Forget stored
        </Button>
      </div>
      {commandError ? (
        <p className="text-critical mt-2 text-sm">{commandError}</p>
      ) : null}

      <form
        onSubmit={handleConnect}
        className="border-control-border mt-5 space-y-3 border-t pt-4"
      >
        <h3 className="text-sm font-semibold">Connect to a network</h3>
        <div>
          <Label htmlFor="wifi-ssid">SSID</Label>
          <Input
            id="wifi-ssid"
            value={ssid}
            onChange={(event) => setSsid(event.target.value)}
            required
          />
        </div>
        <div>
          <Label htmlFor="wifi-password">Password</Label>
          <Input
            id="wifi-password"
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
          />
        </div>
        {connectError ? (
          <p className="text-critical text-sm">{connectError}</p>
        ) : null}
        <Button type="submit" disabled={connectPending}>
          {connectPending ? "Connecting…" : "Connect"}
        </Button>
      </form>
    </Card>
  );
}
```

- [ ] **Step 2: Wire it into `BleDirectRoot`**

In `frontend/src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx`, add the import:

```tsx
import WifiManagerBlock from "./blocks/WifiManagerBlock";
```

Destructure the two new actions from the hook and pass the block:

```tsx
  const { state, connect, disconnect, writeWifiConnect, sendWifiCommand } =
    useBleConnection();
```

```tsx
          <SystemInfoBlock info={data.systemInfo} />
          <WifiManagerBlock
            status={data.wifiStatus}
            stored={data.wifiStoredCredential}
            onConnect={writeWifiConnect}
            onCommand={sendWifiCommand}
          />
```

- [ ] **Step 3: Verify**

```bash
cd /home/dodol/Repositories/mate/mate-things/frontend
PATH=/tmp/mate-node-v24.18.1/bin:$PATH node_modules/.bin/tsc --noEmit
PATH=/tmp/mate-node-v24.18.1/bin:$PATH node_modules/.bin/eslint \
  "src/app/(authenticated)/ble-direct/_components/blocks/WifiManagerBlock.tsx" \
  "src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx"
```

- [ ] **Step 4: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-things
git add "frontend/src/app/(authenticated)/ble-direct/_components/blocks/WifiManagerBlock.tsx" \
        "frontend/src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx"
git commit -m "feat(ble-direct): add WiFi Manager block"
```

---

## Task 8: Frontend — Settings block (the schema-driven core)

**Repo:** `mate-things`

**Files:**
- Create: `frontend/src/app/(authenticated)/ble-direct/_components/blocks/SettingsBlock/SettingsField.tsx`
- Create: `frontend/src/app/(authenticated)/ble-direct/_components/blocks/SettingsBlock/SettingsBlock.tsx`
- Modify: `frontend/src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx`

**Interfaces:**
- Consumes: `ConfigSchemaEntry`, `SettingsSnapshot` from Task 3's `protocol.ts`; `writeSettingsUpdate`/`restartDevice` from Task 5's `useBleConnection`.
- Produces: `export default function SettingsBlock({ schema, snapshot, restartRequired, onSave, onRestart }: { schema: readonly ConfigSchemaEntry[]; snapshot: SettingsSnapshot; restartRequired: boolean; onSave: (partial: Record<string, unknown>) => Promise<void>; onRestart: () => Promise<void>; })`.

- [ ] **Step 1: Write the per-field widget**

Create `frontend/src/app/(authenticated)/ble-direct/_components/blocks/SettingsBlock/SettingsField.tsx`:

```tsx
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";

import type { ConfigSchemaEntry } from "../../../_lib/ble/protocol";

export type SettingsFieldValue = string | boolean;

export default function SettingsField({
  entry,
  value,
  secretSet,
  error,
  onChange,
}: {
  entry: ConfigSchemaEntry;
  value: SettingsFieldValue;
  secretSet: boolean | undefined;
  error: string | undefined;
  onChange: (value: SettingsFieldValue) => void;
}) {
  const fieldId = `ble-settings-${entry.key}`;

  if (entry.type === "bool") {
    return (
      <div className="flex items-center justify-between gap-4 py-2">
        <Label htmlFor={fieldId}>{entry.key}</Label>
        <input
          id={fieldId}
          type="checkbox"
          checked={value === true}
          onChange={(event) => onChange(event.target.checked)}
          className="size-4"
        />
      </div>
    );
  }

  if (entry.type === "uint32") {
    return (
      <div className="py-2">
        <Label htmlFor={fieldId}>{entry.key}</Label>
        <Input
          id={fieldId}
          type="number"
          min={0}
          step={1}
          value={typeof value === "string" ? value : ""}
          onChange={(event) => onChange(event.target.value)}
        />
        {error ? <p className="text-critical mt-1 text-sm">{error}</p> : null}
      </div>
    );
  }

  if (entry.type === "string") {
    const isSecret = secretSet !== undefined;
    return (
      <div className="py-2">
        <Label htmlFor={fieldId}>{entry.key}</Label>
        <Input
          id={fieldId}
          type={isSecret ? "password" : "text"}
          value={typeof value === "string" ? value : ""}
          placeholder={
            isSecret
              ? secretSet
                ? "Already set — leave blank to keep"
                : "Not set"
              : undefined
          }
          onChange={(event) => onChange(event.target.value)}
        />
      </div>
    );
  }

  return (
    <div className="py-2">
      <p className="text-sm font-medium">{entry.key}</p>
      <p className="text-muted-foreground text-sm">
        Unsupported type &quot;{entry.type}&quot; — not editable here.
      </p>
    </div>
  );
}
```

- [ ] **Step 2: Write the block**

Create `frontend/src/app/(authenticated)/ble-direct/_components/blocks/SettingsBlock/SettingsBlock.tsx`:

```tsx
"use client";

import { useMemo, useState } from "react";

import Button from "@/components/ui/button";
import Card from "@/components/ui/card";

import type { ConfigSchemaEntry, SettingsSnapshot } from "../../../_lib/ble/protocol";
import SettingsField, { type SettingsFieldValue } from "./SettingsField";

function initialValues(
  schema: readonly ConfigSchemaEntry[],
  snapshot: SettingsSnapshot,
): Record<string, SettingsFieldValue> {
  const values: Record<string, SettingsFieldValue> = {};
  for (const entry of schema) {
    if (entry.type === "bool") {
      values[entry.key] = snapshot[entry.key] === true;
    } else {
      const raw = snapshot[entry.key];
      values[entry.key] =
        typeof raw === "string" || typeof raw === "number" ? String(raw) : "";
    }
  }
  return values;
}

export default function SettingsBlock({
  schema,
  snapshot,
  restartRequired,
  onSave,
  onRestart,
}: {
  schema: readonly ConfigSchemaEntry[];
  snapshot: SettingsSnapshot;
  restartRequired: boolean;
  onSave: (partial: Record<string, unknown>) => Promise<void>;
  onRestart: () => Promise<void>;
}) {
  const [values, setValues] = useState<Record<string, SettingsFieldValue>>(
    () => initialValues(schema, snapshot),
  );
  const [touched, setTouched] = useState<Record<string, boolean>>({});
  const [saveError, setSaveError] = useState<string>();
  const [savePending, setSavePending] = useState(false);
  const [restartError, setRestartError] = useState<string>();
  const [restartPending, setRestartPending] = useState(false);

  // Reset the form whenever the device gives us a fresh snapshot (after a
  // successful save or a reconnect) - adjusting state during render per
  // https://react.dev/learn/you-might-not-need-an-effect#adjusting-some-state-when-a-prop-changes,
  // not in an effect (eslint's react-hooks/set-state-in-effect rejects
  // synchronous setState-in-effect, and this avoids the extra render pass
  // anyway).
  const [syncedSnapshot, setSyncedSnapshot] = useState(snapshot);
  if (snapshot !== syncedSnapshot) {
    setSyncedSnapshot(snapshot);
    setValues(initialValues(schema, snapshot));
    setTouched({});
  }

  const fieldErrors = useMemo(() => {
    const errors: Record<string, string> = {};
    for (const entry of schema) {
      if (entry.type !== "uint32" || !touched[entry.key]) continue;
      const raw = values[entry.key];
      if (typeof raw !== "string" || raw.trim() === "") continue;
      if (!/^\d+$/.test(raw.trim())) {
        errors[entry.key] = "Enter a whole number.";
      }
    }
    return errors;
  }, [schema, touched, values]);

  const touchedKeys = Object.keys(touched).filter((key) => touched[key]);
  const hasErrors = Object.keys(fieldErrors).length > 0;

  function handleChange(entry: ConfigSchemaEntry, value: SettingsFieldValue) {
    setValues((current) => ({ ...current, [entry.key]: value }));
    setTouched((current) => ({ ...current, [entry.key]: true }));
  }

  async function handleSave() {
    setSaveError(undefined);
    setSavePending(true);
    try {
      const partial: Record<string, unknown> = {};
      for (const entry of schema) {
        if (!touched[entry.key]) continue;
        const raw = values[entry.key];
        if (entry.type === "bool") {
          partial[entry.key] = raw === true;
        } else if (entry.type === "uint32") {
          if (typeof raw === "string" && raw.trim() !== "") {
            partial[entry.key] = Number.parseInt(raw, 10);
          }
        } else if (entry.type === "string") {
          if (typeof raw === "string" && raw !== "") {
            partial[entry.key] = raw;
          }
        }
      }
      await onSave(partial);
    } catch (error) {
      setSaveError(
        error instanceof Error ? error.message : "Failed to save settings.",
      );
    } finally {
      setSavePending(false);
    }
  }

  async function handleRestart() {
    setRestartError(undefined);
    setRestartPending(true);
    try {
      await onRestart();
    } catch (error) {
      setRestartError(
        error instanceof Error ? error.message : "Failed to restart device.",
      );
    } finally {
      setRestartPending(false);
    }
  }

  return (
    <Card>
      <h2 className="font-display text-primary text-xl tracking-wide">
        Settings
      </h2>

      {restartRequired ? (
        <div className="border-warning bg-warning/10 mt-4 flex flex-wrap items-center justify-between gap-3 rounded-xl border p-3">
          <p className="text-sm font-semibold">
            Some changes require a restart to take effect.
          </p>
          <Button
            variant="secondary"
            disabled={restartPending}
            onClick={() => void handleRestart()}
          >
            {restartPending ? "Restarting…" : "Restart now"}
          </Button>
        </div>
      ) : null}
      {restartError ? (
        <p className="text-critical mt-2 text-sm">{restartError}</p>
      ) : null}

      <div className="divide-border mt-4 divide-y">
        {schema.map((entry) => {
          const secretSet =
            entry.type === "string" &&
            !(entry.key in snapshot) &&
            typeof snapshot[`${entry.key}_set`] === "boolean"
              ? (snapshot[`${entry.key}_set`] as boolean)
              : undefined;
          return (
            <SettingsField
              key={entry.key}
              entry={entry}
              value={values[entry.key] ?? (entry.type === "bool" ? false : "")}
              secretSet={secretSet}
              error={fieldErrors[entry.key]}
              onChange={(value) => handleChange(entry, value)}
            />
          );
        })}
      </div>

      {saveError ? (
        <p className="text-critical mt-2 text-sm">{saveError}</p>
      ) : null}
      <Button
        className="mt-4"
        disabled={touchedKeys.length === 0 || hasErrors || savePending}
        onClick={() => void handleSave()}
      >
        {savePending ? "Saving…" : "Save settings"}
      </Button>
    </Card>
  );
}
```

- [ ] **Step 3: Wire it into `BleDirectRoot`**

In `frontend/src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx`, add the import:

```tsx
import SettingsBlock from "./blocks/SettingsBlock/SettingsBlock";
```

Destructure the two new actions:

```tsx
  const {
    state,
    connect,
    disconnect,
    writeWifiConnect,
    sendWifiCommand,
    writeSettingsUpdate,
    restartDevice,
  } = useBleConnection();
```

Add the block after `WifiManagerBlock`:

```tsx
          <SettingsBlock
            schema={data.configSchema}
            snapshot={data.settingsSnapshot}
            restartRequired={data.restartRequired}
            onSave={writeSettingsUpdate}
            onRestart={() => restartDevice(0)}
          />
```

- [ ] **Step 4: Verify**

```bash
cd /home/dodol/Repositories/mate/mate-things/frontend
PATH=/tmp/mate-node-v24.18.1/bin:$PATH node_modules/.bin/tsc --noEmit
PATH=/tmp/mate-node-v24.18.1/bin:$PATH node_modules/.bin/eslint \
  "src/app/(authenticated)/ble-direct/_components/blocks/SettingsBlock/SettingsField.tsx" \
  "src/app/(authenticated)/ble-direct/_components/blocks/SettingsBlock/SettingsBlock.tsx" \
  "src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx"
```

- [ ] **Step 5: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-things
git add "frontend/src/app/(authenticated)/ble-direct/_components/blocks/SettingsBlock/SettingsField.tsx" \
        "frontend/src/app/(authenticated)/ble-direct/_components/blocks/SettingsBlock/SettingsBlock.tsx" \
        "frontend/src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx"
git commit -m "feat(ble-direct): add schema-driven Settings block"
```

---

## Task 9: Frontend — Log block, final wiring, and dev-server render check

**Repo:** `mate-things`

**Files:**
- Create: `frontend/src/app/(authenticated)/ble-direct/_components/blocks/LogBlock.tsx`
- Modify: `frontend/src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx`

**Interfaces:**
- Consumes: `setLogEnabled`/`clearLog` from Task 5's `useBleConnection`.
- Produces: `export default function LogBlock({ lines, enabled, onToggle, onClear }: { lines: readonly string[]; enabled: boolean; onToggle: (enabled: boolean) => Promise<void>; onClear: () => void; })`. This is the last task — `BleDirectRoot` is complete after this.

- [ ] **Step 1: Write the block**

Create `frontend/src/app/(authenticated)/ble-direct/_components/blocks/LogBlock.tsx`:

```tsx
"use client";

import { useEffect, useRef, useState } from "react";

import Button from "@/components/ui/button";
import Card from "@/components/ui/card";

export default function LogBlock({
  lines,
  enabled,
  onToggle,
  onClear,
}: {
  lines: readonly string[];
  enabled: boolean;
  onToggle: (enabled: boolean) => Promise<void>;
  onClear: () => void;
}) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [autoScroll, setAutoScroll] = useState(true);
  const [toggleError, setToggleError] = useState<string>();
  const [togglePending, setTogglePending] = useState(false);

  useEffect(() => {
    if (!autoScroll) return;
    const container = containerRef.current;
    if (!container) return;
    container.scrollTop = container.scrollHeight;
  }, [lines, autoScroll]);

  function handleScroll() {
    const container = containerRef.current;
    if (!container) return;
    const distanceFromBottom =
      container.scrollHeight - container.scrollTop - container.clientHeight;
    setAutoScroll(distanceFromBottom < 24);
  }

  async function handleToggle() {
    setToggleError(undefined);
    setTogglePending(true);
    try {
      await onToggle(!enabled);
    } catch (error) {
      setToggleError(
        error instanceof Error ? error.message : "Failed to update logging.",
      );
    } finally {
      setTogglePending(false);
    }
  }

  return (
    <Card>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="font-display text-primary text-xl tracking-wide">Log</h2>
        <div className="flex items-center gap-2">
          <Button
            variant="secondary"
            disabled={togglePending}
            onClick={() => void handleToggle()}
          >
            {enabled ? "Disable streaming" : "Enable streaming"}
          </Button>
          <Button variant="secondary" onClick={onClear}>
            Clear
          </Button>
        </div>
      </div>
      {toggleError ? (
        <p className="text-critical mt-2 text-sm">{toggleError}</p>
      ) : null}

      <div className="relative mt-4">
        <div
          ref={containerRef}
          onScroll={handleScroll}
          className="border-control-border bg-ink text-surface h-64 overflow-y-auto rounded-xl border p-3 font-mono text-xs leading-5"
        >
          {lines.length === 0 ? (
            <p className="text-surface/60">No log lines yet.</p>
          ) : (
            lines.map((line, index) => <div key={index}>{line}</div>)
          )}
        </div>
        {!autoScroll ? (
          <Button
            variant="secondary"
            className="absolute right-3 bottom-3"
            onClick={() => {
              setAutoScroll(true);
              const container = containerRef.current;
              if (container) container.scrollTop = container.scrollHeight;
            }}
          >
            Jump to latest
          </Button>
        ) : null}
      </div>
    </Card>
  );
}
```

- [ ] **Step 2: Final wiring of `BleDirectRoot`**

Replace the full contents of `frontend/src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx` with:

```tsx
"use client";

import ConnectionBanner from "./ConnectionBanner";
import LogBlock from "./blocks/LogBlock";
import SettingsBlock from "./blocks/SettingsBlock/SettingsBlock";
import SystemInfoBlock from "./blocks/SystemInfoBlock";
import WifiManagerBlock from "./blocks/WifiManagerBlock";
import { useBleConnection } from "../_lib/useBleConnection";

export default function BleDirectRoot() {
  const {
    state,
    connect,
    disconnect,
    writeSettingsUpdate,
    restartDevice,
    writeWifiConnect,
    sendWifiCommand,
    setLogEnabled,
    clearLog,
  } = useBleConnection();

  const data =
    state.status === "connected" || state.status === "reconnecting"
      ? state.data
      : null;

  return (
    <div className="space-y-6">
      <ConnectionBanner state={state} onConnect={connect} onDisconnect={disconnect} />

      {data ? (
        <div
          className={
            state.status === "reconnecting" ? "space-y-6 opacity-60" : "space-y-6"
          }
        >
          <SystemInfoBlock info={data.systemInfo} />
          <WifiManagerBlock
            status={data.wifiStatus}
            stored={data.wifiStoredCredential}
            onConnect={writeWifiConnect}
            onCommand={sendWifiCommand}
          />
          <SettingsBlock
            schema={data.configSchema}
            snapshot={data.settingsSnapshot}
            restartRequired={data.restartRequired}
            onSave={writeSettingsUpdate}
            onRestart={() => restartDevice(0)}
          />
          <LogBlock
            lines={data.logLines}
            enabled={data.logEnabled}
            onToggle={setLogEnabled}
            onClear={clearLog}
          />
        </div>
      ) : state.status === "disconnected" ? (
        <div className="border-control-border bg-muted rounded-2xl border border-dashed p-8 text-center">
          <p className="font-semibold">No device connected</p>
          <p className="text-muted-foreground mt-1 text-sm">
            Press Connect above and choose a device advertising as
            matedev_* to view its system info, WiFi status, settings, and
            live log.
          </p>
        </div>
      ) : null}
    </div>
  );
}
```

- [ ] **Step 3: Full verification pass**

```bash
cd /home/dodol/Repositories/mate/mate-things/frontend
PATH=/tmp/mate-node-v24.18.1/bin:$PATH node_modules/.bin/tsc --noEmit
PATH=/tmp/mate-node-v24.18.1/bin:$PATH node_modules/.bin/eslint "src/app/(authenticated)/ble-direct/**/*.{ts,tsx}" src/types/web-bluetooth.d.ts src/config/navigation.ts
```

Then boot the dev server again and re-confirm the full page renders end to end (nav entry visible, page loads, empty state shows — the connected-state blocks can only be exercised against real hardware, which is the requester's job, not this plan's):

```bash
cd /home/dodol/Repositories/mate/mate-things/frontend
PATH=/tmp/mate-node-v24.18.1/bin:$PATH node_modules/.bin/next dev -p 3100 &
sleep 5
curl -sf http://localhost:3100/ble-direct -o /dev/null -w "%{http_code}\n"
kill %1
```

Expected: no compile/runtime errors, 200 or an auth redirect (not a 500).

- [ ] **Step 4: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-things
git add "frontend/src/app/(authenticated)/ble-direct/_components/blocks/LogBlock.tsx" \
        "frontend/src/app/(authenticated)/ble-direct/_components/BleDirectRoot.tsx"
git commit -m "feat(ble-direct): add Log block, complete page wiring"
```
