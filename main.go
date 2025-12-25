// Author: widat
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"droid-acp/types"
	"droid-acp/utils"

	"github.com/google/uuid"
)

const version = "1.0.2"

var (
	writeMu              sync.Mutex
	droidMsgID           int
	acpMsgID             int
	modelId              string
	currentSession       string
	lastSessionCwd       string
	droidIn              io.Writer
	acpOut               io.Writer
	pendingPromptID      any
	pendingPromptMu      sync.Mutex
	pendingSessionID     any
	pendingSessionMu     sync.Mutex
	permissionRequestMu  sync.Mutex
	permissionRequestMap = make(map[string]permissionRequest)
)

type permissionRequest struct {
	DroidRequestID string
	ToolCallID     string
}

func sendACPResponse(id any, result any) error {
	writeMu.Lock()
	defer writeMu.Unlock()

	resp := types.ACPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "[->ZED] %s\n", string(b))
	_, err = fmt.Fprintln(acpOut, string(b))
	return err
}

func sendACPNotification(method string, params any) error {
	writeMu.Lock()
	defer writeMu.Unlock()

	notif := types.ACPNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	b, err := json.Marshal(notif)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to marshal notification: %v\n", err)
		return err
	}
	fmt.Fprintf(os.Stderr, "[->ZED NOTIF] %s\n", string(b))
	_, err = fmt.Fprintln(acpOut, string(b))
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to write notification: %v\n", err)
	}
	return err
}

func sendACPRequest(method string, params any) (string, error) {
	writeMu.Lock()
	defer writeMu.Unlock()

	acpMsgID++
	id := strconv.Itoa(acpMsgID)

	req := struct {
		JSONRPC string `json:"jsonrpc"`
		ID      string `json:"id"`
		Method  string `json:"method"`
		Params  any    `json:"params,omitempty"`
	}{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	b, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	fmt.Fprintf(os.Stderr, "[->ZED REQ] %s\n", string(b))
	_, err = fmt.Fprintln(acpOut, string(b))
	return id, err
}

func sendDroidResponseWithID(id any, result any) error {
	writeMu.Lock()
	defer writeMu.Unlock()

	resp := map[string]any{
		"jsonrpc":           "2.0",
		"factoryApiVersion": "1.0.0",
		"type":              "response",
		"id":                id,
		"result":            result,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(droidIn, string(b))
	return err
}

func sendDroidOK(id any) error {
	switch v := id.(type) {
	case nil:
		return nil
	case string:
		if v == "" {
			return nil
		}
	}
	return sendDroidResponseWithID(id, map[string]bool{"ok": true})
}

func initializeDroidSession(cwd string) error {
	writeMu.Lock()
	defer writeMu.Unlock()

	req := map[string]any{
		"jsonrpc":           "2.0",
		"factoryApiVersion": "1.0.0",
		"type":              "request",
		"id":                "0",
		"method":            "droid.initialize_session",
		"params": map[string]any{
			"machineId": uuid.New().String(),
			"cwd":       cwd,
		},
	}

	b, err := json.Marshal(req)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "[->DROID] %s\n", string(b))
	_, err = fmt.Fprintln(droidIn, string(b))
	return err
}

func sendDroidUserMessage(params map[string]any) error {
	writeMu.Lock()
	defer writeMu.Unlock()

	droidMsgID++
	id := strconv.Itoa(droidMsgID)

	req := map[string]any{
		"jsonrpc":           "2.0",
		"factoryApiVersion": "1.0.0",
		"type":              "request",
		"id":                id,
		"method":            "droid.add_user_message",
		"params":            params,
	}

	b, err := json.Marshal(req)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "[->DROID] %s\n", string(b))
	_, err = fmt.Fprintln(droidIn, string(b))
	return err
}

func sendDroidRequest(method string, params any) (string, error) {
	writeMu.Lock()
	defer writeMu.Unlock()

	droidMsgID++
	id := strconv.Itoa(droidMsgID)

	req := map[string]any{
		"jsonrpc":           "2.0",
		"factoryApiVersion": "1.0.0",
		"type":              "request",
		"id":                id,
		"method":            method,
		"params":            params,
	}

	b, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	fmt.Fprintf(os.Stderr, "[->DROID] %s\n", string(b))
	_, err = fmt.Fprintln(droidIn, string(b))
	return id, err
}

func handleACPRequest(req types.ACPRequest) {
	if req.Method == "" {
		if req.Error != nil {
			fmt.Fprintf(os.Stderr, "[ACP ERROR] id=%v code=%d message=%s\n", req.ID, req.Error.Code, req.Error.Message)
			return
		}
		if len(req.Result) > 0 {
			var permissionResp struct {
				Outcome struct {
					Outcome  string `json:"outcome"`
					OptionId string `json:"optionId"`
				} `json:"outcome"`
			}
			if err := json.Unmarshal(req.Result, &permissionResp); err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse ACP result: %v\n", err)
				return
			}
			if permissionResp.Outcome.OptionId != "" {
				reqID := fmt.Sprint(req.ID)
				permissionRequestMu.Lock()
				request := permissionRequestMap[reqID]
				delete(permissionRequestMap, reqID)
				permissionRequestMu.Unlock()

				responseID := request.DroidRequestID
				if responseID == "" {
					responseID = request.ToolCallID
				}
				if responseID == "" {
					fmt.Fprintf(os.Stderr, "[WARN] Missing droid request id for permission response (acp id=%s)\n", reqID)
					return
				}

				result := map[string]any{
					"selectedOption": permissionResp.Outcome.OptionId,
				}
				if err := sendDroidResponseWithID(responseID, result); err != nil {
					fmt.Fprintf(os.Stderr, "[ERROR] Failed to send permission result to droid: %v\n", err)
				}
				return
			}
		}
		return
	}

	switch req.Method {
	case "initialize":
		result := types.InitializeResult{
			ProtocolVersion: 1,
			AgentCapabilities: types.AgentCapabilities{
				LoadSession: false,
				PromptCapabilities: types.PromptCapabilities{
					Image:           false,
					Audio:           false,
					EmbeddedContext: true,
				},
				MCP: types.McpInfo{
					Http: false,
					Sse:  false,
				},
			},
			AgentInfo: types.AgentInfo{
				Name:    "droid-acp",
				Title:   "Droid ACP",
				Version: version,
			},
		}
		sendACPResponse(req.ID, result)

	case "session/new":
		var params types.NewSessionParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse session/new params: %v\n", err)
			return
		}

		pendingSessionMu.Lock()
		if pendingSessionID != nil {
			fmt.Fprintf(os.Stderr, "[WARN] Overwriting pending session/new ID\n")
		}
		pendingSessionID = req.ID
		pendingSessionMu.Unlock()

		cwd := params.Cwd
		if cwd == "" {
			cwd = "."
		}
		lastSessionCwd = cwd

		if err := initializeDroidSession(cwd); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize droid session: %v\n", err)
		}

	case "session/prompt":
		var params types.PromptParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse session/prompt params: %v\n", err)
			return
		}

		pendingPromptMu.Lock()
		if pendingPromptID != nil {
			fmt.Fprintf(os.Stderr, "[WARN] Overwriting pending prompt ID\n")
		}
		pendingPromptID = req.ID
		pendingPromptMu.Unlock()

		data := make(map[string]any)
		for _, block := range params.Prompt {
			fmt.Fprintf(os.Stderr, "BLOCK TYPE: %s", block.Type)
			switch block.Type {
			case "text":
				data["text"] = block.Text
			case "resource":
				fileName, err := utils.GetFilenameFromUri(block.ClientResources.Uri)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to get filename : %v\n", err)
				}
				data["text"] = block.ClientResources.Text
				data["attachments"] = []map[string]any{
					{
						"name":     fileName,
						"mimeType": block.ClientResources.MimeType,
						"path":     block.ClientResources.Uri,
					},
				}
			}
		}

		if err := sendDroidUserMessage(data); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to send message to droid: %v\n", err)
		}
	case "session/set_model":
		var params types.SetModelParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse session/set_model params: %v\n", err)
			sendACPResponse(req.ID, map[string]any{})
			return
		}

		sessionID := strings.TrimSpace(params.SessionID)
		if sessionID == "" {
			sessionID = currentSession
		}

		modelId = strings.TrimSpace(string(params.ModelID))
		if modelId == "" {
			fmt.Fprintf(os.Stderr, "[WARN] Missing modelId in session/set_model params\n")
		}

		updateParams := map[string]any{
			"sessionId": sessionID,
			"modelId":   modelId,
		}

		if _, err := sendDroidRequest("droid.update_session_settings", updateParams); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to send model update to droid:  %v\n", err)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown ACP method: %s\n", req.Method)
		sendACPResponse(req.ID, map[string]any{})
	}
}

func handleDroidMessage(msg types.DroidMessage) {
	if msg.Method == "" {
		if msg.Type == "response" {
			pendingSessionMu.Lock()
			hasPendingSession := pendingSessionID != nil
			pendingSessionMu.Unlock()
			if !hasPendingSession {
				return
			}

			var result types.ResultModel
			if err := json.Unmarshal(msg.Result, &result); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to parse response from droid: %v\n", err)
				return
			}

			modelIDs := make(map[types.ModelId]struct{})
			var models []types.ModelInfo
			currentModelID := types.ModelId(strings.TrimSpace(result.Settings.ModelID))

			for _, model := range result.AvailableModels {

				id := strings.TrimSpace(model.ModelID)
				if id == "" {
					id = strings.TrimSpace(model.ID)
				}
				if id == "" {
					continue
				}

				name := strings.TrimSpace(model.DisplayName)
				if name == "" {
					name = strings.TrimSpace(model.ShortDisplayName)
				}
				if name == "" {
					name = id
				}

				desc := strings.TrimSpace(model.DisplayName)
				if desc == "" {
					desc = name
				}

				info := types.ModelInfo{
					ModelId: types.ModelId(id),
					Name:    name,
				}
				if desc != "" && desc != name {
					info.Description = desc
				}

				models = append(models, info)
				modelIDs[info.ModelId] = struct{}{}

			}

			if len(models) == 0 {
				fallbackID := currentModelID
				if fallbackID == "" {
					fallbackID = types.ModelId("droid-default")
				}
				name := string(fallbackID)
				if name == "" {
					name = "Default"
				}
				models = append(models, types.ModelInfo{
					ModelId: fallbackID,
					Name:    name,
				})
				modelIDs[fallbackID] = struct{}{}
				if currentModelID == "" {
					currentModelID = fallbackID
				}
			}

			if currentModelID == "" && len(models) > 0 {
				currentModelID = models[0].ModelId
			}

			if _, ok := modelIDs[currentModelID]; !ok && currentModelID != "" {
				name := string(currentModelID)
				if name == "" {
					name = models[0].Name
				}
				models = append([]types.ModelInfo{{
					ModelId: currentModelID,
					Name:    name,
				}}, models...)
			}

			var modelsRoot = types.Models{
				AvailableModels: models,
				CurrentModelId:  currentModelID,
			}

			currentSession = uuid.New().String()
			response := types.NewSessionResult{
				SessionID: currentSession,
				Models:    modelsRoot,
			}
			pendingSessionMu.Lock()
			acpID := pendingSessionID
			pendingSessionID = nil
			pendingSessionMu.Unlock()
			if acpID == nil {
				fmt.Fprintf(os.Stderr, "[WARN] Missing pending session/new ID; cannot respond\n")
				return
			}
			sendACPResponse(acpID, response)
		}
	} else {
		switch msg.Method {
		case "droid.initialize_session":
			sendDroidOK(msg.ID)

		case "droid.update_session_settings":
			sendDroidOK(msg.ID)

		case "droid.session_notification":
			var params types.DroidNotification
			if err := json.Unmarshal(msg.Params, &params); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to parse droid.session_notification: %v\n", err)
				return
			}

			paramsJSON, _ := json.Marshal(msg.Params)
			fmt.Fprintf(os.Stderr, "[DROID PARAMS] %s\n", string(paramsJSON))

			fmt.Fprintf(os.Stderr, "[DROID NOTIF] type=%s textDelta=%q\n", params.Notification.Type, params.Notification.TextDelta)

			switch params.Notification.Type {
			case "assistant_text_delta":
				content := &types.Content{
					Type: "text",
					Text: params.Notification.TextDelta,
				}
				update := types.SessionUpdateParam{
					SessionID: currentSession,
					Update: types.Update{
						SessionUpdate: "agent_message_chunk",
						Content:       content,
					},
				}

				debugJSON, _ := json.Marshal(update)
				fmt.Fprintf(os.Stderr, "[DEBUG] Sending session/update with params: %s\n", string(debugJSON))

				if err := sendACPNotification("session/update", update); err != nil {
					fmt.Fprintf(os.Stderr, "[ERROR] Failed to send session/update notification: %v\n", err)
				}

			case "thinking_text_delta":
				content := &types.Content{
					Type: "text",
					Text: params.Notification.TextDelta,
				}
				update := types.SessionUpdateParam{
					SessionID: currentSession,
					Update: types.Update{
						SessionUpdate: "agent_thought_chunk",
						Content:       content,
					},
				}

				debugJSON, _ := json.Marshal(update)
				fmt.Fprintf(os.Stderr, "[DEBUG] Sending session/update with params: %s\n", string(debugJSON))

				if err := sendACPNotification("session/update", update); err != nil {
					fmt.Fprintf(os.Stderr, "[ERROR] Failed to send session/update notification: %v\n", err)
				}

			case "create_message":

				var input string

				if len(params.Notification.Message.Content) > 0 && params.Notification.Message.Content[0].Input != nil {
					input = params.Notification.Message.Content[0].Input.Input
				}

				fmt.Fprintf(os.Stderr, "CREATE_MESSAGE: %v\n", input)

				patch, _ := utils.GetPatchResult(input)

				if len(patch.URI) > 0 {

					fmt.Fprintf(os.Stderr, "URI: %s\n", patch.URI)
					fmt.Fprintf(os.Stderr, "BEFORE: %s\n", patch.Before)
					fmt.Fprintf(os.Stderr, "AFTER: %s\n", patch.After)

					content := types.Content{
						Type:    "diff",
						Path:    patch.URI,
						OldText: patch.Before,
						NewText: patch.After,
					}

					update := types.SessionUpdateParam{
						SessionID: currentSession,
						Update: types.Update{
							SessionUpdate: "tool_call",
							ToolCallId:    params.Notification.Message.Content[0].Id,
							Kind:          "edit",
							Status:        "pending",
							Title:         patch.URI,
							Content:       &content,
						},
					}

					if err := sendACPNotification("session/update", update); err != nil {
						fmt.Fprintf(os.Stderr, "[ERROR] Failed to send session/update notification: %v\n", err)
					}
				}

			case "droid_working_state_changed":
				if params.Notification.NewState == "idle" {
					result := types.PromptResult{
						StopReason: "end_turn",
					}
					pendingPromptMu.Lock()
					promptID := pendingPromptID
					pendingPromptID = nil
					pendingPromptMu.Unlock()
					if promptID == nil {
						fmt.Fprintf(os.Stderr, "[WARN] Missing pending prompt ID; cannot respond\n")
						break
					}
					sendACPResponse(promptID, result)
				}

			case "mcp_status_changed":
				sendDroidOK(msg.ID)

			case "settings_updated":
				sendDroidOK(msg.ID)
				cwd := strings.TrimSpace(lastSessionCwd)
				if cwd == "" {
					cwd = "."
				}
				if err := initializeDroidSession(cwd); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to reinitialize droid session after settings update: %v\n", err)
				}

			default:
				fmt.Fprintf(os.Stderr, "Unknown droid Notification.Type: %s\n", params.Notification.Type)
				sendDroidOK(msg.ID)
			}

		case "droid.add_user_message":
			sendDroidOK(msg.ID)

		case "droid.request_permission":
			var params types.DroidNotification
			if err := json.Unmarshal(msg.Params, &params); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to parse droid.session_notification: %v\n", err)
				return
			}
			fmt.Fprintf(os.Stderr, "[DROID REQUEST PERMISSION] %v\n", string(msg.Params))

			toolUsesParent := params.ToolUses
			var options []types.PermissionOption = []types.PermissionOption{}
			var kind, label string
			for _, option := range params.Options {
				switch option.Value {
				case "proceed_once":
					kind = "allow_once"
					label = option.Label
				case "proceed_always":
					kind = "allow_always"
					label = "Yes, always"
				case "cancel":
					kind = "reject_once"
					label = option.Label
				}
				options = append(options, types.PermissionOption{
					OptionId: option.Value,
					Kind:     kind,
					Name:     label,
				})
			}

			for _, toolUses := range toolUsesParent {
				var title string
				if len(toolUses.Details.FullCommand) > 0 {
					title = toolUses.Details.FullCommand
				} else if toolUses.ConfirmationType == "create" {
					title = "create " + toolUses.Details.FilePath + "?"
				} else {
					title = "update"
				}
				register := types.SessionUpdateParam{
					SessionID: currentSession,
					Update: types.Update{
						SessionUpdate: "tool_call",
						ToolCallId:    toolUses.ToolUse.ID,
						Status:        "in_progress",
						Title:         title,
					},
				}

				if err := sendACPNotification("session/update", register); err != nil {
					fmt.Fprintf(os.Stderr, "[ERROR] Failed to send session/update notification: %v\n", err)
				}

				var request types.RequestPermissionParam
				if len(toolUses.Details.FullCommand) == 0 {
					var contents []types.DiffContent = []types.DiffContent{}
					var filePath, oldText, newText string
					inputRaw := toolUses.ToolUse.Input
					switch toolUses.ConfirmationType {
					case "apply_patch":
						var input types.InputApplyPatch
						if err := json.Unmarshal(inputRaw, &input); err != nil {
							fmt.Fprintf(os.Stderr, "Failed to parse droid.session_notification: %v\n", err)
							return
						}
						patch, _ := utils.GetPatchResult(input.Input)
						filePath = patch.URI
						oldText = patch.Before
						newText = patch.After
					case "edit":
						var input types.InputEdit
						if err := json.Unmarshal(inputRaw, &input); err != nil {
							fmt.Fprintf(os.Stderr, "Failed to parse droid.session_notification: %v\n", err)
							return
						}
						filePath = input.FilePath
						oldText = input.OldStr
						newText = input.NewString
					}
					contents = append(contents, types.DiffContent{
						Type:    "diff",
						Path:    filePath,
						OldText: oldText,
						NewText: newText,
					})
					request = types.RequestPermissionParam{
						SessionID: currentSession,
						ToolCall: types.ToolCall{
							ToolCallId: toolUses.ToolUse.ID,
							Title:      filePath,
							Kind:       "edit",
							Status:     "pending",
							Content:    contents,
						},
						Options: options,
					}

				} else {
					request = types.RequestPermissionParam{
						SessionID: currentSession,
						ToolCall: types.ToolCall{
							ToolCallId: toolUses.ToolUse.ID,
						},
						Options: options,
					}
				}
				if reqID, err := sendACPRequest("session/request_permission", request); err == nil {
					permissionRequestMu.Lock()
					permissionRequestMap[reqID] = permissionRequest{
						DroidRequestID: msg.ID,
						ToolCallID:     toolUses.ToolUse.ID,
					}
					permissionRequestMu.Unlock()
				} else {
					fmt.Fprintf(os.Stderr, "[ERROR] Failed to send session/request_permission request: %v\n", err)
				}
			}
		default:
			fmt.Fprintf(os.Stderr, "Unknown droid method: %s\n", msg.Method)
			sendDroidOK(msg.ID)
		}
	}
}

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "-v" || arg == "--version" {
			fmt.Println("v." + version)
			return
		}
	}

	acpOut = os.Stdout
	cmd := exec.Command(
		"droid",
		"exec",
		"--input-format", "stream-jsonrpc",
		"--output-format", "stream-jsonrpc",
	)
	var err error
	droidInPipe, err := cmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create stdin pipe: %v\n", err)
		os.Exit(1)
	}
	droidIn = droidInPipe
	droidOut, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create stdout pipe: %v\n", err)
		os.Exit(1)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start droid: %v\n", err)
		os.Exit(1)
	}
	go func() {
		scanner := bufio.NewScanner(droidOut)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			fmt.Fprintf(os.Stderr, "[DROID->] %s\n", line)

			var msg types.DroidMessage
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to parse droid message: %v\n", err)
				continue
			}
			handleDroidMessage(msg)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		fmt.Fprintf(os.Stderr, "[ZED->] %s\n", line)

		var req types.ACPRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse ACP request: %v\n", err)
			continue
		}

		handleACPRequest(req)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Scanner error: %v\n", err)
	}

	if err := cmd.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Droid exited with error: %v\n", err)
		os.Exit(1)
	}
}
