package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const discordAPIBase = "https://discord.com/api/v10"

// DiscordEditOriginalInteraction 编辑延迟响应（对应先回复 type 5 后的 @original）
func DiscordEditOriginalInteraction(ctx context.Context, applicationID, interactionToken, content string) error {
	applicationID = strings.TrimSpace(applicationID)
	interactionToken = strings.TrimSpace(interactionToken)
	if applicationID == "" || interactionToken == "" {
		return fmt.Errorf("discord: missing application_id or token")
	}
	content = strings.TrimSpace(content)
	if content == "" {
		content = "（无输出）"
	}
	if len(content) > 2000 {
		content = content[:1997] + "..."
	}
	u := fmt.Sprintf("%s/webhooks/%s/%s/messages/@original", discordAPIBase, applicationID, interactionToken)
	body, err := json.Marshal(map[string]string{"content": content})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, u, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := ExternalHTTPDo(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord PATCH @original: %s %s", resp.Status, string(raw))
	}
	return nil
}
