"use client"

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Trash2, Plus } from "lucide-react";
import axios from "@/lib/api";
import type { Policy, ScannerType } from "@/lib/types";

const EVENT_OPTIONS = [
  { value: "finding.created", label: "Finding Created" },
  { value: "finding.status_changed", label: "Status Changed" },
  { value: "finding.reassigned", label: "Finding Reassigned" },
];

const CONDITION_FIELDS = [
  { value: "severity", label: "Severity" },
  { value: "risk_score", label: "Risk Score" },
  { value: "scanner_type", label: "Scanner Type" },
  { value: "status", label: "Status" },
  { value: "cwe_id", label: "CWE ID" },
  { value: "file_path", label: "File Path" },
];

const SEVERITY_VALUES = ["Critical", "High", "Medium", "Low", "Info"];
const STATUS_VALUES = ["open", "confirmed", "false_positive", "fixed"];

const OPS_BY_FIELD: Record<string, string[]> = {
  severity: ["in", "not_in", "eq", "neq"],
  risk_score: ["gte", "lte", "eq", "neq"],
  scanner_type: ["in", "not_in", "eq", "neq"],
  status: ["in", "not_in", "eq", "neq"],
  cwe_id: ["eq", "neq", "contains"],
  file_path: ["eq", "neq", "contains"],
};

const ACTION_TYPES = [
  { value: "webhook", label: "Webhook" },
  { value: "change_status", label: "Change Status" },
  { value: "assign_to", label: "Assign To" },
];

const STATUS_TARGETS = ["open", "confirmed", "false_positive", "fixed"];

const PLATFORM_PRESETS: { value: string; label: string; payload: string }[] = [
  { value: "custom", label: "Custom", payload: "" },
  { value: "discord", label: "Discord", payload: `{"embeds":[{"title":"{{.Finding.Severity}}: {{.Finding.Title}}","description":"{{.Finding.Description}}","color":{{if eq .Finding.Severity "Critical"}}15158332{{else if eq .Finding.Severity "High"}}15105570{{else}}3447003{{end}},"fields":[{"name":"Application","value":"#{{.ApplicationID}}","inline":true},{"name":"File","value":"{{.Finding.FilePath}}","inline":true},{"name":"CWE","value":"{{.Finding.CWEID}}","inline":true}],"timestamp":"{{.Timestamp}}"}]}` },
  { value: "slack", label: "Slack", payload: `{"text":"New finding on app #{{.ApplicationID}}","attachments":[{"color":{{if eq .Finding.Severity "Critical"}}"danger"{{else if eq .Finding.Severity "High"}}"warning"{{else}}"good"{{end}},"fields":[{"title":"Severity","value":"{{.Finding.Severity}}","short":true},{"title":"Title","value":"{{.Finding.Title}}","short":true},{"title":"File","value":"{{.Finding.FilePath}}","short":false}]}]}` },
  { value: "teams", label: "Microsoft Teams", payload: `{"@type":"MessageCard","@context":"https://schema.org/extensions","summary":"{{.Finding.Severity}} finding","title":"{{.Finding.Title}}","themeColor":{{if eq .Finding.Severity "Critical"}}"E01E5A"{{else if eq .Finding.Severity "High"}}"FF8C00"{{else}}"0078D4"{{end}},"sections":[{"activityTitle":"**Severity:** {{.Finding.Severity}}","facts":[{"name":"Application","value":"#{{.ApplicationID}}"},{"name":"File","value":"{{.Finding.FilePath}}"},{"name":"CWE","value":"{{.Finding.CWEID}}"}],"markdown":true}]}` },
];

interface Condition {
  field: string;
  op: string;
  value: string;
}

interface Action {
  type: string;
  target: string;
  url?: string;
  secret?: string;
  payload?: string;
  platform?: string;
}

interface PolicyFormProps {
  initialData?: Policy;
  onSubmit: (data: Partial<Policy>) => Promise<void>;
  saving: boolean;
}

export function PolicyForm({ initialData, onSubmit, saving }: PolicyFormProps) {
  const [name, setName] = useState(initialData?.name || "");
  const [description, setDescription] = useState(initialData?.description || "");
  const [priority, setPriority] = useState(String(initialData?.priority ?? 10));
  const [isActive, setIsActive] = useState(initialData?.isActive ?? true);
  const [scopeType, setScopeType] = useState<string>(initialData?.scopeType || "global");
  const [scopeValue, setScopeValue] = useState(initialData?.scopeValue || "");
  const [selectedEvents, setSelectedEvents] = useState<string[]>(() => {
    if (!initialData?.eventTypes) return [];
    return initialData.eventTypes.split(",").map((s) => s.trim()).filter(Boolean);
  });
  const [conditions, setConditions] = useState<Condition[]>(() => {
    if (!initialData?.conditions) return [{ field: "severity", op: "in", value: "" }];
    try { return JSON.parse(initialData.conditions); } catch { return [{ field: "severity", op: "in", value: "" }]; }
  });
  const [actions, setActions] = useState<Action[]>(() => {
    if (!initialData?.actions) return [{ type: "webhook", target: "" }];
    try { return JSON.parse(initialData.actions); } catch { return [{ type: "webhook", target: "" }]; }
  });

  const [apps, setApps] = useState<{ id: number; name: string }[]>([]);
  const [scannerTypes, setScannerTypes] = useState<ScannerType[]>([]);

  useEffect(() => {
    if (scopeType === "application") {
      axios.get("/api/applications").then((res) => setApps(res.data)).catch(() => {});
    }
    axios.get("/api/scanner-types", { params: { enabled: "true" } })
      .then((res) => setScannerTypes(res.data))
      .catch(() => {});
  }, [scopeType]);

  const toggleEvent = (ev: string) => {
    setSelectedEvents((prev) =>
      prev.includes(ev) ? prev.filter((e) => e !== ev) : [...prev, ev]
    );
  };

  const addCondition = () => setConditions([...conditions, { field: "severity", op: "in", value: "" }]);
  const removeCondition = (i: number) => setConditions(conditions.filter((_, idx) => idx !== i));
  const updateCondition = (i: number, field: keyof Condition, val: string) => {
    const updated = conditions.map((c, idx) => {
      if (idx !== i) return c;
      const next = { ...c, [field]: val };
      if (field === "field") next.op = OPS_BY_FIELD[val]?.[0] || "eq";
      return next;
    });
    setConditions(updated);
  };

  const addAction = () => setActions([...actions, { type: "webhook", target: "", platform: "custom" }]);
  const removeAction = (i: number) => setActions(actions.filter((_, idx) => idx !== i));
  const updateAction = (i: number, field: keyof Action, val: string) => {
    const updated = actions.map((c, idx) => (idx !== i ? c : { ...c, [field]: val }));
    setActions(updated);
  };
  const setPlatform = (i: number, platform: string) => {
    const preset = PLATFORM_PRESETS.find((p) => p.value === platform);
    const updated = actions.map((c, idx) =>
      idx !== i ? c : { ...c, platform, payload: preset && platform !== "custom" ? preset.payload : c.payload }
    );
    setActions(updated);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    const scope = scopeType === "global" ? "" : scopeValue;

    await onSubmit({
      name: name.trim(),
      description: description.trim(),
      scopeType: scopeType as "application" | "group" | "global",
      scopeValue: scope,
      priority: parseInt(priority) || 10,
      isActive,
      eventTypes: selectedEvents.join(","),
      conditions: JSON.stringify(conditions),
      actions: JSON.stringify(actions),
    });
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-2">
        <div className="grid gap-2">
          <Label htmlFor="policyName">Name *</Label>
          <Input
            id="policyName"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. Notify on critical findings"
            required
            maxLength={200}
            autoComplete="off"
            data-1p-ignore
          />
        </div>

        <div className="grid gap-2">
          <Label htmlFor="policyPriority">Priority</Label>
          <Input
            id="policyPriority"
            type="number"
            min={1}
            max={100}
            value={priority}
            onChange={(e) => setPriority(e.target.value)}
            placeholder="10"
            autoComplete="off"
            data-1p-ignore
          />
          <p className="text-xs text-muted-foreground">Higher values are evaluated first</p>
        </div>
      </div>

      <div className="grid gap-2">
        <Label htmlFor="policyDescription">Description</Label>
        <Input
          id="policyDescription"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Optional description of this policy"
          maxLength={500}
          autoComplete="off"
          data-1p-ignore
        />
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="grid gap-2">
          <Label>Scope</Label>
          <Select value={scopeType} onValueChange={(v) => { setScopeType(v); setScopeValue(""); }}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="global">Global</SelectItem>
              <SelectItem value="application">Application</SelectItem>
              <SelectItem value="group">Group</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {scopeType === "application" && (
          <div className="grid gap-2">
            <Label>Application</Label>
            <Select value={scopeValue} onValueChange={setScopeValue}>
              <SelectTrigger>
                <SelectValue placeholder="Select application..." />
              </SelectTrigger>
              <SelectContent>
                {apps.map((a) => (
                  <SelectItem key={a.id} value={String(a.id)}>{a.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        )}

        {scopeType === "group" && (
          <div className="grid gap-2">
            <Label>Group ID</Label>
            <Input
              value={scopeValue}
              onChange={(e) => setScopeValue(e.target.value)}
              placeholder="Enter group ID"
              autoComplete="off"
              data-1p-ignore
            />
          </div>
        )}
      </div>

      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          id="policyActive"
          checked={isActive}
          onChange={(e) => setIsActive(e.target.checked)}
          className="h-4 w-4 rounded border-gray-300 text-purple-600 focus:ring-purple-500"
        />
        <Label htmlFor="policyActive" className="text-sm font-normal">Active</Label>
      </div>

      <div className="grid gap-2">
        <Label>Event Types</Label>
        <div className="flex flex-wrap gap-3">
          {EVENT_OPTIONS.map((ev) => (
            <label key={ev.value} className="flex items-center gap-1.5 text-sm">
              <input
                type="checkbox"
                checked={selectedEvents.includes(ev.value)}
                onChange={() => toggleEvent(ev.value)}
                className="h-4 w-4 rounded border-gray-300 text-purple-600 focus:ring-purple-500"
              />
              {ev.label}
            </label>
          ))}
        </div>
        {selectedEvents.length === 0 && (
          <p className="text-xs text-muted-foreground">Select at least one event type</p>
        )}
      </div>

      <div className="grid gap-3">
        <div className="flex items-center justify-between">
          <Label>Conditions</Label>
          <Button type="button" variant="outline" size="sm" onClick={addCondition} className="h-8 px-2.5 text-xs gap-1">
            <Plus className="h-3.5 w-3.5" />
            Add condition
          </Button>
        </div>
        {conditions.map((cond, i) => (
          <div key={i} className="flex flex-wrap items-end gap-2 p-3 rounded-lg border bg-muted/30">
            <div className="grid gap-1.5">
              <span className="text-xs text-muted-foreground">Field</span>
              <Select value={cond.field} onValueChange={(v) => updateCondition(i, "field", v)}>
                <SelectTrigger className="h-8 w-32">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {CONDITION_FIELDS.map((f) => (
                    <SelectItem key={f.value} value={f.value}>{f.label}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="grid gap-1.5">
              <span className="text-xs text-muted-foreground">Op</span>
              <Select value={cond.op} onValueChange={(v) => updateCondition(i, "op", v)}>
                <SelectTrigger className="h-8 w-24">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {(OPS_BY_FIELD[cond.field] || ["eq"]).map((op) => (
                    <SelectItem key={op} value={op}>{op}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="grid gap-1.5 flex-1 min-w-[120px]">
              <span className="text-xs text-muted-foreground">Value</span>
              {cond.field === "severity" && (cond.op === "in" || cond.op === "not_in") ? (
                <div className="flex flex-wrap gap-1.5 items-center h-8">
                  {SEVERITY_VALUES.map((sv) => (
                    <label key={sv} className="flex items-center gap-1 text-xs">
                      <input
                        type="checkbox"
                        checked={cond.value.split(",").map((s) => s.trim()).includes(sv)}
                        onChange={() => {
                          const current = cond.value.split(",").map((s) => s.trim()).filter(Boolean);
                          const next = current.includes(sv)
                            ? current.filter((v) => v !== sv)
                            : [...current, sv];
                          updateCondition(i, "value", next.join(","));
                        }}
                        className="h-3.5 w-3.5 rounded border-gray-300 text-purple-600 focus:ring-purple-500"
                      />
                      {sv}
                    </label>
                  ))}
                </div>
              ) : cond.field === "status" && (cond.op === "in" || cond.op === "not_in") ? (
                <div className="flex flex-wrap gap-1.5 items-center h-8">
                  {STATUS_VALUES.map((sv) => (
                    <label key={sv} className="flex items-center gap-1 text-xs">
                      <input
                        type="checkbox"
                        checked={cond.value.split(",").map((s) => s.trim()).includes(sv)}
                        onChange={() => {
                          const current = cond.value.split(",").map((s) => s.trim()).filter(Boolean);
                          const next = current.includes(sv)
                            ? current.filter((v) => v !== sv)
                            : [...current, sv];
                          updateCondition(i, "value", next.join(","));
                        }}
                        className="h-3.5 w-3.5 rounded border-gray-300 text-purple-600 focus:ring-purple-500"
                      />
                      {sv}
                    </label>
                  ))}
                </div>
              ) : cond.field === "scanner_type" && (cond.op === "in" || cond.op === "not_in") ? (
                <div className="flex flex-wrap gap-1.5 items-center h-8">
                  {scannerTypes.map((st) => (
                    <label key={st.id} className="flex items-center gap-1 text-xs">
                      <input
                        type="checkbox"
                        checked={cond.value.split(",").map((s) => s.trim()).includes(st.name)}
                        onChange={() => {
                          const current = cond.value.split(",").map((s) => s.trim()).filter(Boolean);
                          const next = current.includes(st.name)
                            ? current.filter((v) => v !== st.name)
                            : [...current, st.name];
                          updateCondition(i, "value", next.join(","));
                        }}
                        className="h-3.5 w-3.5 rounded border-gray-300 text-purple-600 focus:ring-purple-500"
                      />
                      {st.name}
                    </label>
                  ))}
                </div>
              ) : (
                <Input
                  className="h-8"
                  value={cond.value}
                  onChange={(e) => updateCondition(i, "value", e.target.value)}
                  placeholder={cond.field === "risk_score" ? "e.g. 0.5" : "Value"}
                  autoComplete="off"
                  data-1p-ignore
                />
              )}
            </div>

            <Button
              type="button"
              variant="destructive-ghost"
              size="icon"
              className="h-8 w-8 shrink-0"
              onClick={() => removeCondition(i)}
              disabled={conditions.length <= 1}
            >
              <Trash2 className="h-3.5 w-3.5" />
            </Button>
          </div>
        ))}
      </div>

      <div className="grid gap-3">
        <div className="flex items-center justify-between">
          <Label>Actions</Label>
          <Button type="button" variant="outline" size="sm" onClick={addAction} className="h-8 px-2.5 text-xs gap-1">
            <Plus className="h-3.5 w-3.5" />
            Add action
          </Button>
        </div>
        {actions.map((act, i) => (
          <div key={i} className="flex flex-wrap items-end gap-2 p-3 rounded-lg border bg-muted/30">
            <div className="grid gap-1.5">
              <span className="text-xs text-muted-foreground">Type</span>
              <Select value={act.type} onValueChange={(v) => updateAction(i, "type", v)}>
                <SelectTrigger className="h-8 w-36">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {ACTION_TYPES.map((at) => (
                    <SelectItem key={at.value} value={at.value}>{at.label}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {act.type === "change_status" && (
              <div className="grid gap-1.5">
                <span className="text-xs text-muted-foreground">Target Status</span>
                <Select value={act.target} onValueChange={(v) => updateAction(i, "target", v)}>
                  <SelectTrigger className="h-8 w-32">
                    <SelectValue placeholder="Select..." />
                  </SelectTrigger>
                  <SelectContent>
                    {STATUS_TARGETS.map((st) => (
                      <SelectItem key={st} value={st}>{st}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}

            {act.type === "assign_to" && (
              <div className="grid gap-1.5 flex-1 min-w-[120px]">
                <span className="text-xs text-muted-foreground">User ID</span>
                <Input
                  className="h-8"
                  value={act.target}
                  onChange={(e) => updateAction(i, "target", e.target.value)}
                  placeholder="Enter user ID"
                  autoComplete="off"
                  data-1p-ignore
                />
              </div>
            )}

            {act.type === "webhook" && (
              <div className="grid gap-1.5 flex-1 min-w-[120px]">
                <span className="text-xs text-muted-foreground">Platform</span>
                <Select value={act.platform || "custom"} onValueChange={(v) => setPlatform(i, v)}>
                  <SelectTrigger className="h-8 w-40">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {PLATFORM_PRESETS.map((p) => (
                      <SelectItem key={p.value} value={p.value}>{p.label}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <span className="text-xs text-muted-foreground">Webhook URL</span>
                <Input
                  className="h-8"
                  value={act.url || ""}
                  onChange={(e) => updateAction(i, "url", e.target.value)}
                  placeholder="https://discord.com/api/webhooks/..."
                  autoComplete="off"
                  data-1p-ignore
                />
                <span className="text-xs text-muted-foreground">Secret (optional)</span>
                <Input
                  className="h-8"
                  type="password"
                  value={act.secret || ""}
                  onChange={(e) => updateAction(i, "secret", e.target.value)}
                  placeholder="HMAC-SHA256 secret"
                  autoComplete="off"
                  data-1p-ignore
                />
                <span className="text-xs text-muted-foreground">Payload</span>
                <Textarea
                  value={act.payload || ""}
                  onChange={(e) => updateAction(i, "payload", e.target.value)}
                  placeholder={`{"content": "[{{.Finding.Severity}}] {{.Finding.Title}}"}`}
                  rows={5}
                  autoComplete="off"
                  data-1p-ignore
                />
              </div>
            )}

            <Button
              type="button"
              variant="destructive-ghost"
              size="icon"
              className="h-8 w-8 shrink-0"
              onClick={() => removeAction(i)}
              disabled={actions.length <= 1}
            >
              <Trash2 className="h-3.5 w-3.5" />
            </Button>
          </div>
        ))}
      </div>

      <div className="flex items-center gap-2 pt-2">
        <Button type="submit" disabled={saving || !name.trim() || selectedEvents.length === 0}>
          {saving ? "Saving..." : initialData ? "Save changes" : "Create policy"}
        </Button>
      </div>
    </form>
  );
}
