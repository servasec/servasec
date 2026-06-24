package mcp

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"github.com/servasec/servasec/backend/models"
)

type RPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.Number     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type RPCResponse struct {
	JSONRPC string       `json:"jsonrpc"`
	ID      json.Number  `json:"id"`
	Result  any          `json:"result,omitempty"`
	Error   *RPCError    `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type ToolDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"inputSchema"`
}

type ToolResultContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ToolResult struct {
	Content []ToolResultContent `json:"content"`
	IsError bool                `json:"isError"`
}

type InitializeResult struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ServerInfo      ServerInfo     `json:"serverInfo"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ToolListResult struct {
	Tools []ToolDefinition `json:"tools"`
}

type toolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

type HandlerContext struct {
	DB            *gorm.DB
	UserID        uint
	User          *models.User
	AccessibleIDs []string
}

type toolHandler func(HandlerContext, map[string]any) (string, bool)

type toolDef struct {
	Name        string
	Description string
	InputSchema map[string]any
	Handler     toolHandler
}

var tools = []toolDef{
	{
		Name:        "list_applications",
		Description: "List all applications",
		InputSchema: toolObjectSchema(nil, nil),
		Handler:     handleListApplications,
	},
	{
		Name:        "get_application",
		Description: "Get a single application by ID or slug",
		InputSchema: toolObjectSchema(map[string]map[string]any{
			"id": stringProp("Application ID or slug"),
		}, []string{"id"}),
		Handler: handleGetApplication,
	},
	{
		Name:        "list_findings",
		Description: "List findings with optional filters",
		InputSchema: toolObjectSchema(map[string]map[string]any{
			"applicationId":        stringProp("Filter by application ID"),
			"severity":             stringProp("Filter by severity (critical, high, medium, low)"),
			"status":               stringProp("Filter by status (open, confirmed, false_positive, fixed, closed)"),
			"sortBy":               stringProp("Sort field (created_at, risk_score, severity, title)"),
			"order":                stringProp("Sort order (asc or desc)"),
			"page":                 stringProp("Page number"),
			"perPage":              stringProp("Items per page (max 200)"),
			"applicationVersionId": stringProp("Filter by application version ID"),
			"scannerTypeId":        stringProp("Filter by scanner type ID"),
			"scanId":               stringProp("Filter by scan ID"),
		}, nil),
		Handler: handleListFindings,
	},
	{
		Name:        "get_finding",
		Description: "Get a single finding by ID",
		InputSchema: toolObjectSchema(map[string]map[string]any{
			"id": stringProp("Finding ID"),
		}, []string{"id"}),
		Handler: handleGetFinding,
	},
	{
		Name:        "update_finding_status",
		Description: "Update the status of a finding",
		InputSchema: toolObjectSchema(map[string]map[string]any{
			"id":     stringProp("Finding ID"),
			"status": stringProp("New status: open, confirmed, false_positive, fixed, closed"),
		}, []string{"id", "status"}),
		Handler: handleUpdateFindingStatus,
	},
	{
		Name:        "assign_finding",
		Description: "Assign a finding to a user with optional due date",
		InputSchema: toolObjectSchema(map[string]map[string]any{
			"id":      stringProp("Finding ID"),
			"userId":  intProp("User ID to assign to (omit to unassign)"),
			"dueDate": stringProp("Due date in ISO 8601 format (e.g. 2025-12-31T23:59:59Z)"),
		}, []string{"id"}),
		Handler: handleAssignFinding,
	},
	{
		Name:        "list_scans",
		Description: "List scans with optional filters",
		InputSchema: toolObjectSchema(map[string]map[string]any{
			"applicationId":        stringProp("Filter by application ID"),
			"applicationVersionId": stringProp("Filter by application version ID"),
		}, nil),
		Handler: handleListScans,
	},
	{
		Name:        "get_scan",
		Description: "Get a single scan by ID",
		InputSchema: toolObjectSchema(map[string]map[string]any{
			"id": stringProp("Scan ID"),
		}, []string{"id"}),
		Handler: handleGetScan,
	},
	{
		Name:        "get_dashboard_stats",
		Description: "Get dashboard statistics (findings by severity, status, top findings, etc.)",
		InputSchema: toolObjectSchema(nil, nil),
		Handler:     handleDashboardStats,
	},
	{
		Name:        "list_versions",
		Description: "List versions of an application",
		InputSchema: toolObjectSchema(map[string]map[string]any{
			"applicationId": stringProp("Application ID"),
		}, []string{"applicationId"}),
		Handler: handleListVersions,
	},
	{
		Name:        "get_version",
		Description: "Get a single application version by ID",
		InputSchema: toolObjectSchema(map[string]map[string]any{
			"applicationId": stringProp("Application ID"),
			"versionId":     stringProp("Version ID"),
		}, []string{"applicationId", "versionId"}),
		Handler: handleGetVersion,
	},
	{
		Name:        "list_version_findings",
		Description: "List findings for a specific application version",
		InputSchema: toolObjectSchema(map[string]map[string]any{
			"applicationId": stringProp("Application ID"),
			"versionId":     stringProp("Version ID"),
		}, []string{"applicationId", "versionId"}),
		Handler: handleListVersionFindings,
	},
}

func dispatchRequest(ctx HandlerContext, req *RPCRequest) RPCResponse {
	switch req.Method {
	case "initialize":
		return RPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: InitializeResult{
				ProtocolVersion: "2024-11-05",
				Capabilities:    map[string]any{"tools": map[string]any{}},
				ServerInfo: ServerInfo{
					Name:    "servasec",
					Version: "0.1.0",
				},
			},
		}

	case "tools/list":
		defs := make([]ToolDefinition, len(tools))
		for i, t := range tools {
			defs[i] = ToolDefinition{
				Name:        t.Name,
				Description: t.Description,
				InputSchema: t.InputSchema,
			}
		}
		return RPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  ToolListResult{Tools: defs},
		}

	case "tools/call":
		var params toolCallParams
		if len(req.Params) > 0 {
			if err := json.Unmarshal(req.Params, &params); err != nil {
				return toolError(req.ID, -32602, fmt.Sprintf("invalid params: %v", err))
			}
		}
		if params.Name == "" {
			return toolError(req.ID, -32602, "tool name is required")
		}
		if params.Arguments == nil {
			params.Arguments = map[string]any{}
		}

		var handler toolHandler
		for _, t := range tools {
			if t.Name == params.Name {
				handler = t.Handler
				break
			}
		}
		if handler == nil {
			return RPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: ToolResult{
					Content: []ToolResultContent{{
						Type: "text",
						Text: fmt.Sprintf("unknown tool: %s", params.Name),
					}},
					IsError: true,
				},
			}
		}

		text, isErr := handler(ctx, params.Arguments)
		return RPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: ToolResult{
				Content: []ToolResultContent{{Type: "text", Text: text}},
				IsError: isErr,
			},
		}

	default:
		return toolError(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
	}
}

func toolError(id json.Number, code int, msg string) RPCResponse {
	return RPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: msg,
		},
	}
}

func toolObjectSchema(properties map[string]map[string]any, required []string) map[string]any {
	schema := map[string]any{"type": "object"}
	if properties != nil {
		schema["properties"] = properties
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func stringProp(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}

func intProp(description string) map[string]any {
	return map[string]any{"type": "number", "description": description}
}

func getString(args map[string]any, key string) string {
	v, ok := args[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func dumpJSON(v any) (string, bool) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", true
	}
	return string(data), false
}

func findingsQuery(db *gorm.DB, accessibleIDs []string) *gorm.DB {
	q := db.Model(&models.Finding{}).
		Joins("JOIN application_versions ON application_versions.id = findings.application_version_id")

	if accessibleIDs != nil {
		if len(accessibleIDs) == 0 {
			return q.Where("1 = 0")
		}
		q = q.Where("application_versions.application_id IN ?", accessibleIDs)
	}

	return q
}

func scansQuery(db *gorm.DB, accessibleIDs []string) *gorm.DB {
	q := db.Model(&models.Scan{}).
		Joins("JOIN application_versions ON application_versions.id = scans.application_version_id")

	if accessibleIDs != nil {
		if len(accessibleIDs) == 0 {
			return q.Where("1 = 0")
		}
		q = q.Where("application_versions.application_id IN ?", accessibleIDs)
	}

	return q
}

// tool handlers

func handleListApplications(ctx HandlerContext, args map[string]any) (string, bool) {
	q := ctx.DB.Model(&models.Application{}).Preload("Versions")
	if ctx.AccessibleIDs != nil {
		if len(ctx.AccessibleIDs) == 0 {
			return "[]", false
		}
		q = q.Where("id IN ?", ctx.AccessibleIDs)
	}
	var apps []models.Application
	if err := q.Find(&apps).Error; err != nil {
		return fmt.Sprintf("error: %v", err), true
	}
	result := make([]map[string]any, len(apps))
	for i, app := range apps {
		result[i] = map[string]any{
			"id":                app.ID,
			"name":              app.Name,
			"description":       app.Description,
			"slug":              app.Slug,
			"groupId":           app.GroupID,
			"repositoryUrl":     app.RepositoryURL,
			"assetCriticality":  app.AssetCriticality,
			"createdAt":         app.CreatedAt,
			"updatedAt":         app.UpdatedAt,
		}
	}
	return dumpJSON(result)
}

func handleGetApplication(ctx HandlerContext, args map[string]any) (string, bool) {
	id := getString(args, "id")
	if id == "" {
		return "id is required", true
	}

	var app models.Application
	if err := ctx.DB.Preload("Versions").
		Where("id = ? OR slug = ?", id, id).
		First(&app).Error; err != nil {
		return fmt.Sprintf("application not found: %v", err), true
	}

	if ctx.AccessibleIDs != nil {
		ok := false
		for _, aid := range ctx.AccessibleIDs {
			if aid == fmt.Sprint(app.ID) {
				ok = true
				break
			}
		}
		if !ok {
			return "access denied", true
		}
	}

	return dumpJSON(map[string]any{
		"id":               app.ID,
		"name":             app.Name,
		"description":      app.Description,
		"slug":             app.Slug,
		"groupId":          app.GroupID,
		"repositoryUrl":    app.RepositoryURL,
		"assetCriticality": app.AssetCriticality,
		"createdAt":        app.CreatedAt,
		"updatedAt":        app.UpdatedAt,
	})
}

func handleListFindings(ctx HandlerContext, args map[string]any) (string, bool) {
	page := 1
	perPage := 50
	if p := getString(args, "page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	if pp := getString(args, "perPage"); pp != "" {
		if n, err := strconv.Atoi(pp); err == nil && n > 0 {
			perPage = n
			if perPage > 200 {
				perPage = 200
			}
		}
	}
	offset := (page - 1) * perPage

	base := findingsQuery(ctx.DB, ctx.AccessibleIDs)

	if appID := getString(args, "applicationId"); appID != "" {
		base = base.Where("application_versions.application_id = ?", appID)
	}
	if versionID := getString(args, "applicationVersionId"); versionID != "" {
		base = base.Where("findings.application_version_id = ?", versionID)
	}
	if severity := getString(args, "severity"); severity != "" {
		base = base.Where("findings.severity = ?", strings.ToLower(severity))
	}
	if status := getString(args, "status"); status != "" {
		base = base.Where("findings.status = ?", status)
	}
	if scannerTypeID := getString(args, "scannerTypeId"); scannerTypeID != "" {
		base = base.Where("findings.scanner_type_id = ?", scannerTypeID)
	}
	if scanID := getString(args, "scanId"); scanID != "" {
		base = base.Where("findings.scan_id = ?", scanID)
	}

	sortBy := getString(args, "sortBy")
	sortOrder := getString(args, "order")
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	orderClause := "findings.created_at DESC"
	switch sortBy {
	case "risk_score":
		orderClause = "findings.risk_score " + sortOrder + " NULLS LAST"
	case "severity":
		orderClause = "findings.severity " + sortOrder
	case "title":
		orderClause = "findings.title " + sortOrder
	case "created_at":
		orderClause = "findings.created_at " + sortOrder
	}

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return fmt.Sprintf("count error: %v", err), true
	}

	var findings []models.Finding
	if err := base.
		Session(&gorm.Session{}).
		Preload("ScannerType").
		Preload("ApplicationVersion").
		Preload("AssignedToUser").
		Preload("ReviewedByUser").
		Order(orderClause).
		Offset(offset).
		Limit(perPage).
		Find(&findings).Error; err != nil {
		return fmt.Sprintf("query error: %v", err), true
	}

	if findings == nil {
		findings = []models.Finding{}
	}

	return dumpJSON(map[string]any{
		"data":    findings,
		"total":   total,
		"page":    page,
		"perPage": perPage,
	})
}

func handleGetFinding(ctx HandlerContext, args map[string]any) (string, bool) {
	id := getString(args, "id")
	if id == "" {
		return "id is required", true
	}

	var finding models.Finding
	if err := ctx.DB.
		Preload("ScannerType").
		Preload("ApplicationVersion").
		Preload("AssignedToUser").
		Preload("ReviewedByUser").
		First(&finding, id).Error; err != nil {
		return fmt.Sprintf("finding not found: %v", err), true
	}

	if ctx.AccessibleIDs != nil {
		q := findingsQuery(ctx.DB, ctx.AccessibleIDs).Where("findings.id = ?", id)
		var count int64
		q.Count(&count)
		if count == 0 {
			return "access denied", true
		}
	}

	return dumpJSON(finding)
}

func handleUpdateFindingStatus(ctx HandlerContext, args map[string]any) (string, bool) {
	id := getString(args, "id")
	status := getString(args, "status")
	if id == "" || status == "" {
		return "id and status are required", true
	}
	valid := map[string]bool{"open": true, "confirmed": true, "false_positive": true, "fixed": true, "closed": true}
	if !valid[status] {
		return "invalid status: must be one of: open, confirmed, false_positive, fixed, closed", true
	}

	var finding models.Finding
	if err := ctx.DB.First(&finding, id).Error; err != nil {
		return fmt.Sprintf("finding not found: %v", err), true
	}

	finding.Status = status
	if status == "fixed" && finding.FixedAt == nil {
		now := time.Now()
		finding.FixedAt = &now
	}
	if err := ctx.DB.Save(&finding).Error; err != nil {
		return fmt.Sprintf("update error: %v", err), true
	}

	return dumpJSON(finding)
}

func handleAssignFinding(ctx HandlerContext, args map[string]any) (string, bool) {
	id := getString(args, "id")
	if id == "" {
		return "id is required", true
	}

	var finding models.Finding
	if err := ctx.DB.First(&finding, id).Error; err != nil {
		return fmt.Sprintf("finding not found: %v", err), true
	}

	if v, ok := args["userId"]; ok {
		switch val := v.(type) {
		case float64:
			u := uint(val)
			var user models.User
			if err := ctx.DB.First(&user, u).Error; err != nil {
				return fmt.Sprintf("user not found: %v", err), true
			}
			finding.AssignedTo = &u
		case nil:
			finding.AssignedTo = nil
		}
	}
	if v, ok := args["dueDate"].(string); ok && v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err == nil {
			finding.DueDate = &t
		}
	}

	if err := ctx.DB.Save(&finding).Error; err != nil {
		return fmt.Sprintf("assign error: %v", err), true
	}

	return dumpJSON(finding)
}

func handleListScans(ctx HandlerContext, args map[string]any) (string, bool) {
	base := scansQuery(ctx.DB, ctx.AccessibleIDs)

	if appID := getString(args, "applicationId"); appID != "" {
		base = base.Where("application_versions.application_id = ?", appID)
	}
	if versionID := getString(args, "applicationVersionId"); versionID != "" {
		base = base.Where("scans.application_version_id = ?", versionID)
	}

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return fmt.Sprintf("count error: %v", err), true
	}

	var scans []models.Scan
	if err := base.
		Preload("ApplicationVersion").
		Preload("ScannerType").
		Order("created_at DESC").
		Find(&scans).Error; err != nil {
		return fmt.Sprintf("query error: %v", err), true
	}

	if scans == nil {
		scans = []models.Scan{}
	}

	return dumpJSON(map[string]any{
		"data":    scans,
		"total":   total,
		"page":    1,
		"perPage": len(scans),
	})
}

func handleGetScan(ctx HandlerContext, args map[string]any) (string, bool) {
	id := getString(args, "id")
	if id == "" {
		return "id is required", true
	}

	var scan models.Scan
	if err := ctx.DB.
		Preload("ApplicationVersion").
		Preload("ScannerType").
		First(&scan, id).Error; err != nil {
		return fmt.Sprintf("scan not found: %v", err), true
	}

	return dumpJSON(scan)
}

func handleDashboardStats(ctx HandlerContext, args map[string]any) (string, bool) {
	df := findingsQuery(ctx.DB, ctx.AccessibleIDs)
	ds := scansQuery(ctx.DB, ctx.AccessibleIDs)

	stats := map[string]any{}

	var totalFindings int64
	df.Session(&gorm.Session{}).Count(&totalFindings)
	stats["totalFindings"] = totalFindings

	var bySeverity []map[string]any
	df.Session(&gorm.Session{}).
		Select("severity, count(*) as count").
		Group("severity").
		Order("count DESC").
		Find(&bySeverity)
	stats["bySeverity"] = bySeverity

	var byStatus []map[string]any
	df.Session(&gorm.Session{}).
		Select("status, count(*) as count").
		Group("status").
		Order("count DESC").
		Find(&byStatus)
	stats["byStatus"] = byStatus

	var byScanner []map[string]any
	df.Session(&gorm.Session{}).
		Joins("JOIN scanner_types ON scanner_types.id = findings.scanner_type_id").
		Select("scanner_types.name as scanner_type, count(*) as count").
		Group("scanner_types.name").
		Find(&byScanner)
	stats["byScanner"] = byScanner

	var recentScans int64
	ds.Where("status = ?", "completed").Count(&recentScans)
	stats["recentScans"] = recentScans

	var topFindings []map[string]any
	df.Session(&gorm.Session{}).
		Select("rule_id, title, count(*) as count").
		Group("rule_id, title").
		Order("count DESC").
		Limit(5).
		Find(&topFindings)
	stats["topFindings"] = topFindings

	var myOpenFindings int64
	df.Where("findings.assigned_to = ? AND findings.status IN ?", ctx.UserID, []string{"open", "confirmed"}).
		Count(&myOpenFindings)
	stats["myOpenFindings"] = myOpenFindings

	var overdueFindings int64
	df.Where("findings.due_date IS NOT NULL AND findings.due_date < NOW() AND findings.status NOT IN ?",
		[]string{"fixed", "closed", "false_positive"}).
		Count(&overdueFindings)
	stats["overdueFindings"] = overdueFindings

	var avgRiskScore *float64
	df.Where("risk_score IS NOT NULL").
		Select("AVG(risk_score)").
		Scan(&avgRiskScore)
	if avgRiskScore != nil {
		stats["avgRiskScore"] = *avgRiskScore
	}

	var topRisky []map[string]any
	df.Where("risk_score IS NOT NULL").
		Order("risk_score DESC").
		Limit(5).
		Select("id, title, risk_score, severity").
		Scan(&topRisky)
	if len(topRisky) > 0 {
		stats["topRiskyFindings"] = topRisky
	}

	var riskDist []map[string]any
	df.Where("risk_score IS NOT NULL").
		Select(`
			CASE
				WHEN risk_score >= 0.8 THEN '80-100'
				WHEN risk_score >= 0.6 THEN '60-80'
				WHEN risk_score >= 0.4 THEN '40-60'
				WHEN risk_score >= 0.2 THEN '20-40'
				ELSE '0-20'
			END AS label,
			COUNT(*) AS count
		`).
		Group("label").
		Order("label").
		Scan(&riskDist)
	if len(riskDist) > 0 {
		stats["riskDistribution"] = riskDist
	}

	return dumpJSON(stats)
}

func handleListVersions(ctx HandlerContext, args map[string]any) (string, bool) {
	appID := getString(args, "applicationId")
	if appID == "" {
		return "applicationId is required", true
	}

	var versions []models.ApplicationVersion
	if err := ctx.DB.Where("application_id = ?", appID).
		Order("created_at DESC").
		Find(&versions).Error; err != nil {
		return fmt.Sprintf("query error: %v", err), true
	}
	if versions == nil {
		versions = []models.ApplicationVersion{}
	}
	return dumpJSON(versions)
}

func handleGetVersion(ctx HandlerContext, args map[string]any) (string, bool) {
	appID := getString(args, "applicationId")
	versionID := getString(args, "versionId")
	if appID == "" || versionID == "" {
		return "applicationId and versionId are required", true
	}

	var version models.ApplicationVersion
	if err := ctx.DB.
		Where("id = ? AND application_id = ?", versionID, appID).
		First(&version).Error; err != nil {
		return fmt.Sprintf("version not found: %v", err), true
	}
	return dumpJSON(version)
}

func handleListVersionFindings(ctx HandlerContext, args map[string]any) (string, bool) {
	appID := getString(args, "applicationId")
	versionID := getString(args, "versionId")
	if appID == "" || versionID == "" {
		return "applicationId and versionId are required", true
	}

	base := findingsQuery(ctx.DB, ctx.AccessibleIDs).
		Where("findings.application_version_id = ?", versionID).
		Where("application_versions.application_id = ?", appID)

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return fmt.Sprintf("count error: %v", err), true
	}

	var findings []models.Finding
	if err := base.
		Preload("ScannerType").
		Preload("ApplicationVersion").
		Preload("AssignedToUser").
		Preload("ReviewedByUser").
		Order("findings.created_at DESC").
		Find(&findings).Error; err != nil {
		return fmt.Sprintf("query error: %v", err), true
	}
	if findings == nil {
		findings = []models.Finding{}
	}

	return dumpJSON(map[string]any{
		"data":    findings,
		"total":   total,
		"page":    1,
		"perPage": len(findings),
	})
}
