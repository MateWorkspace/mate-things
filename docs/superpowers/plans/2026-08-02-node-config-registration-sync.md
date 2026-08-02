# Node Config Registration Sync Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** the ESP32 firmware reports its current preloaded config values as
part of every MQTT registration message; the backend upserts
`node_config_values` from that report on every registration (not just
first-time device creation), keeping BLE-set and DB-stored config from
drifting apart — with no schema migration and no default-value concept
anywhere.

**Architecture:** One signature change cascades through `mate-espidf-base`'s
publish stack (domain contract → infra JSON builder → application layer
that gathers the actual values). One new field cascades through
`mate-things`'s registration DTO/domain type into the existing
`Register` usecase, which gains a config-sync step using a validator
extracted out of the existing `config_value` usecase so both share it.

**Tech Stack:** ESP-IDF v6.0.2 / NimBLE / cJSON (firmware), Go / paho.mqtt
(backend).

## Global Constraints

- **No schema migration, no default-value concept anywhere** — this is an
  explicit, repeated scope cut from the requester. Do not add a
  `default_value` column, a preloaded-defaults header move, or any
  registration-time row-seeding step.
- Every config value crossing any wire in this system is a **string**,
  regardless of its real type (`node_config_values.value` is `TEXT`; the
  existing MQTT config-set payload's `value` is a string) — the new
  registration `config` payload follows the same convention.
- `Register` syncs config on **every** registration (every reconnect), not
  gated on the `created` bool `UpsertRegistration` returns.
- An unrecognized config key (not in the node's current firmware schema)
  or an invalid value for its type is logged and **skipped**, never a hard
  failure of the whole registration — matches the firmware's own MQTT
  config handler's "unknown key" philosophy.
- No automated test suite is being planned or added (explicit scope
  decision). Verification is: does it build, and does it run. Existing
  tests whose call sites move during a refactor must still pass — fix
  those call sites, that's maintenance, not new test-writing.

---

## Task 1 (firmware): extend the registration publish signature to carry config

**Repo:** `mate-espidf-base`

**Files:**
- Modify: `main/include/domain/models/preloaded.h`
- Modify: `main/include/domain/contracts/messaging/def_pub.h`
- Modify: `main/include/infrastructure/messaging/def_pub/mqtt_impl_utils.h`
- Modify: `main/src/infrastructure/messaging/def_pub/mqtt_impl_utils.c`
- Modify: `main/src/infrastructure/messaging/def_pub/mqtt_impl.c`
- Modify: `main/src/infrastructure/messaging/def_pub/stub_impl.c`

**Interfaces:**
- Produces (consumed by Task 2): `dom_models_preloaded_kv_t { const char* key; const char* value; }` (in `domain/models/preloaded.h`); `dom_contracts_messaging_def_pub_t.registration(self, device_id, device_info, firmware_name, config, config_count)` — two new trailing params, `const dom_models_preloaded_kv_t* config` and `size_t config_count`.

- [ ] **Step 1: Add the key/value pair type**

In `main/include/domain/models/preloaded.h`, insert right after the
`dom_models_preloaded_schema_entry_t` struct (before the
`DOMAIN_MODELS_PRELOADED_SCHEMA_COUNT_X` macro):

```c
typedef struct {
    const char* key;
    const char* value;
} dom_models_preloaded_kv_t;

```

(i.e. the file's `typedef struct { const char* key; ... } dom_models_preloaded_schema_entry_t;` block is followed by this new block, then the existing `#define DOMAIN_MODELS_PRELOADED_SCHEMA_COUNT_X ...` continues unchanged.)

- [ ] **Step 2: Extend the publish contract**

In `main/include/domain/contracts/messaging/def_pub.h`, add the include and
extend `registration`:

```c
#include "domain/models/device_status.h"
#include "domain/models/error.h"
```
→
```c
#include "domain/models/device_status.h"
#include "domain/models/error.h"
#include "domain/models/preloaded.h"
```

```c
    dom_models_error_t (*registration)(
        dom_contracts_messaging_def_pub_t* self,
        const char*                        device_id,
        const char*                        device_info,
        const char*                        firmware_name
    );
```
→
```c
    dom_models_error_t (*registration)(
        dom_contracts_messaging_def_pub_t* self,
        const char*                        device_id,
        const char*                        device_info,
        const char*                        firmware_name,
        const dom_models_preloaded_kv_t*   config,
        size_t                             config_count
    );
```

- [ ] **Step 3: Extend the JSON builder's declaration**

In `main/include/infrastructure/messaging/def_pub/mqtt_impl_utils.h`, add
the include and extend the declaration:

```c
#include "domain/models/device_status.h"
#include "domain/models/error.h"
#include "infrastructure/messaging/def_pub/mqtt_impl_types.h"
```
→
```c
#include "domain/models/device_status.h"
#include "domain/models/error.h"
#include "domain/models/preloaded.h"
#include "infrastructure/messaging/def_pub/mqtt_impl_types.h"
```

```c
char* inf_messaging_def_pub_mqtt_impl_build_registration_json(
    const char* device_id,
    const char* device_info,
    const char* firmware_name
);
```
→
```c
char* inf_messaging_def_pub_mqtt_impl_build_registration_json(
    const char*                       device_id,
    const char*                       device_info,
    const char*                       firmware_name,
    const dom_models_preloaded_kv_t*  config,
    size_t                            config_count
);
```

- [ ] **Step 4: Extend the JSON builder's implementation**

In `main/src/infrastructure/messaging/def_pub/mqtt_impl_utils.c`, replace
the whole `inf_messaging_def_pub_mqtt_impl_build_registration_json`
function:

```c
char* inf_messaging_def_pub_mqtt_impl_build_registration_json(
    const char* device_id,
    const char* device_info,
    const char* firmware_name
) {
    if (!cstr_available(device_id) || !cstr_available(device_info) || !cstr_available(firmware_name)) {
        return NULL;
    }

    cJSON* root = cJSON_CreateObject();
    if (!root) {
        return NULL;
    }

    if (!cJSON_AddStringToObject(root, "device_id", device_id) ||
        !cJSON_AddStringToObject(root, "device_info", device_info) ||
        !cJSON_AddStringToObject(root, "firmware_name", firmware_name)) {
        cJSON_Delete(root);
        return NULL;
    }

    char* json = cJSON_PrintUnformatted(root);
    cJSON_Delete(root);

    return json;
}
```

with:

```c
char* inf_messaging_def_pub_mqtt_impl_build_registration_json(
    const char*                      device_id,
    const char*                      device_info,
    const char*                      firmware_name,
    const dom_models_preloaded_kv_t* config,
    size_t                           config_count
) {
    if (!cstr_available(device_id) || !cstr_available(device_info) || !cstr_available(firmware_name)) {
        return NULL;
    }

    cJSON* root = cJSON_CreateObject();
    if (!root) {
        return NULL;
    }

    if (!cJSON_AddStringToObject(root, "device_id", device_id) ||
        !cJSON_AddStringToObject(root, "device_info", device_info) ||
        !cJSON_AddStringToObject(root, "firmware_name", firmware_name)) {
        cJSON_Delete(root);
        return NULL;
    }

    /* Always emit a (possibly empty) "config" object - every value is a
       string, matching how node_config_values.value and the config-set
       MQTT payload already represent config values on the backend, so the
       registration handler there doesn't need a second parsing convention. */
    cJSON* config_obj = cJSON_AddObjectToObject(root, "config");
    if (!config_obj) {
        cJSON_Delete(root);
        return NULL;
    }
    for (size_t i = 0; i < config_count; i++) {
        if (!config || !config[i].key || !config[i].value) {
            continue;
        }
        if (!cJSON_AddStringToObject(config_obj, config[i].key, config[i].value)) {
            cJSON_Delete(root);
            return NULL;
        }
    }

    char* json = cJSON_PrintUnformatted(root);
    cJSON_Delete(root);

    return json;
}
```

- [ ] **Step 5: Thread it through `mqtt_impl.c`**

In `main/src/infrastructure/messaging/def_pub/mqtt_impl.c`, update the
prototype:

```c
static dom_models_error_t registration_impl(
    dom_contracts_messaging_def_pub_t* self,
    const char*                        device_id,
    const char*                        device_info,
    const char*                        firmware_name
);
```
→
```c
static dom_models_error_t registration_impl(
    dom_contracts_messaging_def_pub_t* self,
    const char*                        device_id,
    const char*                        device_info,
    const char*                        firmware_name,
    const dom_models_preloaded_kv_t*   config,
    size_t                             config_count
);
```

And the implementation:

```c
static dom_models_error_t registration_impl(
    dom_contracts_messaging_def_pub_t* self,
    const char*                        device_id,
    const char*                        device_info,
    const char*                        firmware_name
) {
    if (!self || !self->ctx || !device_id || device_id[0] == '\0' || !device_info || device_info[0] == '\0' || !firmware_name || firmware_name[0] == '\0') {
        return DOMAIN_MODELS_ERROR_BAD_ARGUMENT;
    }

    inf_messaging_def_pub_mqtt_impl_ctx_t* ctx = self->ctx;

    return inf_messaging_def_pub_mqtt_impl_publish_json(
        ctx,
        "/pub/registration",
        inf_messaging_def_pub_mqtt_impl_build_registration_json(device_id, device_info, firmware_name),
        INF_MESSAGING_DEF_PUB_MQTT_IMPL_QOS_DEFAULT,
        false
    );
}
```
→
```c
static dom_models_error_t registration_impl(
    dom_contracts_messaging_def_pub_t* self,
    const char*                        device_id,
    const char*                        device_info,
    const char*                        firmware_name,
    const dom_models_preloaded_kv_t*   config,
    size_t                             config_count
) {
    if (!self || !self->ctx || !device_id || device_id[0] == '\0' || !device_info || device_info[0] == '\0' || !firmware_name || firmware_name[0] == '\0') {
        return DOMAIN_MODELS_ERROR_BAD_ARGUMENT;
    }

    inf_messaging_def_pub_mqtt_impl_ctx_t* ctx = self->ctx;

    return inf_messaging_def_pub_mqtt_impl_publish_json(
        ctx,
        "/pub/registration",
        inf_messaging_def_pub_mqtt_impl_build_registration_json(device_id, device_info, firmware_name, config, config_count),
        INF_MESSAGING_DEF_PUB_MQTT_IMPL_QOS_DEFAULT,
        false
    );
}
```

- [ ] **Step 6: Update the test-seed stub's signature**

In `main/src/infrastructure/messaging/def_pub/stub_impl.c`, update the
prototype:

```c
static dom_models_error_t registration_impl(
    dom_contracts_messaging_def_pub_t* self,
    const char*                        device_id,
    const char*                        device_info,
    const char*                        firmware_name
);
```
→
```c
static dom_models_error_t registration_impl(
    dom_contracts_messaging_def_pub_t* self,
    const char*                        device_id,
    const char*                        device_info,
    const char*                        firmware_name,
    const dom_models_preloaded_kv_t*   config,
    size_t                             config_count
);
```

And the implementation (the stub only ever records
device_id/device_info/firmware_name for test-seed inspection today —
nothing currently reads a recorded config back out, so `config`/
`config_count` are accepted for signature compatibility and otherwise
unused here):

```c
static dom_models_error_t registration_impl(
    dom_contracts_messaging_def_pub_t* self,
    const char*                        device_id,
    const char*                        device_info,
    const char*                        firmware_name
) {
    if (!self || !self->ctx) {
        return DOMAIN_MODELS_ERROR_BAD_ARGUMENT;
    }

    return inf_messaging_def_pub_stub_impl_set_registration(self->ctx, device_id, device_info, firmware_name);
}
```
→
```c
static dom_models_error_t registration_impl(
    dom_contracts_messaging_def_pub_t* self,
    const char*                        device_id,
    const char*                        device_info,
    const char*                        firmware_name,
    const dom_models_preloaded_kv_t*   config,
    size_t                             config_count
) {
    (void)config;
    (void)config_count;

    if (!self || !self->ctx) {
        return DOMAIN_MODELS_ERROR_BAD_ARGUMENT;
    }

    return inf_messaging_def_pub_stub_impl_set_registration(self->ctx, device_id, device_info, firmware_name);
}
```

- [ ] **Step 7: Verify**

```bash
source /home/dodol/.espressif/v6.0.2/esp-idf/export.sh
cd /home/dodol/Repositories/mate/mate-espidf-base
idf.py reconfigure && idf.py build
```

If the compiler toolchain isn't actually installed in the working
environment (confirmed missing earlier this session — the Python/export
scaffolding is present but `xtensa-esp-elf`/`riscv32-esp-elf` etc. report
"no installed versions"), verify manually instead: re-read every file
touched in this task end-to-end, confirm the new struct/signature is
identical across all 5 call/declaration sites, and grep the whole `main/`
tree for `dom_models_preloaded_kv_t` and `registration(` to confirm every
site was updated consistently. State plainly whether a real build ran.

- [ ] **Step 8: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-espidf-base
git add main/include/domain/models/preloaded.h \
        main/include/domain/contracts/messaging/def_pub.h \
        main/include/infrastructure/messaging/def_pub/mqtt_impl_utils.h \
        main/src/infrastructure/messaging/def_pub/mqtt_impl_utils.c \
        main/src/infrastructure/messaging/def_pub/mqtt_impl.c \
        main/src/infrastructure/messaging/def_pub/stub_impl.c
git commit -m "feat(messaging): extend registration publish to carry a config key/value list"
```

---

## Task 2 (firmware): send the device's current config at registration

**Repo:** `mate-espidf-base`

**Files:**
- Modify: `main/src/application/internal/messaging_callbacks/impl.c`

**Interfaces:**
- Consumes: `dom_models_preloaded_kv_t` and the extended `registration(...)` signature from Task 1; `ctx->cfg.preloaded_repository`'s existing getters (`get_mqtt_proto`, `get_mqtt_host`, `get_mqtt_port`, `get_mqtt_user`, `get_mqtt_pass`, `get_system_restart_after_ms`, `get_wifi_sta_try_connect_on_init` — all already used by `application/internal/settings/impl_utils.c`'s `load_snapshot`, same buffer-size conventions apply: `mqtt_proto[16]`, `mqtt_host[128]`, `mqtt_port[8]`, `mqtt_user[64]`, `mqtt_pass[128]`).

- [ ] **Step 1: Add the includes needed for the key macros**

In `main/src/application/internal/messaging_callbacks/impl.c`, add:

```c
#include "application/internal/messaging_callbacks/impl_types.h"
#include "application/internal/messaging_callbacks/impl_utils.h"
#include "domain/models/device_status.h"
#include "domain/models/error.h"
#include "domain/models/system.h"
#include "domain/usecases/internal/messaging_callbacks.h"
```
→
```c
#include "application/internal/messaging_callbacks/impl_types.h"
#include "application/internal/messaging_callbacks/impl_utils.h"
#include "domain/models/device_status.h"
#include "domain/models/error.h"
#include "domain/models/preloaded.h"
#include "domain/models/system.h"
#include "domain/usecases/internal/messaging_callbacks.h"
```

- [ ] **Step 2: Build the config array and pass it to `registration(...)`**

Replace the whole `publish_registration_impl` function:

```c
static dom_models_error_t publish_registration_impl(
    dom_usecases_internal_messaging_callbacks_t* self
) {
    const char* tag = BASE_TAG "/publish_registration";

    app_internal_messaging_callbacks_impl_ctx_t* ctx = NULL;
    dom_models_error_t                           err = get_ctx(self, &ctx);
    if (err != DOMAIN_MODELS_ERROR_OK) {
        return err;
    }

    char device_id_str[37];
    err = load_device_id_str(ctx, device_id_str, sizeof(device_id_str));
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to load device id: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    dom_models_system_project_info_t project_info;
    err = ctx->cfg.system_info->get_project_info(ctx->cfg.system_info, &project_info);
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to load project info: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    dom_models_system_chip_info_t chip_info;
    err = ctx->cfg.system_info->get_chip_info(ctx->cfg.system_info, &chip_info);
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to load chip info: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    char firmware_name[96];
    err = app_internal_messaging_callbacks_impl_build_firmware_name(&project_info, firmware_name, sizeof(firmware_name));
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to build firmware name: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    char device_info[96];
    err = app_internal_messaging_callbacks_impl_build_device_info(&chip_info, device_info, sizeof(device_info));
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to build device info: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    err = ctx->cfg.def_pub->registration(ctx->cfg.def_pub, device_id_str, device_info, firmware_name);
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to publish registration: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    ctx->cfg.logger->info(ctx->cfg.logger, tag, "Registration published successfully");

    return DOMAIN_MODELS_ERROR_OK;
}
```

with:

```c
static dom_models_error_t publish_registration_impl(
    dom_usecases_internal_messaging_callbacks_t* self
) {
    const char* tag = BASE_TAG "/publish_registration";

    app_internal_messaging_callbacks_impl_ctx_t* ctx = NULL;
    dom_models_error_t                           err = get_ctx(self, &ctx);
    if (err != DOMAIN_MODELS_ERROR_OK) {
        return err;
    }

    char device_id_str[37];
    err = load_device_id_str(ctx, device_id_str, sizeof(device_id_str));
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to load device id: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    dom_models_system_project_info_t project_info;
    err = ctx->cfg.system_info->get_project_info(ctx->cfg.system_info, &project_info);
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to load project info: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    dom_models_system_chip_info_t chip_info;
    err = ctx->cfg.system_info->get_chip_info(ctx->cfg.system_info, &chip_info);
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to load chip info: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    char firmware_name[96];
    err = app_internal_messaging_callbacks_impl_build_firmware_name(&project_info, firmware_name, sizeof(firmware_name));
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to build firmware name: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    char device_info[96];
    err = app_internal_messaging_callbacks_impl_build_device_info(&chip_info, device_info, sizeof(device_info));
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to build device info: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    char mqtt_proto[16];
    err = ctx->cfg.preloaded_repository->get_mqtt_proto(ctx->cfg.preloaded_repository, mqtt_proto, sizeof(mqtt_proto));
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to load mqtt_proto: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    char mqtt_host[128];
    err = ctx->cfg.preloaded_repository->get_mqtt_host(ctx->cfg.preloaded_repository, mqtt_host, sizeof(mqtt_host));
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to load mqtt_host: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    char mqtt_port[8];
    err = ctx->cfg.preloaded_repository->get_mqtt_port(ctx->cfg.preloaded_repository, mqtt_port, sizeof(mqtt_port));
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to load mqtt_port: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    char mqtt_user[64];
    err = ctx->cfg.preloaded_repository->get_mqtt_user(ctx->cfg.preloaded_repository, mqtt_user, sizeof(mqtt_user));
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to load mqtt_user: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    char mqtt_pass[128];
    err = ctx->cfg.preloaded_repository->get_mqtt_pass(ctx->cfg.preloaded_repository, mqtt_pass, sizeof(mqtt_pass));
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to load mqtt_pass: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    uint32_t system_restart_after_ms = 0;
    err                              = ctx->cfg.preloaded_repository->get_system_restart_after_ms(ctx->cfg.preloaded_repository, &system_restart_after_ms);
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to load system_restart_after_ms: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }
    char system_restart_after_ms_str[11];
    int  written = snprintf(system_restart_after_ms_str, sizeof(system_restart_after_ms_str), "%lu", (unsigned long)system_restart_after_ms);
    if (written <= 0 || (size_t)written >= sizeof(system_restart_after_ms_str)) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to format system_restart_after_ms: %s (%d)", dom_models_error_str(DOMAIN_MODELS_ERROR_FAILURE), (int)DOMAIN_MODELS_ERROR_FAILURE);
        return DOMAIN_MODELS_ERROR_FAILURE;
    }

    bool wifi_try_init = false;
    err                = ctx->cfg.preloaded_repository->get_wifi_sta_try_connect_on_init(ctx->cfg.preloaded_repository, &wifi_try_init);
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to load wifi_try_init: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    const dom_models_preloaded_kv_t config[] = {
        {DOMAIN_MODELS_PRELOADED_MQTT_PROTO_KEY, mqtt_proto},
        {DOMAIN_MODELS_PRELOADED_MQTT_HOST_KEY, mqtt_host},
        {DOMAIN_MODELS_PRELOADED_MQTT_PORT_KEY, mqtt_port},
        {DOMAIN_MODELS_PRELOADED_MQTT_USER_KEY, mqtt_user},
        {DOMAIN_MODELS_PRELOADED_MQTT_PASS_KEY, mqtt_pass},
        {DOMAIN_MODELS_PRELOADED_SYSTEM_RESTART_AFTER_MS_KEY, system_restart_after_ms_str},
        {DOMAIN_MODELS_PRELOADED_WIFI_STA_TRY_CONNECT_ON_INIT_KEY, wifi_try_init ? "true" : "false"},
    };

    err = ctx->cfg.def_pub->registration(ctx->cfg.def_pub, device_id_str, device_info, firmware_name, config, sizeof(config) / sizeof(config[0]));
    if (err != DOMAIN_MODELS_ERROR_OK) {
        ctx->cfg.logger->error(ctx->cfg.logger, tag, "Failed to publish registration: %s (%d)", dom_models_error_str(err), (int)err);
        return err;
    }

    ctx->cfg.logger->info(ctx->cfg.logger, tag, "Registration published successfully");

    return DOMAIN_MODELS_ERROR_OK;
}
```

- [ ] **Step 3: Verify**

Same as Task 1 Step 7 — attempt `idf.py reconfigure && idf.py build`;
fall back to manual re-read + grep if the toolchain isn't installed, and
state plainly which one happened. Specifically re-check: every new stack
buffer size matches `application/internal/settings/impl_utils.c`'s
`load_snapshot` exactly (`mqtt_proto[16]`, `mqtt_host[128]`,
`mqtt_port[8]`, `mqtt_user[64]`, `mqtt_pass[128]`), and the `config[]`
array's 7 entries match the 7 keys in `DOMAIN_MODELS_PRELOADED_SCHEMA`
one-for-one.

- [ ] **Step 4: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-espidf-base
git add main/src/application/internal/messaging_callbacks/impl.c
git commit -m "feat(messaging): send current preloaded config values at registration"
```

---

## Task 3 (backend): extract the shared config-value validator

**Repo:** `mate-things` (backend)

**Files:**
- Modify: `internal/application/shared/validation.go`
- Modify: `internal/application/node/config_value/usecase.go`
- Modify: `internal/application/node/config_value/usecase_test.go`

**Interfaces:**
- Produces (consumed by Task 5): `applicationshared.ValidateConfigValue(value string, valueType string) error` (package `applicationshared`, `github.com/MateWorkspace/mate-things/backend/internal/application/shared`).

- [ ] **Step 1: Move the validator into `application/shared/validation.go`**

Add at the end of `internal/application/shared/validation.go`:

```go
// ValidateConfigValue checks a config value string against the value_type
// declared for its key in the owning firmware's config schema (string,
// uint32, or bool) - shared by any usecase that accepts a
// (key, value, value_type) config write, whether that write comes from an
// HTTP request or a device's own MQTT registration report.
func ValidateConfigValue(value string, valueType string) error {
	switch valueType {
	case "uint32":
		if _, err := strconv.ParseUint(value, 10, 32); err != nil {
			return domainmodels.NewError("config value must be a valid uint32", domainmodels.ErrTypeValidation, err)
		}
	case "bool":
		if value != "true" && value != "false" {
			return domainmodels.NewError("config value must be \"true\" or \"false\"", domainmodels.ErrTypeValidation, nil)
		}
	case "string":
		// any string value is acceptable
	default:
		return domainmodels.NewError("unrecognized config value_type", domainmodels.ErrTypeFailure, nil)
	}

	return nil
}
```

`strconv` is already imported in this file (used by `validateNamePattern`) —
no new import needed.

- [ ] **Step 2: Remove the private copy and update the call site**

In `internal/application/node/config_value/usecase.go`, remove the whole
private function:

```go
func validateConfigValue(value string, valueType string) error {
	switch valueType {
	case "uint32":
		if _, err := strconv.ParseUint(value, 10, 32); err != nil {
			return domainmodels.NewError("config value must be a valid uint32", domainmodels.ErrTypeValidation, err)
		}
	case "bool":
		if value != "true" && value != "false" {
			return domainmodels.NewError("config value must be \"true\" or \"false\"", domainmodels.ErrTypeValidation, nil)
		}
	case "string":
		// any string value is acceptable
	default:
		return domainmodels.NewError("unrecognized config value_type", domainmodels.ErrTypeFailure, nil)
	}

	return nil
}
```

Update its one call site:

```go
	if err := validateConfigValue(request.Value, valueType); err != nil {
		return err
	}
```
→
```go
	if err := applicationshared.ValidateConfigValue(request.Value, valueType); err != nil {
		return err
	}
```

Update the import block — remove `"strconv"` (no longer used in this file
once the function moves) and add the shared package:

```go
import (
	"context"
	"strconv"

	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsnode "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/node"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
)
```
→
```go
import (
	"context"

	applicationshared "github.com/MateWorkspace/mate-things/backend/internal/application/shared"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsnode "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/node"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
)
```

- [ ] **Step 3: Keep the existing table test passing**

In `internal/application/node/config_value/usecase_test.go`, update the
call site and import so the existing test still compiles and passes
against the moved function:

```go
import (
	"errors"
	"testing"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)
```
→
```go
import (
	"errors"
	"testing"

	applicationshared "github.com/MateWorkspace/mate-things/backend/internal/application/shared"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)
```

```go
			err := validateConfigValue(test.value, test.valueType)

			if test.wantError && !errors.Is(err, domainmodels.ErrTypeValidation) {
				t.Fatalf("validateConfigValue() error = %v, want validation error", err)
			}
			if !test.wantError && err != nil {
				t.Fatalf("validateConfigValue() error = %v, want nil", err)
			}
```
→
```go
			err := applicationshared.ValidateConfigValue(test.value, test.valueType)

			if test.wantError && !errors.Is(err, domainmodels.ErrTypeValidation) {
				t.Fatalf("ValidateConfigValue() error = %v, want validation error", err)
			}
			if !test.wantError && err != nil {
				t.Fatalf("ValidateConfigValue() error = %v, want nil", err)
			}
```

- [ ] **Step 4: Verify**

```bash
cd /home/dodol/Repositories/mate/mate-things/backend
go build ./...
go vet ./...
go test ./internal/application/shared/... ./internal/application/node/config_value/...
```

Expected: clean build, clean vet, tests pass (the existing
`TestValidateConfigValuePreservesStringDomainAndRejectsInvalidTypedValues`
now exercises `applicationshared.ValidateConfigValue`).

- [ ] **Step 5: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-things
git add backend/internal/application/shared/validation.go \
        backend/internal/application/node/config_value/usecase.go \
        backend/internal/application/node/config_value/usecase_test.go
git commit -m "refactor(config): extract ValidateConfigValue into application/shared"
```

---

## Task 4 (backend): registration DTO/domain type gains `Config`

**Repo:** `mate-things` (backend)

**Files:**
- Modify: `internal/presentation/mqtt/dto/registration.go`
- Modify: `internal/domain/usecases/node/messaging_callback.go`

**Interfaces:**
- Produces (consumed by Task 5): `domainusecasesnode.RegisterNodeMessageRequest.Config map[string]string`; `presentationmqttdto.DecodeRegistration` passes it through.

- [ ] **Step 1: Add `Config` to the domain request type**

In `internal/domain/usecases/node/messaging_callback.go`:

```go
type RegisterNodeMessageRequest struct {
	DeviceId     string
	DeviceInfo   string
	FirmwareName string
}
```
→
```go
type RegisterNodeMessageRequest struct {
	DeviceId     string
	DeviceInfo   string
	FirmwareName string
	Config       map[string]string
}
```

- [ ] **Step 2: Add `Config` to the wire DTO and pass it through**

In `internal/presentation/mqtt/dto/registration.go`:

```go
type Registration struct {
	DeviceId     string `json:"device_id"`
	DeviceInfo   string `json:"device_info"`
	FirmwareName string `json:"firmware_name"`
}
```
→
```go
type Registration struct {
	DeviceId     string            `json:"device_id"`
	DeviceInfo   string            `json:"device_info"`
	FirmwareName string            `json:"firmware_name"`
	Config       map[string]string `json:"config"`
}
```

```go
	return domainusecasesnode.RegisterNodeMessageRequest{
		DeviceId:     dto.DeviceId,
		DeviceInfo:   dto.DeviceInfo,
		FirmwareName: dto.FirmwareName,
	}, nil
```
→
```go
	return domainusecasesnode.RegisterNodeMessageRequest{
		DeviceId:     dto.DeviceId,
		DeviceInfo:   dto.DeviceInfo,
		FirmwareName: dto.FirmwareName,
		Config:       dto.Config,
	}, nil
```

A registration payload from firmware that predates this change (no
`config` key at all) unmarshals `dto.Config` as a nil map — Task 5's sync
step must treat a nil/empty map as "nothing to sync" rather than an error.

- [ ] **Step 3: Verify**

```bash
cd /home/dodol/Repositories/mate/mate-things/backend
go build ./...
go vet ./...
```

- [ ] **Step 4: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-things
git add backend/internal/domain/usecases/node/messaging_callback.go \
        backend/internal/presentation/mqtt/dto/registration.go
git commit -m "feat(registration): add config field to the registration message"
```

---

## Task 5 (backend): sync `node_config_values` on every registration

**Repo:** `mate-things` (backend)

**Files:**
- Modify: `internal/application/node/messaging_callback/usecase.go`
- Modify: `internal/composition/main/application.go`

**Interfaces:**
- Consumes: `applicationshared.ValidateConfigValue` (Task 3); `RegisterNodeMessageRequest.Config` (Task 4); existing `domaincontractsrepository.FirmwareConfigParameter.ReadByFirmwareId` and `domaincontractsrepository.NodeConfigValue.Upsert` (both already defined, already used by the `config_value` usecase — no repository changes needed here).

- [ ] **Step 1: Add the two new dependencies to the usecase struct and constructor**

In `internal/application/node/messaging_callback/usecase.go`:

```go
type usecase struct {
	node          domainusecasesrepocache.Node
	actionLog     domaincontractsrepository.ActionLog
	nodeLog       domaincontractsrepository.NodeLog
	publisher     domaincontractsnode.Publish
	subscriptions domaincontractsnode.Subscriptions
	logger        domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	node domainusecasesrepocache.Node,
	actionLog domaincontractsrepository.ActionLog,
	nodeLog domaincontractsrepository.NodeLog,
	publisher domaincontractsnode.Publish,
	subscriptions domaincontractsnode.Subscriptions,
	logger domaincontractslogger.Leveled,
) domainusecasesnode.MessagingCallback {
	return &usecase{
		node:          node,
		actionLog:     actionLog,
		nodeLog:       nodeLog,
		publisher:     publisher,
		subscriptions: subscriptions,
		logger:        logger,
	}
}
```
→
```go
type usecase struct {
	node                 domainusecasesrepocache.Node
	actionLog            domaincontractsrepository.ActionLog
	nodeLog              domaincontractsrepository.NodeLog
	firmwareConfigParams domaincontractsrepository.FirmwareConfigParameter
	nodeConfigValues     domaincontractsrepository.NodeConfigValue
	publisher            domaincontractsnode.Publish
	subscriptions        domaincontractsnode.Subscriptions
	logger               domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	node domainusecasesrepocache.Node,
	actionLog domaincontractsrepository.ActionLog,
	nodeLog domaincontractsrepository.NodeLog,
	firmwareConfigParams domaincontractsrepository.FirmwareConfigParameter,
	nodeConfigValues domaincontractsrepository.NodeConfigValue,
	publisher domaincontractsnode.Publish,
	subscriptions domaincontractsnode.Subscriptions,
	logger domaincontractslogger.Leveled,
) domainusecasesnode.MessagingCallback {
	return &usecase{
		node:                 node,
		actionLog:            actionLog,
		nodeLog:              nodeLog,
		firmwareConfigParams: firmwareConfigParams,
		nodeConfigValues:     nodeConfigValues,
		publisher:            publisher,
		subscriptions:        subscriptions,
		logger:               logger,
	}
}
```

- [ ] **Step 2: Add the config-sync step to `Register`, and its helper**

Replace:

```go
func (u *usecase) Register(ctx context.Context, request domainusecasesnode.RegisterNodeMessageRequest) error {
	const tag = "node/messaging_callback/Register"

	deviceId, err := applicationshared.RequiredDeviceId(request.DeviceId, "device_id")
	if err != nil {
		u.logger.Warn(ctx, tag, "invalid device_id in registration", domainmodels.LoggerMeta{
			"err":       err,
			"device_id": request.DeviceId,
		})
		return err
	}

	node, created, err := u.node.UpsertRegistration(ctx, deviceId, request.DeviceInfo, request.FirmwareName)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to upsert node registration", domainmodels.LoggerMeta{
			"err":           err,
			"device_id":     request.DeviceId,
			"firmware_name": request.FirmwareName,
		})
		return err
	}

	if created {
		if err := u.subscribeNode(ctx, node.DeviceId); err != nil {
			u.logger.Error(ctx, tag, "failed to subscribe node topics", domainmodels.LoggerMeta{
				"err":       err,
				"device_id": node.DeviceId,
				"node_id":   node.Id,
			})
			return err
		}
	}

	if err := u.publisher.RegistrationAck(ctx, node.DeviceId); err != nil {
		u.logger.Error(ctx, tag, "failed to publish registration ack", domainmodels.LoggerMeta{
			"err":       err,
			"device_id": node.DeviceId,
			"node_id":   node.Id,
		})
		return err
	}

	return nil
}
```

with:

```go
func (u *usecase) Register(ctx context.Context, request domainusecasesnode.RegisterNodeMessageRequest) error {
	const tag = "node/messaging_callback/Register"

	deviceId, err := applicationshared.RequiredDeviceId(request.DeviceId, "device_id")
	if err != nil {
		u.logger.Warn(ctx, tag, "invalid device_id in registration", domainmodels.LoggerMeta{
			"err":       err,
			"device_id": request.DeviceId,
		})
		return err
	}

	node, created, err := u.node.UpsertRegistration(ctx, deviceId, request.DeviceInfo, request.FirmwareName)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to upsert node registration", domainmodels.LoggerMeta{
			"err":           err,
			"device_id":     request.DeviceId,
			"firmware_name": request.FirmwareName,
		})
		return err
	}

	if err := u.syncReportedConfig(ctx, node, request.Config); err != nil {
		u.logger.Error(ctx, tag, "failed to sync reported node config", domainmodels.LoggerMeta{
			"err":         err,
			"device_id":   node.DeviceId,
			"node_id":     node.Id,
			"firmware_id": node.FirmwareId,
		})
		return err
	}

	if created {
		if err := u.subscribeNode(ctx, node.DeviceId); err != nil {
			u.logger.Error(ctx, tag, "failed to subscribe node topics", domainmodels.LoggerMeta{
				"err":       err,
				"device_id": node.DeviceId,
				"node_id":   node.Id,
			})
			return err
		}
	}

	if err := u.publisher.RegistrationAck(ctx, node.DeviceId); err != nil {
		u.logger.Error(ctx, tag, "failed to publish registration ack", domainmodels.LoggerMeta{
			"err":       err,
			"device_id": node.DeviceId,
			"node_id":   node.Id,
		})
		return err
	}

	return nil
}

// syncReportedConfig upserts node_config_values from what the device itself
// reported at registration - the device is the source of truth for its own
// live config, so this runs on every registration (not just first-time
// creation), keeping the DB from drifting away from a value set out-of-band
// (e.g. via BLE) since the node's last reconnect. A key the current
// firmware's schema doesn't recognize, or a value that fails the schema's
// declared value_type, is logged and skipped - never a reason to fail the
// whole registration, matching the firmware's own MQTT config handler's
// "unknown key" behavior.
func (u *usecase) syncReportedConfig(ctx context.Context, node *domainmodels.Node, reported map[string]string) error {
	const tag = "node/messaging_callback/syncReportedConfig"

	if len(reported) == 0 {
		return nil
	}

	params, err := u.firmwareConfigParams.ReadByFirmwareId(ctx, node.FirmwareId)
	if err != nil {
		return err
	}

	valueTypes := make(map[string]string, len(params))
	for _, param := range params {
		valueTypes[param.Key] = param.ValueType
	}

	for key, value := range reported {
		valueType, known := valueTypes[key]
		if !known {
			u.logger.Warn(ctx, tag, "registration reported a config key not in the node's current firmware schema", domainmodels.LoggerMeta{
				"device_id":   node.DeviceId,
				"node_id":     node.Id,
				"firmware_id": node.FirmwareId,
				"key":         key,
			})
			continue
		}

		if err := applicationshared.ValidateConfigValue(value, valueType); err != nil {
			u.logger.Warn(ctx, tag, "registration reported an invalid config value for its type, skipping", domainmodels.LoggerMeta{
				"err":        err,
				"device_id":  node.DeviceId,
				"node_id":    node.Id,
				"key":        key,
				"value_type": valueType,
			})
			continue
		}

		if err := u.nodeConfigValues.Upsert(ctx, node.Id, node.FirmwareId, key, value, nil); err != nil {
			return err
		}
	}

	return nil
}
```

- [ ] **Step 3: Wire the two new dependencies in composition**

In `internal/composition/main/application.go`:

```go
	nodeMessagingCallback := applicationnodemessagingcallback.NewUsecaseImpl(
		nodeRepoCache,
		l.infra.actionLogRepository,
		l.infra.nodeLogRepository,
		l.infra.nodePublisher,
		l.infra.nodeSubscriptions,
		l.infra.logger,
	)
```
→
```go
	nodeMessagingCallback := applicationnodemessagingcallback.NewUsecaseImpl(
		nodeRepoCache,
		l.infra.actionLogRepository,
		l.infra.nodeLogRepository,
		l.infra.firmwareConfigParameterRepository,
		l.infra.nodeConfigValueRepository,
		l.infra.nodePublisher,
		l.infra.nodeSubscriptions,
		l.infra.logger,
	)
```

(`l.infra.firmwareConfigParameterRepository` and
`l.infra.nodeConfigValueRepository` are both already constructed earlier
in this same file for other usecases — this just passes the existing
instances to one more consumer, no new repository construction.)

- [ ] **Step 4: Verify**

```bash
cd /home/dodol/Repositories/mate/mate-things/backend
go build ./...
go vet ./...
go test ./...
```

Expected: clean build, clean vet, full test suite passes.

- [ ] **Step 5: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-things
git add backend/internal/application/node/messaging_callback/usecase.go \
        backend/internal/composition/main/application.go
git commit -m "feat(registration): sync node_config_values from the device's reported config on every registration"
```

---

## Task 6: full verification — builds, and a live end-to-end check

**Repos:** both

**Files:** none (verification only).

- [ ] **Step 1: Firmware build**

```bash
source /home/dodol/.espressif/v6.0.2/esp-idf/export.sh
cd /home/dodol/Repositories/mate/mate-espidf-base
idf.py reconfigure && idf.py build
```

If the toolchain genuinely isn't installed in this environment, this step
was already covered by Tasks 1-2's manual-verification fallback — don't
re-litigate it here, just note the same limitation applies.

- [ ] **Step 2: Backend build and test**

```bash
cd /home/dodol/Repositories/mate/mate-things/backend
go build ./...
go vet ./...
go test ./...
```

- [ ] **Step 3: Live check — rebuild and run the backend stack**

```bash
cd /home/dodol/Repositories/mate/mate-things
docker compose build --no-cache app
docker compose up -d
```

(`--no-cache` because this environment's `docker compose build` has
repeatedly shown stale-cache false positives for source changes earlier
this session.)

- [ ] **Step 4: Confirm a real registration carries `config` and lands in `node_config_values`**

However this environment can exercise a registration (a real flashed
device, or a scripted MQTT publish matching the new payload shape against
the running broker/backend), confirm:

- The registration payload actually published now includes a `config`
  object with all 7 preloaded keys as strings.
- `node_config_values` in Postgres has rows for that node's `firmware_id`
  matching the reported values (query directly:
  `SELECT key, value FROM node_config_values WHERE node_id = '<id>' AND deleted_at IS NULL ORDER BY key;`).
- Re-registering the same device again (a second connect) doesn't error
  and leaves the same rows in place (a no-op upsert of identical values).

- [ ] **Step 5: Confirm the existing set→push path still works**

Using the node config UI (`/nodes/{id}`'s "Device configuration" section)
or a direct `PUT`/`PATCH` to the node config endpoint, set a value for a
key the node's firmware declares, and confirm:

- The HTTP call succeeds and `node_config_values` reflects the new value.
- The device receives the pushed value over `/sub/{device_id}/config` and
  applies it (e.g. re-read the relevant BLE settings characteristic or
  check device logs, whichever is observable in this environment) — this
  is the "re-ensure the configuration menu on node MQTT is properly
  wired" check from the request; only make further code changes here if
  this step actually surfaces a bug (none is expected — this path was
  traced end-to-end during design and found intact).
