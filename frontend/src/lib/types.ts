export interface User {
  id: number;
  username: string;
  email: string;
  role: string;
  banned: boolean;
  avatarUrl?: string;
  oauthProvider?: string;
  createdAt: string;
}

export interface Application {
  id: number;
  name: string;
  description: string;
  slug: string;
  groupId: number;
  repositoryUrl: string;
  apiToken?: string;
  versions?: ApplicationVersion[];
  defaultVersion?: ApplicationVersion | null;
  assetCriticality?: string;
  createdAt: string;
  updatedAt: string;
}

export interface ScannerType {
  id: number;
  name: string;
  description: string;
  parser: string;
}

export interface ApplicationVersion {
  id: number;
  applicationId: number;
  name: string;
  branch: string;
  tag: string;
  isDefault: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface Finding {
  id: number;
  scanId: number;
  applicationVersionId: number;
  applicationVersion: ApplicationVersion | null;
  scannerTypeId: number;
  scannerType: ScannerType | null;
  ruleId: string;
  title: string;
  severity: string;
  description: string;
  filePath: string;
  lineStart: number | null;
  lineEnd: number | null;
  cweId: string;
  remediation: string;
  status: string;
  assignedTo: number | null;
  assignedToUser: User | null;
  dueDate: string | null;
  reviewedBy: number | null;
  reviewedByUser: User | null;
  fixedAt: string | null;
  riskScore: number | null;
  epssScore: number | null;
  createdAt: string;
  updatedAt: string;
}

export interface ScanItem {
  id: number;
  applicationVersionId: number;
  applicationVersion: ApplicationVersion | null;
  scannerTypeId: number;
  scannerType: ScannerType | null;
  status: string;
  startedAt: string | null;
  completedAt: string | null;
  findingsCount: number;
  createdAt: string;
}

export interface TopRiskyFinding {
  id: number;
  title: string;
  riskScore: number;
  severity: string;
}

export interface RiskDistribution {
  label: string;
  count: number;
}

export interface DashboardStats {
  totalUsers: number;
  adminUsers: number;
  memberUsers: number;
  bannedUsers: number;
  registeredAt: string;
  totalFindings: number;
  bySeverity: { severity: string; count: number }[];
  byScanner: { scannerType: string; count: number }[];
  byStatus: { status: string; count: number }[];
  recentScans: number;
  topFindings: { ruleId: string; title: string; count: number }[];
  myOpenFindings: number;
  overdueFindings: number;
  avgRiskScore?: number;
  topRiskyFindings?: TopRiskyFinding[];
  riskDistribution?: RiskDistribution[];
}

export interface Comment {
  id: number;
  findingId: number;
  userId: number;
  user: User;
  body: string;
  createdAt: string;
}

export interface Webhook {
  id: number;
  applicationId: number;
  url: string;
  secret: string;
  events: string;
  isActive: boolean;
  createdAt: string;
}

export interface CompareResult {
  from: ApplicationVersion;
  to: ApplicationVersion;
  fixed: Finding[];
  new: Finding[];
  stillPresent: Finding[];
}

export interface Permission {
  userId: number;
  user: User;
  resource: string;
  action: string;
  createdAt: string;
}
