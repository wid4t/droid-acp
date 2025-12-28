# Droid ACP

**Unofficial Droid ACP Connector for Zed Editor**
*Size: under 5 MB*

<p align="center">
  <img src="image/example.png" width="1000">
</p>

---

## Supported Models

* Claude Opus 4.5
* Claude Sonnet 4.5
* Claude Haiku 4.5
* GPT-5.1 Codex
* GLM 4.7
* Gemini 3 Flash
* Gemini 3 Pro

> Other models are **to be confirmed**.

---

## Build

1. Open this repository in a terminal.
2. Run:

   ```bash
   go build .
   ```
3. This command produces the `droid-acp` binary
   (Windows: `droid-acp.exe`) in the current working directory.

---

## Add to Environment Variables

1. Choose a directory for the binary (example: `C:\Tools\droid-acp`).
2. Move the compiled binary into that directory.
3. Add the directory to your system `PATH` via **System Environment Variables**.
4. Verify by opening a new terminal and running:

   ```bash
   droid-acp -v
   ```

---

## Configure Zed

Ensure Zed knows how to invoke this agent.

Edit the `settings.json` file located at:

```
C:\Users\{user}\AppData\Roaming\Zed\settings.json
```

Add the following configuration:

```json
"agent_servers": {
  "Droid": {
    "type": "custom",
    "command": "droid-acp"
  }
}
```

---

### Show Only Custom Models

*(Supported in v1.0.5 and later)*

To display **only custom models** in Zed:

```json
"agent_servers": {
  "Droid": {
    "type": "custom",
    "command": "droid-acp",
    "args": ["--model=custom"]
  }
}
```

Possible values for `--model`:

```
custom | common | all
```

Once the binary is available in `PATH`, Zed can invoke `droid-acp` directly via the CLI.

---

## Changelog

### v1.0.5

* **feat:** Add option to display only custom models

### v1.0.4

* **feat:** Implement Droid autonomy level in the AI mode menu
* **note:** Click the prompt message, then press **Ctrl + I**

### v1.0.3

* **feat:** Implement edit tracking (code review)
* **fix:** Fix error when creating responses from Droid using Gemini 3 Pro

### v1.0.2

* **refactor:** Code cleanup and general improvements
* **feat:** Support file edit confirmation (users can review changes before applying)
* **note:**
  To enable file edit confirmation:

  1. Open **Start → PowerShell**
  2. Run `droid`
  3. Press **Ctrl + T**
  4. Ensure it shows:
     `Auto (Off) – all actions require approval`
  5. Start a new session for `droid-acp`

### v1.0.1

* **feat:** Allow switching models directly from Zed Editor

### v1.0.0

* **feat:** Initial project scaffold
* Zed Editor always uses the default Droid ACP model

---

## Status

This project is **under active development** 🚧
