import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

export interface SystemPreferences {
  // General Settings
  site_name: string;
  site_description: string;
  allow_registration: boolean;
  maintenance_mode: boolean;

  // Database Settings (read-only for security)
  database_host: string;
  database_port: string;
  database_name: string;
  database_user: string;

  // Email Settings
  smtp_host: string;
  smtp_port: string;
  smtp_user: string;
  smtp_from_email: string;
  email_enabled: boolean;

  // Security Settings
  session_timeout: number;
  password_min_length: number;
  require_two_factor: boolean;
  jwt_secret_exists: boolean; // Don't expose actual secret

  // Metadata
  updated_at: string;
  updated_by: string;
}

export interface UpdateSettingsRequest {
  // General Settings
  site_name?: string;
  site_description?: string;
  allow_registration?: boolean;
  maintenance_mode?: boolean;

  // Email Settings
  smtp_host?: string;
  smtp_port?: string;
  smtp_user?: string;
  smtp_from_email?: string;
  email_enabled?: boolean;

  // Security Settings
  jwt_secret?: string; // Only for updates
  session_timeout?: number;
  password_min_length?: number;
  require_two_factor?: boolean;
}

export interface SettingsResponse {
  data: SystemPreferences;
  error?: string;
}

export interface ApiResponse<T> {
  data?: T;
  error?: string;
  message?: string;
}

class SettingsService {
  private baseURL = `${API_BASE_URL}/settings`;

  private getAuthHeaders() {
    const token = localStorage.getItem('access_token');
    return {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    };
  }

  async getSettings(): Promise<SystemPreferences> {
    try {
      const response = await axios.get<SettingsResponse>(this.baseURL, {
        headers: this.getAuthHeaders(),
      });
      return response.data.data;
    } catch (error) {
      console.error('Error fetching settings:', error);
      throw error;
    }
  }

  async updateSettings(settingsData: UpdateSettingsRequest): Promise<SystemPreferences> {
    try {
      const response = await axios.patch<SettingsResponse>(this.baseURL, settingsData, {
        headers: this.getAuthHeaders(),
      });
      return response.data.data;
    } catch (error) {
      console.error('Error updating settings:', error);
      throw error;
    }
  }

  async testDatabaseConnection(): Promise<void> {
    try {
      await axios.post<ApiResponse<any>>(`${this.baseURL}/test-connection`, {}, {
        headers: this.getAuthHeaders(),
      });
    } catch (error) {
      console.error('Error testing database connection:', error);
      throw error;
    }
  }

  async testEmailConfiguration(): Promise<void> {
    try {
      await axios.post<ApiResponse<any>>(`${this.baseURL}/test-email`, {}, {
        headers: this.getAuthHeaders(),
      });
    } catch (error) {
      console.error('Error testing email configuration:', error);
      throw error;
    }
  }
}

const settingsService = new SettingsService();
export default settingsService;
