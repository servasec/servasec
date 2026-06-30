package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
)

const maxRecursionDepth = 3

type Condition struct {
	Field string      `json:"field"`
	Op    string      `json:"op"`
	Value interface{} `json:"value"`
}

type webhookTemplateData struct {
	Event         string
	ApplicationID uint
	Finding       *models.Finding
	Timestamp     string
}

type Action struct {
	Type    string `json:"type"`
	Target  string `json:"target,omitempty"`
	URL     string `json:"url,omitempty"`
	Secret  string `json:"secret,omitempty"`
	Payload string `json:"payload,omitempty"`
}

func EvaluatePolicies(eventType string, finding *models.Finding, app *models.Application, depth int) {
	if depth >= maxRecursionDepth {
		log.Printf("policy: max recursion depth reached for finding %d, event %s", finding.ID, eventType)
		return
	}

	var policies []models.Policy
	appStr := strconv.FormatUint(uint64(app.ID), 10)
	groupStr := strconv.FormatUint(uint64(app.GroupID), 10)

	config.DB.Where("is_active = ?", true).
		Where("event_types LIKE ?", "%"+eventType+"%").
		Where("(scope_type = 'application' AND scope_value = ?) OR (scope_type = 'group' AND scope_value = ?) OR scope_type = 'global'", appStr, groupStr).
		Order("priority DESC").
		Find(&policies)

	if len(policies) == 0 {
		return
	}

	var scannerType models.ScannerType
	config.DB.First(&scannerType, finding.ScannerTypeID)

	for i := range policies {
		policy := &policies[i]
		matched := true
		if policy.Conditions != "" && policy.Conditions != "null" {
			matched = evaluateConditions(policy.Conditions, finding, &scannerType)
		}

		if !matched {
			logPolicyExecution(policy, finding.ID, eventType, false, "skip", "conditions_not_met", "")
			continue
		}

		if policy.Actions == "" || policy.Actions == "null" {
			logPolicyExecution(policy, finding.ID, eventType, true, "skip", "no_actions", "")
			continue
		}

		var actions []Action
		if err := json.Unmarshal([]byte(policy.Actions), &actions); err != nil {
			logPolicyExecution(policy, finding.ID, eventType, true, "skip", fmt.Sprintf("invalid actions json: %v", err), "")
			continue
		}

		for _, action := range actions {
			result, detail := executeAction(action, finding, app, eventType)
			logPolicyExecution(policy, finding.ID, eventType, true, action.Type, result, detail)

			if (action.Type == "change_status" || action.Type == "assign_to") && result == "ok" {
				EvaluatePolicies(eventType, finding, app, depth+1)
			}
		}
	}
}

func evaluateConditions(conditionsJSON string, finding *models.Finding, scannerType *models.ScannerType) bool {
	var conditions []Condition
	if err := json.Unmarshal([]byte(conditionsJSON), &conditions); err != nil {
		log.Printf("policy: failed to parse conditions: %v", err)
		return true
	}

	for _, c := range conditions {
		if !evaluateCondition(c, finding, scannerType) {
			return false
		}
	}
	return true
}

func evaluateCondition(c Condition, finding *models.Finding, scannerType *models.ScannerType) bool {
	switch c.Field {
	case "severity":
		return matchStringList(c.Op, c.Value, finding.Severity)
	case "risk_score":
		return matchNumeric(c.Op, c.Value, finding.RiskScore)
	case "scanner_type":
		name := ""
		if scannerType != nil {
			name = scannerType.Name
		}
		return matchStringList(c.Op, c.Value, name)
	case "status":
		return matchStringList(c.Op, c.Value, finding.Status)
	case "cwe_id":
		return matchStringRegex(c.Op, c.Value, finding.CWEID)
	case "file_path":
		return matchStringRegex(c.Op, c.Value, finding.FilePath)
	default:
		return true
	}
}

func matchStringList(op string, value interface{}, field string) bool {
	list, ok := value.([]interface{})
	if !ok {
		return true
	}
	var strs []string
	for _, v := range list {
		if s, ok := v.(string); ok {
			strs = append(strs, s)
		}
	}
	for _, s := range strs {
		if field == s {
			return op == "in" || op == "eq"
		}
	}
	return op == "not_in" || op == "neq"
}

func matchNumeric(op string, value interface{}, field *float64) bool {
	if field == nil {
		return false
	}
	target, ok := toFloat64(value)
	if !ok {
		return true
	}
	switch op {
	case "gte":
		return *field >= target
	case "lte":
		return *field <= target
	case "eq":
		return *field == target
	case "neq":
		return *field != target
	default:
		return true
	}
}

func matchStringRegex(op string, value interface{}, field string) bool {
	pattern, ok := value.(string)
	if !ok || pattern == "" {
		return true
	}
	matched, err := regexp.MatchString(pattern, field)
	if err != nil {
		return true
	}
	switch op {
	case "regex", "in", "eq":
		return matched
	case "not_regex", "not_in", "neq":
		return !matched
	default:
		return true
	}
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func executeAction(action Action, finding *models.Finding, app *models.Application, eventType string) (result, detail string) {
	switch action.Type {
	case "webhook":
		return executeWebhookAction(action, finding, app, eventType)
	case "change_status":
		return executeChangeStatusAction(finding, action.Target)
	case "assign_to":
		return executeAssignAction(finding, action.Target)
	default:
		return "skip", fmt.Sprintf("unknown action type: %s", action.Type)
	}
}

func executeWebhookAction(action Action, finding *models.Finding, app *models.Application, eventType string) (string, string) {
	body, err := buildWebhookPayload(action.Payload, eventType, app, finding)
	if err != nil {
		return "failed", fmt.Sprintf("build payload: %v", err)
	}

	if action.URL != "" {
		if err := doSendWebhook(action.URL, action.Secret, body); err != nil {
			return "failed", fmt.Sprintf("direct webhook: %v", err)
		}
		return "ok", "sent to policy webhook target"
	}

	var webhooks []models.Webhook
	config.DB.Where("application_id = ? AND is_active = ?", app.ID, true).Find(&webhooks)
	if len(webhooks) == 0 {
		return "skip", "no active webhooks"
	}

	sent := 0
	for _, w := range webhooks {
		if err := sendWebhook(w, body); err == nil {
			sent++
		}
	}
	if sent == 0 {
		return "failed", "all webhooks failed"
	}
	return "ok", fmt.Sprintf("sent to %d webhook(s)", sent)
}

func executeChangeStatusAction(finding *models.Finding, target string) (string, string) {
	valid := map[string]bool{"open": true, "confirmed": true, "false_positive": true, "fixed": true, "closed": true}
	if !valid[target] {
		return "skip", fmt.Sprintf("invalid status: %s", target)
	}

	updates := map[string]interface{}{
		"status": target,
	}
	if target == "fixed" {
		now := time.Now()
		updates["fixed_at"] = &now
	}
	if err := config.DB.Model(finding).Updates(updates).Error; err != nil {
		return "failed", err.Error()
	}

	if target == "fixed" {
		now := time.Now()
		finding.FixedAt = &now
	}
	finding.Status = target
	return "ok", fmt.Sprintf("status changed to %s", target)
}

func executeAssignAction(finding *models.Finding, target string) (string, string) {
	id, err := strconv.ParseUint(target, 10, 64)
	if err != nil {
		return "skip", fmt.Sprintf("invalid user ID: %s", target)
	}
	uid := uint(id)

	var user models.User
	if err := config.DB.First(&user, uid).Error; err != nil {
		return "skip", fmt.Sprintf("user %d not found", uid)
	}

	if err := config.DB.Model(finding).Update("assigned_to", uid).Error; err != nil {
		return "failed", err.Error()
	}
	return "ok", fmt.Sprintf("assigned to user %d", uid)
}

func buildWebhookPayload(payloadTemplate string, eventType string, app *models.Application, finding *models.Finding) ([]byte, error) {
	if payloadTemplate == "" {
		log.Printf("webhook: using default payload for finding %d, event %s", finding.ID, eventType)
		payload := map[string]interface{}{
			"event":         eventType,
			"applicationId": app.ID,
			"finding":       finding,
			"timestamp":     time.Now().UTC().Format(time.RFC3339),
		}
		return json.Marshal(payload)
	}

	tmpl, err := template.New("webhook").Parse(payloadTemplate)
	if err != nil {
		return nil, fmt.Errorf("template parse: %w", err)
	}

	data := webhookTemplateData{
		Event:         eventType,
		ApplicationID: app.ID,
		Finding:       finding,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("template execute: %w", err)
	}

	result := buf.Bytes()
	log.Printf("webhook: rendered custom payload (%d bytes) for finding %d: %s", len(result), finding.ID, string(result))
	return result, nil
}

func sendWebhook(w models.Webhook, body []byte) error {
	return doSendWebhook(w.URL, w.Secret, body)
}

func doSendWebhook(url, secret string, body []byte) error {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		log.Printf("webhook: failed to create request for %s: %v", url, err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Servasec-Event", "policy.evaluation")

	if secret != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		sig := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Servasec-Signature", sig)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("webhook: failed to send to %s: %v", url, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("webhook: %s returned %d — body: %s", url, resp.StatusCode, string(respBody))
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

func logPolicyExecution(policy *models.Policy, findingID uint, eventType string, matched bool, actionType, result, detail string) {
	log.Printf("policy: [%s] policy=%d finding=%d action=%s result=%s detail=%s",
		eventType, policy.ID, findingID, actionType, result, detail)

	entry := models.PolicyLog{
		PolicyID:      policy.ID,
		FindingID:     findingID,
		EventType:     eventType,
		ConditionsMet: matched,
		ActionType:    actionType,
		ActionResult:  result,
		Detail:        truncate(detail, 500),
	}
	if err := config.DB.Create(&entry).Error; err != nil {
		log.Printf("policy: failed to log execution: %v", err)
	}
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max]
	}
	return s
}

func ScanToApp(scanID uint) *models.Application {
	var scan models.Scan
	if err := config.DB.Preload("ApplicationVersion").First(&scan, scanID).Error; err != nil {
		return nil
	}
	var app models.Application
	if err := config.DB.First(&app, scan.ApplicationVersion.ApplicationID).Error; err != nil {
		return nil
	}
	return &app
}

func NormalizeEventType(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "_")
}
