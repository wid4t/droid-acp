Droid ACP ( Droid Unofficial ACP connector for Zed Editor | size : under 5 MB )
=========

<p align="center">
  <img src="image/example.png" width="1000">
</p>

support : 
- claude opus 4.5
- claude sonnet 4.5
- claude haiku 4.5
- gpt 5.1 codex
- glm 4.7 
- gemini 3 flash
- gemini 3 pro

- other model to be confirm

Build
-----
1. Open this repository in a terminal and run `go build .`.
2. The command produces the `droid-acp` binary (Windows: `droid-acp.exe`) in the current working directory.

Add to Environment Variables
----------------------------
1. Choose a directory for the binary (example: `C:\Tools\droid-acp`).
2. Move the build output into that directory.
3. Add the directory to `PATH` via **System Environment Variables** so any CLI can find `droid-acp`.
4. Verify by opening a new terminal and running `droid-acp -v`

Configure Zed
-------------
Make sure Zed knows how to call this agent. Add the following snippet to `settings.json` (see `C:\Users\{user}\AppData\Roaming\Zed\settings.json`):

```json
  "agent_servers": {
    "Droid": {
      "type": "custom",
      "command": "droid-acp"
    }
  }
```

Once the binary is on PATH, Zed can invoke `droid-acp` directly through the CLI.

Change Log
------
### v1.0.4
- feat: implement droid autonomy level in ai mode menu
- note: click on prompt message, and then push ctrl + I on your keyboard

### v1.0.3
- feat: implement edit tracking feature ( code review )
- fix: error get response create from droid when use gemini 3 pro 

### v1.0.2
- refactor code and improvement
- feat: support file edit confirmation ( user can review code before apply this changes )
- note: for support file edit confirmation, start -> powersheel -> run droid -> ctrl + T and make sure show like this
Auto (Off) - all actions require approval. After that create new session for droid-acp

### v1.0.1
- feat: allow switching models from within Zed Editor

### v1.0.0
- feat: initial project scaffold; Zed Editor always uses the default Droid ACP model

Status
------
The project is still under active development
