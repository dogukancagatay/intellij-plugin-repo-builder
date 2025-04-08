# Intellij IDEA Custom Plugin Repository Builder

This tool builds a **Custom Plugin Repository** for JetBrains IDEs like IntelliJ IDEA. It supports:

- Automatic download of public plugins from JetBrains Plugin Repository
- Hosting via built-in HTTP server
- Support for **local/private plugins** with rich metadata (name, vendor, description)
- Generates `updatePlugins.xml` in [JetBrains plugin repository format](https://plugins.jetbrains.com/docs/intellij/update-plugins-format.html)

## 🔧 Quickstart

### 1. Build the repository

This downloads remote plugins (latest versions) and copies local plugin files, then generates `updatePlugins.xml`.

```sh
./repo-builder -build
```

### 2. Serve the repository

Launch an HTTP server to serve `updatePlugins.xml` and plugin files.

```sh
./repo-builder -serve
```

---

## ⚙️ Configuration (`config.yaml`)

Your configuration should look like this:

```yaml
serverUrl: http://localhost:3000
bindIp: 0.0.0.0
port: "3000"
dir: out

# Remote JetBrains plugins (by numeric plugin ID)
plugins:
  - "164"       # IdeaVim
  - "10080"     # .env files support

# Local plugin entries
localPlugins:
  - id: com.local
    version: 2.2.0
    since: "211.1.*"
    until: "999.*"
    file: ./plugins.zip
    name: plugin name
    vendor: abc
    vendorEmail: abc@abc.com
    vendorUrl: https://www.google.com
    description: describe
```

### 🔍 Field Descriptions

| Field         | Type     | Description |
|---------------|----------|-------------|
| `serverUrl`   | string   | Base URL to be used in `updatePlugins.xml` (e.g., `http://localhost:3000`) |
| `bindIp`      | string   | IP to bind the HTTP server (default: `0.0.0.0`) |
| `port`        | string   | Port for HTTP server (default: `3000`) |
| `dir`         | string   | Output directory for files and XML |
| `plugins`     | list     | List of public JetBrains plugin numeric IDs |
| `localPlugins`| list     | List of local plugins with metadata and file path |

---

## 🧩 IntelliJ IDE Setup

1. Open **Settings → Plugins**
2. Click the ⚙️ icon → **Manage Plugin Repositories**
3. Add your custom repo URL:

```
http://localhost:3000/updatePlugins.xml
```

You’ll now see your custom/private plugins in the **Marketplace** tab.

---

## 🛠 Advanced Features

- ✅ Full support for `<name>`, `<vendor>`, and `<description>` fields (including HTML)
- ✅ `<![CDATA[ ... ]]>` block for plugin descriptions
- ✅ Support for JetBrains build constraints via `<idea-version since-build="..." until-build="..."/>`
- ✅ Offline usage

---

## 📌 Notes

- Remote plugin IDs can be found in plugin URLs, e.g.:  
  `https://plugins.jetbrains.com/plugin/164-ideavim` → ID is `164`
- Remote plugins fetch **latest release** automatically
- Local plugins must specify a valid `.zip` or `.jar` plugin file

---

## 🚧 TODO

- [ ] Add support for multiple JetBrains IDE versions in output
- [ ] Auto-sync from private plugin git repo or artifact repo (e.g., Nexus/Artifactory)
- [ ] Web UI for local plugin upload

---

## 📚 References

- [JetBrains Plugin Repository Format](https://plugins.jetbrains.com/docs/intellij/update-plugins-format.html)
- [Custom Plugin Repositories in IntelliJ](https://www.jetbrains.com/help/idea/managing-plugins.html#repos)

