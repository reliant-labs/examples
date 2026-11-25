import kcl_plugin.forge as plugin
# Per-environment app config VALUES (user-owned — edit here).
# The typed AppConfig schema + projection are generated from
# proto/notes/config/v1/config.proto; this instance supplies the per-env values
# they project. Only fields this env pins are set — every other field
# inherits its AppConfig schema default.

import config_gen

app_config: config_gen.AppConfig = {
    auto_migrate = True
    environment = "development"
}

# ── Frontend runtime config (PUBLIC — served to the browser) ──────────
# Projected into each frontend's config.js. These values reach the browser
# at RUNTIME, not baked into the bundle, so one built bundle can be
# promoted between environments.
import frontend_config_gen

web_config: frontend_config_gen.WebConfig = {
    api_url = "http://localhost:" + str(plugin.resolve_port("notes-dev-api", 8085))
}
