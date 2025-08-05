import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

export interface SystemStats {
  total_users: number;
  active_users: number;
  total_roles: number;
  total_collections: number;
  total_sessions: number;
}

export interface UserSummary {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  role_name: string;
  status: string;
  created_at: string;
  last_access: string | null;
}

export interface UserInsights {
  users_by_status: Record<string, number>;
  users_by_role: Record<string, number>;
  new_users_this_week: number;
  new_users_this_month: number;
  recent_registrations: UserSummary[];
  most_active_users: UserSummary[];
}

export interface CollectionSummary {
  collection: string;
  icon: string | null;
  note: string | null;
  hidden: boolean;
  singleton: boolean;
  item_count: number;
  created_at: string;
}

export interface CollectionMetrics {
  total_collections: number;
  collections_by_type: Record<string, number>;
  recent_collections: CollectionSummary[];
  most_active_collections: CollectionSummary[];
}

export interface ActivityItem {
  id: string;
  action: string;
  user_id: string | null;
  user_name: string | null;
  collection: string | null;
  item: string | null;
  comment: string | null;
  timestamp: string;
  ip: string | null;
}

export interface SystemHealth {
  database_connected: boolean;
  server_uptime: string;
  version: string;
  last_backup: string | null;
}

export interface DashboardOverview {
  system_stats: SystemStats;
  user_insights: UserInsights;
  collection_metrics: CollectionMetrics;
  recent_activity: ActivityItem[];
  system_health: SystemHealth;
}

export interface ApiResponse<T> {
  data: T;
  error?: string;
}

class DashboardService {
  private baseURL = API_BASE_URL;

  // Get complete dashboard overview
  async getDashboardOverview(): Promise<DashboardOverview> {
    const response = await axios.get<ApiResponse<DashboardOverview>>(
      `${this.baseURL}/dashboard`
    );
    return response.data.data;
  }

  // Get basic system statistics
  async getSystemStats(): Promise<SystemStats> {
    const response = await axios.get<ApiResponse<SystemStats>>(
      `${this.baseURL}/dashboard/stats`
    );
    return response.data.data;
  }

  // Get recent activity with optional limit
  async getRecentActivity(limit?: number): Promise<ActivityItem[]> {
    const params = limit ? { limit } : {};
    const response = await axios.get<ApiResponse<ActivityItem[]>>(
      `${this.baseURL}/dashboard/activity`,
      { params }
    );
    return response.data.data;
  }

  // Get user insights
  async getUserInsights(): Promise<UserInsights> {
    const response = await axios.get<ApiResponse<UserInsights>>(
      `${this.baseURL}/dashboard/users`
    );
    return response.data.data;
  }

  // Get collection metrics
  async getCollectionInsights(): Promise<CollectionMetrics> {
    const response = await axios.get<ApiResponse<CollectionMetrics>>(
      `${this.baseURL}/dashboard/collections`
    );
    return response.data.data;
  }
}

const dashboardService = new DashboardService();
export default dashboardService;
