#!/usr/bin/env python3
"""Generate individual Bruno demo-setup requests.

Each .bru file IS the real HTTP request (method, URL, body correct).
Scripts only: generate run_id (pre-request) or store entity IDs (post-response).
No axios calls in scripts.
"""

import os, json, shutil

BASE = os.path.dirname(os.path.abspath(__file__))
OUT = os.path.join(BASE, "demo-setup")
SAMPLES = os.path.join(OUT, "samples")
generated = set()

SEQ = 0
def next_seq():
    global SEQ
    SEQ += 1
    return SEQ

def w(path, content):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as f:
        f.write(content.lstrip("\n"))
    generated.add(os.path.relpath(path, OUT))

def meta(name, seq):
    return f"meta {{\n  name: {name}\n  type: http\n  seq: {seq}\n}}\n"

# ==============================================================
# Sample scanner data (5 files)
# ==============================================================
os.makedirs(SAMPLES, exist_ok=True)

samples = {
    "semgrep-webapp.json": {"results": [
        {"check_id": "javascript.express.security.audit.xss", "path": "src/web/routes/users.js", "start": {"line": 15}, "end": {"line": 22}, "extra": {"message": "Reflected XSS via unsanitized user input in search query", "severity": "ERROR", "metadata": {"cwe": "CWE-79"}}},
        {"check_id": "javascript.express.security.audit.xss", "path": "src/web/views/comments.ejs", "start": {"line": 42}, "end": {"line": 48}, "extra": {"message": "Stored XSS in comment rendering", "severity": "ERROR", "metadata": {"cwe": "CWE-79"}}},
        {"check_id": "javascript.react.security.audit.dangerouslysetinnerhtml", "path": "src/web/components/Profile.jsx", "start": {"line": 88}, "end": {"line": 92}, "extra": {"message": "dangerouslySetInnerHTML used with unsanitized input", "severity": "WARNING", "metadata": {"cwe": "CWE-79"}}},
        {"check_id": "javascript.lang.security.audit.insecure-randomness", "path": "src/utils/tokens.js", "start": {"line": 15}, "end": {"line": 15}, "extra": {"message": "Insecure Math.random() used for token generation", "severity": "WARNING", "metadata": {"cwe": "CWE-330"}}},
        {"check_id": "javascript.express.security.audit.path-traversal", "path": "src/web/routes/files.js", "start": {"line": 34}, "end": {"line": 38}, "extra": {"message": "Path traversal via user-controlled filename", "severity": "ERROR", "metadata": {"cwe": "CWE-22"}}},
    ]},
    "semgrep-api.json": {"results": [
        {"check_id": "python.lang.security.audit.dangerous-system-call.dangerous-system-call", "path": "src/api/deploy.py", "start": {"line": 30}, "end": {"line": 30}, "extra": {"message": "Command injection via os.system() in deploy script", "severity": "ERROR", "metadata": {"cwe": "CWE-78"}}},
        {"check_id": "python.lang.security.audit.dangerous-system-call.dangerous-system-call", "path": "src/api/backup.py", "start": {"line": 45}, "end": {"line": 45}, "extra": {"message": "Shell injection in backup command", "severity": "ERROR", "metadata": {"cwe": "CWE-78"}}},
        {"check_id": "python.lang.security.audit.exec-unsafe", "path": "src/api/exec.py", "start": {"line": 22}, "end": {"line": 22}, "extra": {"message": "Unsafe eval() of user-controlled input", "severity": "ERROR", "metadata": {"cwe": "CWE-95"}}},
        {"check_id": "python.flask.security.audit.flask-secure-cookie", "path": "src/web/session.py", "start": {"line": 15}, "end": {"line": 15}, "extra": {"message": "Session cookie without Secure/HttpOnly flags", "severity": "WARNING", "metadata": {"cwe": "CWE-614"}}},
        {"check_id": "python.flask.security.audit.flask-debug-mode", "path": "src/api/app.py", "start": {"line": 5}, "end": {"line": 5}, "extra": {"message": "Flask debug mode enabled in production", "severity": "HIGH", "metadata": {"cwe": "CWE-489"}}},
        {"check_id": "python.lang.security.audit.sql-injection", "path": "src/api/users.py", "start": {"line": 60}, "end": {"line": 65}, "extra": {"message": "SQL injection via f-string in query", "severity": "ERROR", "metadata": {"cwe": "CWE-89"}}},
    ]},
    "semgrep-secrets.json": {"results": [
        {"check_id": "generic.secrets.security.detected-private-key", "path": "config/keys/id_rsa", "start": {"line": 1}, "end": {"line": 1}, "extra": {"message": "SSH private key exposed in repository", "severity": "ERROR", "metadata": {"cwe": "CWE-312"}}},
        {"check_id": "generic.secrets.security.detected-private-key", "path": "config/keys/deploy.key", "start": {"line": 1}, "end": {"line": 1}, "extra": {"message": "Deploy key exposed in repository", "severity": "ERROR", "metadata": {"cwe": "CWE-312"}}},
        {"check_id": "generic.secrets.security.detected-aws-credentials", "path": "config/.env", "start": {"line": 5}, "end": {"line": 5}, "extra": {"message": "AWS access key ID in configuration file", "severity": "ERROR", "metadata": {"cwe": "CWE-798"}}},
        {"check_id": "generic.secrets.security.detected-database-credentials", "path": "config/database.yml", "start": {"line": 8}, "end": {"line": 8}, "extra": {"message": "Database password in plaintext config file", "severity": "WARNING", "metadata": {"cwe": "CWE-798"}}},
        {"check_id": "generic.secrets.security.detected-generic-secret", "path": "config/.env", "start": {"line": 12}, "end": {"line": 12}, "extra": {"message": "JWT signing secret exposed in .env", "severity": "ERROR", "metadata": {"cwe": "CWE-312"}}},
    ]},
    "trivy-deps.json": {"Results": [
        {"Target": "package-lock.json", "Type": "npm", "Vulnerabilities": [
            {"VulnerabilityID": "CVE-2024-10001", "PkgName": "lodash", "Severity": "CRITICAL", "Title": "Prototype Pollution", "Description": "Prototype pollution via defaultsDeep in lodash", "InstalledVersion": "4.17.20", "FixedVersion": "4.17.21"},
            {"VulnerabilityID": "CVE-2024-10002", "PkgName": "express", "Severity": "HIGH", "Title": "DoS in body parser", "Description": "Denial of service via malformed HTTP body", "InstalledVersion": "4.18.1", "FixedVersion": "4.19.0"},
            {"VulnerabilityID": "CVE-2024-10003", "PkgName": "follow-redirects", "Severity": "MEDIUM", "Title": "Credentials leak", "Description": "Credentials exposed on HTTP redirect", "InstalledVersion": "1.15.4", "FixedVersion": "1.15.6"},
            {"VulnerabilityID": "CVE-2024-10004", "PkgName": "axios", "Severity": "HIGH", "Title": "SSRF via URL parser", "Description": "Server-Side Request Forgery in URL parsing", "InstalledVersion": "1.6.0", "FixedVersion": "1.7.0"},
            {"VulnerabilityID": "CVE-2024-10005", "PkgName": "json5", "Severity": "MEDIUM", "Title": "Prototype Pollution", "Description": "Prototype pollution in JSON5 parser", "InstalledVersion": "2.2.3", "FixedVersion": "2.2.4"},
            {"VulnerabilityID": "CVE-2024-10006", "PkgName": "undici", "Severity": "HIGH", "Title": "HTTP Request Smuggling", "Description": "Request smuggling via chunked encoding", "InstalledVersion": "5.28.0", "FixedVersion": "5.28.5"},
            {"VulnerabilityID": "CVE-2024-10007", "PkgName": "ejs", "Severity": "CRITICAL", "Title": "RCE via template injection", "Description": "Remote code execution in EJS templates", "InstalledVersion": "3.1.9", "FixedVersion": "3.1.10"},
            {"VulnerabilityID": "CVE-2024-10008", "PkgName": "path-to-regexp", "Severity": "MEDIUM", "Title": "ReDoS", "Description": "Regular expression denial of service", "InstalledVersion": "0.1.7", "FixedVersion": "0.1.10"},
        ]},
        {"Target": "python-requirements.txt", "Type": "pip", "Vulnerabilities": [
            {"VulnerabilityID": "CVE-2024-10009", "PkgName": "flask", "Severity": "HIGH", "Title": "DoS via malformed multipart", "Description": "Denial of service in Flask multipart parsing", "InstalledVersion": "2.3.0", "FixedVersion": "2.3.3"},
            {"VulnerabilityID": "CVE-2024-10010", "PkgName": "django", "Severity": "CRITICAL", "Title": "SQL injection in queryset", "Description": "SQL injection in Django queryset filtering", "InstalledVersion": "4.2.0", "FixedVersion": "4.2.15"},
        ]},
        {"Target": "Dockerfile", "Type": "dockerfile", "Misconfigurations": [
            {"ID": "DS001", "Title": "Missing USER directive", "Severity": "MEDIUM", "Message": "Container should specify a non-root USER", "CauseMetadata": {"StartLine": 1, "EndLine": 25}},
            {"ID": "DS002", "Title": "Image tag uses latest", "Severity": "LOW", "Message": "Using 'latest' tag is not reproducible", "CauseMetadata": {"StartLine": 1, "EndLine": 1}},
        ]},
    ]},
    "trivy-infra.json": {"Results": [
        {"Target": "deployment.yaml", "Type": "kubernetes", "Misconfigurations": [
            {"ID": "KSV001", "Title": "Container runs as root", "Severity": "HIGH", "Message": "Container should run as non-root user", "CauseMetadata": {"StartLine": 10, "EndLine": 15}},
            {"ID": "KSV002", "Title": "Privileged container", "Severity": "CRITICAL", "Message": "Container should not run in privileged mode", "CauseMetadata": {"StartLine": 12, "EndLine": 12}},
            {"ID": "KSV003", "Title": "No resource limits", "Severity": "MEDIUM", "Message": "Container should have resource limits defined", "CauseMetadata": {"StartLine": 10, "EndLine": 25}},
            {"ID": "KSV004", "Title": "readOnlyRootFilesystem not set", "Severity": "LOW", "Message": "Container root filesystem should be read-only", "CauseMetadata": {"StartLine": 10, "EndLine": 25}},
        ]},
        {"Target": "terraform/main.tf", "Type": "terraform", "Misconfigurations": [
            {"ID": "AVD-AWS-0001", "Title": "S3 bucket without encryption", "Severity": "HIGH", "Message": "S3 bucket should have server-side encryption enabled", "CauseMetadata": {"StartLine": 20, "EndLine": 35}},
            {"ID": "AVD-AWS-0002", "Title": "Security group overly permissive", "Severity": "CRITICAL", "Message": "Security group allows 0.0.0.0/0 inbound on port 22", "CauseMetadata": {"StartLine": 50, "EndLine": 60}},
            {"ID": "AVD-AWS-0003", "Title": "EBS volume not encrypted", "Severity": "HIGH", "Message": "EBS volume encryption should be enabled", "CauseMetadata": {"StartLine": 70, "EndLine": 75}},
        ]},
    ]},
}

for fname, data in samples.items():
    w(os.path.join(SAMPLES, fname), json.dumps(data, indent=2) + "\n")

# ==============================================================
# Helper: Bruno body:json serializer
# ==============================================================
def bru_json_body(obj):
    """Serialize dict to Bruno body:json content (indented, Bruno-compliant).
    Pure template variables {{var}} are kept unquoted (for numeric IDs).
    """
    items = []
    for k, v in obj.items():
        if isinstance(v, str) and v.startswith("{{") and v.endswith("}}"):
            items.append(f'    "{k}": {v}')
        else:
            items.append(f'    "{k}": {json.dumps(v)}')
    return "  {\n" + ",\n".join(items) + "\n  }"

# ==============================================================
# Helper: generate a POST create request with post-response
# ==============================================================
def gen_post(name, seq, url, body, var_store=None, extra_headers=None):
    """Generate a .bru file for a real POST request.
    body is a dict or string. var_store is (var_name, json_path) or list of tuples.
    """
    lines = [meta(name, seq)]
    lines.append(f"post {{\n  url: {url}\n  body: json\n  auth: none\n}}\n")
    lines.append("headers {\n  Content-Type: application/json\n  X-CSRF-Token: {{csrf_token}}\n")
    if extra_headers:
        for k, v in extra_headers.items():
            lines.append(f"  {k}: {v}\n")
    lines.append("}\n")
    if isinstance(body, dict):
        body_str = bru_json_body(body)
    else:
        body_str = body
    lines.append(f"body:json {{\n{body_str}\n}}\n")
    if var_store:
        if isinstance(var_store, tuple):
            var_store = [var_store]
        lines.append("script:post-response {\n")
        for var_name, json_path in var_store:
            lines.append(f"  bru.setVar(\"{var_name}\", {json_path});\n")
        lines.append("}\n")
    w(os.path.join(OUT, f"{name}.bru"), "".join(lines))

# ==============================================================
# 10-csrf.bru (seq 1)
# ==============================================================
w(os.path.join(OUT, "10-csrf.bru"), meta("10-csrf", next_seq()) + """
get {
  url: {{base_url}}{{api_prefix}}/csrf-token
  body: none
  auth: none
}

script:post-response {
  bru.setVar("csrf_token", res.body.csrfToken);
}
""")

# ==============================================================
# 20-login.bru (seq 2)
# ==============================================================
w(os.path.join(OUT, "20-login.bru"), meta("20-login", next_seq()) + """
post {
  url: {{base_url}}{{api_prefix}}/login
  body: json
  auth: none
}
headers {
  X-CSRF-Token: {{csrf_token}}
  Content-Type: application/json
}
body:json {
  {
    "username": "admin",
    "password": "{{admin_password}}"
  }
}
script:post-response {
  bru.setVar("access_token", res.body.accessToken);
  bru.setEnvVar("access_token", res.body.accessToken);
}
""")

# ==============================================================
# 25-run-id.bru (seq 3) - generates random run_id
# ==============================================================
w(os.path.join(OUT, "25-run-id.bru"), meta("25-run-id", next_seq()) + """
get {
  url: {{base_url}}/ping
  body: none
  auth: none
}
script:pre-request {
  bru.setVar("run_id", Math.random().toString(36).substring(2, 8));
}
""")

# ==============================================================
# Groups (seq 4-6)
# ==============================================================
groups = [
    ("30-group-engineering", "engineering", "Engineering", "Engineering & Platform teams"),
    ("31-group-security",    "security",    "Security",    "Security & Compliance team"),
    ("32-group-product",     "product",     "Product",     "Product & Design teams"),
]
for fname, path, name, desc in groups:
    gen_post(fname, next_seq(),
        url="{{base_url}}{{api_prefix}}/groups",
        body={"name": f"{name} {{{{run_id}}}}", "description": desc, "path": f"{path}-{{{{run_id}}}}"},
        var_store=[(f"{path}_group_id", "res.body.id")])

# ==============================================================
# Users (seq 7-9)
# ==============================================================
users = [
    ("55-user-dev", "dev", "dev-{{run_id}}", "dev-{{run_id}}@example.com"),
    ("56-user-sec", "sec", "sec-{{run_id}}", "sec-{{run_id}}@example.com"),
    ("57-user-pm",  "pm",  "pm-{{run_id}}",  "pm-{{run_id}}@example.com"),
]
for fname, key, username, email in users:
    gen_post(fname, next_seq(),
        url="{{base_url}}{{api_prefix}}/register",
        body={"username": username, "email": email, "password": "Demo123!"},
        var_store=[(f"{key}_user_id", "res.body.id")])

# ==============================================================
# Applications (seq 10-23)
# ==============================================================
apps = [
    # (filename prefix, name, slug suffix, group var, criticality)
    ("40-app-apigateway-1",  "API Gateway 1",   "api-gateway-1",   "engineering", "critical"),
    ("41-app-apigateway-2",  "API Gateway 2",   "api-gateway-2",   "engineering", "critical"),
    ("42-app-webapp-1",      "Web Application 1", "web-app-1",     "engineering", "high"),
    ("43-app-webapp-2",      "Web Application 2", "web-app-2",     "engineering", "high"),
    ("44-app-microservice-1", "Microservice 1",   "microservice-1", "engineering", "high"),
    ("45-app-microservice-2", "Microservice 2",   "microservice-2", "engineering", "high"),
    ("46-app-vault-1",       "Vault Service 1",   "vault-service-1", "security",   "critical"),
    ("47-app-vault-2",       "Vault Service 2",   "vault-service-2", "security",   "critical"),
    ("48-app-auth-1",        "Auth Service 1",    "auth-service-1",  "security",   "high"),
    ("49-app-auth-2",        "Auth Service 2",    "auth-service-2",  "security",   "high"),
    ("50-app-customer-1",    "Customer Portal 1", "customer-portal-1", "product",  "high"),
    ("51-app-customer-2",    "Customer Portal 2", "customer-portal-2", "product",  "high"),
    ("52-app-analytics-1",   "Analytics Engine 1","analytics-1",     "product",   "medium"),
    ("53-app-analytics-2",   "Analytics Engine 2","analytics-2",     "product",   "medium"),
]
for fname, display_name, slug, group_path, crit in apps:
    var_name = slug.replace("-", "_") + "_app_id"
    gen_post(fname, next_seq(),
        url="{{base_url}}{{api_prefix}}/applications",
        body={
            "name": f"{display_name} {{{{run_id}}}}",
            "slug": f"{slug}-{{{{run_id}}}}",
            "description": f"{display_name} instance",
            "groupId": "{{" + group_path + "_group_id}}",
            "repositoryUrl": f"https://github.com/servasec/{slug}-{{{{run_id}}}}",
            "assetCriticality": crit
        },
        var_store=[(var_name, "res.body.id")])

# ==============================================================
# Versions (seq 21-34)
# ==============================================================
versions = [
    # (app_var_name, version_name, tag, branch)
    ("api_gateway_1_app_id",  "main",      "main",      "main"),
    ("api_gateway_1_app_id",  "develop",   "develop",   "develop"),
    ("api_gateway_1_app_id",  "v1.0.0",    "v1.0.0",    "main"),
    ("api_gateway_1_app_id",  "v1.1.0",    "v1.1.0",    "main"),
    ("vault_service_1_app_id", "main",     "main",      "main"),
    ("vault_service_1_app_id", "v2.1.0",   "v2.1.0",    "main"),
    ("vault_service_1_app_id", "v2.2.0-beta", "v2.2.0-beta", "main"),
    ("customer_portal_1_app_id", "main",   "main",      "main"),
    ("customer_portal_1_app_id", "staging","staging",   "staging"),
    ("customer_portal_1_app_id", "v3.0.0", "v3.0.0",    "main"),
    ("web_app_1_app_id",       "main",     "main",      "main"),
    ("web_app_1_app_id",       "develop",  "develop",   "develop"),
    ("auth_service_1_app_id",  "main",     "main",      "main"),
    ("auth_service_1_app_id",  "v1.0.0",   "v1.0.0",    "main"),
]

for i, (app_var, vname, tag, branch) in enumerate(versions):
    app_name = app_var.replace("_app_id", "")
    safe_name = vname.replace(".", "").replace("-", "_").replace("beta", "b")
    fname = f"54-version-{app_name}-{safe_name}"
    gen_post(fname, next_seq(),
        url=f"{{{{base_url}}}}{{{{api_prefix}}}}/applications/{{{{{app_var}}}}}/versions",
        body={"name": vname, "tag": tag, "branch": branch})

# ==============================================================
# Teams (seq 35-37)
# ==============================================================
teams = [
    ("68-team-platform",    "platform", "Platform Engineering", "Platform engineering & SRE team"),
    ("69-team-security",    "security", "Security Team",        "Security operations & compliance team"),
    ("70-team-product",     "product",  "Product Delivery",     "Product delivery & release management"),
]
for fname, key, name, desc in teams:
    gen_post(fname, next_seq(),
        url="{{base_url}}{{api_prefix}}/teams",
        body={"name": f"{name} {{{{run_id}}}}", "description": desc},
        var_store=[(f"{key}_team_id", "res.body.id")])

# ==============================================================
# Team membership (seq 38-40)
# ==============================================================
members = [
    ("80-team-platform-member", "platform", "dev", "member"),
    ("81-team-security-member", "security", "sec", "member"),
    ("82-team-product-member",  "product",  "pm",  "member"),
]
for fname, team_key, user_key, role in members:
    gen_post(fname, next_seq(),
        url="{{base_url}}{{api_prefix}}/teams/{{" + team_key + "_team_id}}/members",
        body={"userId": "{{" + user_key + "_user_id}}", "role": role})

# ==============================================================
# Ingests (seq 41-49) - multipart with file upload
# ==============================================================
ingests = [
    ("71-ingest-portal",    "customer_portal_1_app_id",  "semgrep-webapp.json", "main"),
    ("72-ingest-gateway",   "api_gateway_1_app_id",      "semgrep-api.json",    "main"),
    ("73-ingest-vault",     "vault_service_1_app_id",    "semgrep-secrets.json", "main"),
    ("74-ingest-webapp",    "web_app_1_app_id",          "trivy-deps.json",     "main"),
    ("75-ingest-auth",      "auth_service_1_app_id",     "trivy-infra.json",    "main"),
    ("76-ingest-webapp2",   "web_app_2_app_id",          "trivy-deps.json",     "staging"),
    ("77-ingest-analytics", "analytics_1_app_id",        "semgrep-webapp.json", "develop"),
    ("78-ingest-microservice", "microservice_1_app_id",  "semgrep-api.json",    "main"),
    ("79-ingest-portal2",   "customer_portal_2_app_id",  "trivy-infra.json",    "staging"),
]

for fname, app_var, sample, branch in ingests:
    s = next_seq()
    lines = [meta(fname, s)]
    lines.append(f"post {{\n  url: {{{{base_url}}}}{{{{api_prefix}}}}/applications/{{{{{app_var}}}}}/ingest\n  body: multipartForm\n  auth: none\n}}\n")
    lines.append("headers {\n  X-CSRF-Token: {{csrf_token}}\n}\n")
    lines.append("body:multipart-form {\n")
    lines.append(f"  file: @file(demo-setup/samples/{sample})\n")
    lines.append(f"  version: {branch}\n")
    lines.append("}\n")
    w(os.path.join(OUT, f"{fname}.bru"), "".join(lines))

# ==============================================================
# 99-findings.bru (seq 99)
# ==============================================================
w(os.path.join(OUT, "99-findings.bru"), meta("99-findings", 99) + """
get {
  url: {{base_url}}{{api_prefix}}/findings
  body: none
  auth: none
}
headers {
  Authorization: Bearer {{access_token}}
  X-CSRF-Token: {{csrf_token}}
}
""")

# ==============================================================
# Cleanup stale files
# ==============================================================
def cleanup():
    removed = []
    for root, dirs, files in os.walk(OUT):
        for f in files:
            fpath = os.path.join(root, f)
            rel = os.path.relpath(fpath, OUT)
            if rel not in generated:
                os.remove(fpath)
                removed.append(rel)
    for root, dirs, files in os.walk(OUT, topdown=False):
        for d in dirs:
            dpath = os.path.join(root, d)
            if dpath == SAMPLES:
                continue
            try:
                if not os.listdir(dpath):
                    os.rmdir(dpath)
                    removed.append(os.path.relpath(dpath, OUT) + "/")
            except OSError:
                pass
    return removed

removed = cleanup()

# Summary
bru_files = [f for f in generated if f.endswith(".bru")]
print(f"Generated {len(generated)} items in {OUT}/:")
for f in sorted(generated):
    print(f"  {f}")
print(f"\n.bru files: {len(bru_files)} | Sample files: {len([f for f in generated if 'samples/' in f])}")
if removed:
    print(f"\nRemoved stale: {removed}")
