<div align="center">
  <h1>trix</h1>
  <p><strong>Kubernetes Security Scanner with AI-Powered Triage</strong></p>

  <p>
    Query vulnerabilities, compliance issues, and security posture from your cluster.<br>
    Use AI to investigate findings and get actionable remediation advice.
  </p>

  <p>
    <a href="https://github.com/davealtena/trix/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License"></a>
    <a href="https://golang.org/"><img src="https://img.shields.io/badge/Go-1.21+-00ADD8.svg" alt="Go Version"></a>
  </p>

  <p>
    <a href="#how-it-works"><strong>How it Works</strong></a> |
    <a href="#installation"><strong>Installation</strong></a> |
    <a href="#usage"><strong>Usage</strong></a> |
    <a href="#ai-powered-investigation"><strong>AI Investigation</strong></a>
  </p>
</div>

---

## How it Works

trix is a local CLI tool that queries security data from your Kubernetes cluster via kubeconfig. [Trivy Operator](https://aquasecurity.github.io/trivy-operator/) runs in-cluster and scans your workloads - trix reads those results and makes them actionable.

### What trix finds

| Finding Type | Description |
|--------------|-------------|
| **Vulnerabilities** | CVEs in container images with CVSS scores |
| **Compliance** | Misconfigurations and policy violations |
| **RBAC Issues** | Overly permissive roles and bindings |
| **Exposed Secrets** | Secrets found in container images |
| **NetworkPolicy Gaps** | Pods without network protection |
| **Software Inventory** | SBOM data for all images |

## Installation

### Prerequisites

- Access to a Kubernetes cluster
- [Trivy Operator](https://aquasecurity.github.io/trivy-operator/) installed in your cluster

<details>
<summary>Install Trivy Operator (if not already installed)</summary>

```bash
helm repo add aqua https://aquasecurity.github.io/helm-charts/
helm repo update
helm install trivy-operator aqua/trivy-operator \
  --namespace trivy-system \
  --create-namespace
```

</details>

### Install trix

**From source:**

```bash
git clone https://github.com/davealtena/trix.git
cd trix
go build -o trix .
sudo mv trix /usr/local/bin/
```

**Verify installation:**

```bash
trix version
trix status  # Check Trivy Operator connection
```

## Usage

### Query Security Findings

```bash
# View all findings across namespaces
trix query findings -A

# Summary with severity breakdown
trix query summary -A

# Filter by namespace
trix query findings -n production

# JSON output for automation
trix query findings -A -o json
```

### Check NetworkPolicy Coverage

```bash
trix query network -A
```

### Search Software Inventory (SBOM)

```bash
# List all images and components
trix query sbom -A

# Search for specific packages (e.g., log4j)
trix query sbom -A --package log4j
```

### Trigger Rescans

```bash
# Rescan vulnerabilities in a namespace
trix scan vulns -n default

# Rescan everything (with confirmation skip)
trix scan all -A -y
```

### Example Output

```
$ trix query summary -A

Security Findings Summary
=========================

Total Findings: 884

By Severity:
  CRITICAL:  12
  HIGH:      45
  MEDIUM:    234
  LOW:       593

By Type:
  vulnerability: 763
  compliance:    47
  rbac:          11

Top Affected Resources:
  kube-system/etcd-control-plane - 112 findings
  kube-system/kube-apiserver - 89 findings
```

## AI-Powered Investigation

Use natural language to investigate your cluster's security posture. trix uses AI to query findings, analyze RBAC, and provide actionable remediation steps.

**Bring Your Own Key (BYOK):** You provide your own LLM API key. Your data stays between you and your LLM provider.

### Setup

```bash
export ANTHROPIC_API_KEY=your-key-here
```

### Ask Questions

```bash
# Single question
trix ask "What are the top 5 security risks in my cluster?"

# Interactive mode for follow-up questions
trix ask "What critical vulnerabilities do I have?" -i
```

### Interactive Mode

```
$ trix ask "What critical vulnerabilities are in my cluster?" -i
Investigating...
  → trix query summary -A
  → trix query findings --severity=CRITICAL
  [tokens: 2477 in, 357 out | total: 5548 in, 507 out]

## Critical Security Issues Summary
Your cluster has 20 critical vulnerabilities across 8 workloads...

> How do I fix CVE-2024-45337?
Investigating...
  → trix finding detail CVE-2024-45337
  [tokens: 3200 in, 450 out | total: 8748 in, 957 out]

## How to Patch CVE-2024-45337
Update the golang.org/x/crypto package to version 0.31.0 or later...
```

**Commands in interactive mode:**
- Type your question and press Enter
- `clear` - Reset conversation context
- `exit` or `quit` - Exit

### Supported LLM Providers

| Provider | Status | Environment Variable |
|----------|--------|---------------------|
| Anthropic (Claude) | Supported | `ANTHROPIC_API_KEY` |
| OpenAI | Planned | - |
| Ollama (local) | Planned | - |

Use `--model` to specify a model:

```bash
trix ask "..." --model claude-sonnet-4-20250514
```

## Data Sources

| Source | Status | Notes |
|--------|--------|-------|
| Trivy Operator | Supported | Vulnerabilities, compliance, RBAC, secrets, SBOM |
| Kubernetes | Supported | NetworkPolicy coverage analysis |
| Kyverno | Planned | Policy violations |
| Falco | Planned | Runtime security events |

## License

Distributed under the Apache 2.0 License. See [LICENSE](LICENSE) for more information.
