#!/usr/bin/env bash
# PiSpec Flow — EFS CSI driver and secrets setup script
# Run this once against your EKS cluster before deploying manifests.

set -euo pipefail

# ── Configuration ──────────────────────────────────────────────────────────────
CLUSTER_NAME="${CLUSTER_NAME:?Set CLUSTER_NAME}"
AWS_REGION="${AWS_REGION:-us-east-1}"
EFS_FILESYSTEM_ID="${EFS_FILESYSTEM_ID:?Set EFS_FILESYSTEM_ID (e.g. fs-0abc1234)}"
AWS_ACCOUNT_ID="${AWS_ACCOUNT_ID:?Set AWS_ACCOUNT_ID}"

# GitHub PATs — one per agent for least-privilege scoping
# Required scopes: contents:read, metadata:read, pull_requests:read (adjust per agent)
GITHUB_PAT_PLANNER="${GITHUB_PAT_PLANNER:?Set GITHUB_PAT_PLANNER}"
GITHUB_PAT_REVIEWER="${GITHUB_PAT_REVIEWER:?Set GITHUB_PAT_REVIEWER}"
GITHUB_PAT_SECURITY="${GITHUB_PAT_SECURITY:?Set GITHUB_PAT_SECURITY}"

# API keys
ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:?Set ANTHROPIC_API_KEY}"
OPENAI_API_KEY="${OPENAI_API_KEY:?Set OPENAI_API_KEY}"

IRSA_ROLE_NAME="${IRSA_ROLE_NAME:-pispec-flow-irsa}"

# ── Helpers ────────────────────────────────────────────────────────────────────
log() { echo "[$(date -u +%H:%M:%S)] $*"; }
require() { command -v "$1" &>/dev/null || { echo "ERROR: $1 not found in PATH"; exit 1; }; }

require kubectl
require aws
require helm
require eksctl

# ── 1. Install EFS CSI Driver via Helm ─────────────────────────────────────────
log "Installing AWS EFS CSI driver..."
helm repo add aws-efs-csi-driver https://kubernetes-sigs.github.io/aws-efs-csi-driver/ 2>/dev/null || true
helm repo update

helm upgrade --install aws-efs-csi-driver aws-efs-csi-driver/aws-efs-csi-driver \
  --namespace kube-system \
  --set controller.serviceAccount.annotations."eks\.amazonaws\.com/role-arn"="arn:aws:iam::${AWS_ACCOUNT_ID}:role/AmazonEKS_EFS_CSI_DriverRole" \
  --wait

log "EFS CSI driver installed."

# ── 2. Create IRSA role for EFS CSI Driver ────────────────────────────────────
log "Creating IRSA role for EFS CSI driver..."
eksctl create iamserviceaccount \
  --cluster "${CLUSTER_NAME}" \
  --region "${AWS_REGION}" \
  --namespace kube-system \
  --name efs-csi-controller-sa \
  --attach-policy-arn arn:aws:iam::aws:policy/service-role/AmazonEFSCSIDriverPolicy \
  --approve \
  --override-existing-serviceaccounts 2>/dev/null || log "IRSA for EFS CSI may already exist, continuing."

# ── 3. Update EFS StorageClass with filesystem ID ────────────────────────────
log "Patching k8s/01-efs-storage.yaml with EFS filesystem ID..."
sed -i "s/{{EFS_FILESYSTEM_ID}}/${EFS_FILESYSTEM_ID}/g" k8s/01-efs-storage.yaml

# ── 4. Update ServiceAccount with IRSA role ───────────────────────────────────
log "Patching k8s/04-serviceaccounts.yaml with AWS account and IRSA role..."
sed -i "s/{{AWS_ACCOUNT_ID}}/${AWS_ACCOUNT_ID}/g" k8s/04-serviceaccounts.yaml
sed -i "s/{{IRSA_ROLE_NAME}}/${IRSA_ROLE_NAME}/g" k8s/04-serviceaccounts.yaml

# ── 5. Apply namespace first ──────────────────────────────────────────────────
log "Creating pi-agents namespace..."
kubectl apply -f k8s/00-namespace.yaml

# ── 6. Create Kubernetes Secrets ──────────────────────────────────────────────
log "Creating GitHub PAT secrets..."

kubectl create secret generic github-pat-planner \
  --namespace pi-agents \
  --from-literal=GITHUB_PAT="${GITHUB_PAT_PLANNER}" \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl create secret generic github-pat-reviewer \
  --namespace pi-agents \
  --from-literal=GITHUB_PAT="${GITHUB_PAT_REVIEWER}" \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl create secret generic github-pat-security \
  --namespace pi-agents \
  --from-literal=GITHUB_PAT="${GITHUB_PAT_SECURITY}" \
  --dry-run=client -o yaml | kubectl apply -f -

log "Creating API key secrets..."

kubectl create secret generic anthropic-api-key \
  --namespace pi-agents \
  --from-literal=ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY}" \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl create secret generic openai-api-key \
  --namespace pi-agents \
  --from-literal=OPENAI_API_KEY="${OPENAI_API_KEY}" \
  --dry-run=client -o yaml | kubectl apply -f -

log "All secrets created."

# ── 7. Apply remaining manifests in order ────────────────────────────────────
log "Applying Kubernetes manifests..."

for f in \
  k8s/01-efs-storage.yaml \
  k8s/02-secrets.yaml \
  k8s/03-configmaps.yaml \
  k8s/04-serviceaccounts.yaml \
  k8s/05-planner-deployment.yaml \
  k8s/06-reviewer-deployment.yaml \
  k8s/07-security-deployment.yaml \
  k8s/08-services.yaml \
  k8s/09-viewer-deployment.yaml
do
  log "  Applying ${f}..."
  kubectl apply -f "${f}"
done

# ── 8. Wait for deployments ───────────────────────────────────────────────────
log "Waiting for deployments to become ready (timeout 5m)..."
for deploy in pi-planner pi-reviewer pi-security specs-viewer; do
  kubectl rollout status deployment/"${deploy}" -n pi-agents --timeout=300s
done

# ── 9. Print summary ──────────────────────────────────────────────────────────
log ""
log "════════════════════════════════════════════════════"
log "  PiSpec Flow deployed successfully!"
log ""
log "  Agents:"
PLANNER_IP=$(kubectl get svc pi-planner -n pi-agents -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "pending")
VIEWER_URL=$(kubectl get svc specs-viewer -n pi-agents -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || echo "use kubectl port-forward")
log "    Planner  : ws://pi-planner.pi-agents.svc.cluster.local:8787"
log "    Reviewer : ws://pi-reviewer.pi-agents.svc.cluster.local:8788"
log "    Security : ws://pi-security.pi-agents.svc.cluster.local:8789"
log ""
log "  Specs viewer : http://${VIEWER_URL}"
log ""
log "  Start a feature:"
log "    acpx-orchestrator run --workflow workflows/pispec-orchestrator.yaml \\"
log "      --var feature=my-feature-name"
log ""
log "  PAT rotation reminder: rotate GitHub PATs every 90 days."
log "════════════════════════════════════════════════════"
