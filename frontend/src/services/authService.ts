import axios, { AxiosResponse } from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  role: string;
  status?: string;
  created_at?: string;
  last_access?: string | null;
}

export interface LoginResponse {
  access_token: string;
  expires_in: number;
  token_type: string;
  user: User;
}

export interface RefreshResponse {
  access_token: string;
  expires_in: number;
  token_type: string;
}

export interface ApiResponse<T> {
  data?: T;
  error?: string;
  message?: string;
}

class AuthService {
  private baseURL = API_BASE_URL;

  constructor() {
    this.setupAxiosInterceptors();
  }

  private setupAxiosInterceptors() {
    // Request interceptor to add auth token
    axios.interceptors.request.use(
      (config) => {
        const token = this.getToken();
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // Response interceptor to handle token expiration
    axios.interceptors.response.use(
      (response) => response,
      async (error) => {
        const originalRequest = error.config;

        if (error.response?.status === 401 && !originalRequest._retry) {
          originalRequest._retry = true;

          try {
            await this.refreshToken();
            // Retry the original request
            const token = this.getToken();
            originalRequest.headers.Authorization = `Bearer ${token}`;
            return axios(originalRequest);
          } catch (refreshError) {
            // Refresh failed, redirect to login
            this.logout();
            window.location.href = '/login';
            return Promise.reject(refreshError);
          }
        }

        return Promise.reject(error);
      }
    );
  }

  async login(email: string, password: string): Promise<LoginResponse> {
    try {
      const response: AxiosResponse<LoginResponse> = await axios.post(
        `${this.baseURL}/auth/login`,
        { email, password }
      );

      const { access_token, user } = response.data;
      
      // Store token and user data
      localStorage.setItem('access_token', access_token);
      localStorage.setItem('user', JSON.stringify(user));

      return response.data;
    } catch (error: any) {
      throw new Error(error.response?.data?.error || 'Login failed');
    }
  }

  async logout(): Promise<void> {
    try {
      // Call logout endpoint
      await axios.post(`${this.baseURL}/auth/logout`);
    } catch (error) {
      // Even if logout API call fails, we should still clear local storage
      console.warn('Logout API call failed:', error);
    } finally {
      // Always clear local storage
      localStorage.removeItem('access_token');
      localStorage.removeItem('user');
    }
  }

  async refreshToken(): Promise<void> {
    try {
      const response: AxiosResponse<RefreshResponse> = await axios.post(
        `${this.baseURL}/auth/refresh`
      );

      const { access_token } = response.data;
      localStorage.setItem('access_token', access_token);
    } catch (error: any) {
      // Clear tokens if refresh fails
      localStorage.removeItem('access_token');
      localStorage.removeItem('user');
      throw new Error(error.response?.data?.error || 'Token refresh failed');
    }
  }

  async getCurrentUser(): Promise<User> {
    try {
      const response: AxiosResponse<ApiResponse<User>> = await axios.get(
        `${this.baseURL}/auth/me`
      );

      if (response.data.data) {
        // Update stored user data
        localStorage.setItem('user', JSON.stringify(response.data.data));
        return response.data.data;
      } else {
        throw new Error('No user data received');
      }
    } catch (error: any) {
      throw new Error(error.response?.data?.error || 'Failed to get user data');
    }
  }

  getToken(): string | null {
    return localStorage.getItem('access_token');
  }

  getStoredUser(): User | null {
    const userData = localStorage.getItem('user');
    if (userData) {
      try {
        return JSON.parse(userData);
      } catch (error) {
        console.error('Failed to parse stored user data:', error);
        localStorage.removeItem('user');
        return null;
      }
    }
    return null;
  }

  isAuthenticated(): boolean {
    const token = this.getToken();
    if (!token) return false;

    try {
      // Simple check - decode JWT payload to check expiration
      const payload = JSON.parse(atob(token.split('.')[1]));
      const now = Date.now() / 1000;
      
      return payload.exp > now;
    } catch (error) {
      console.error('Invalid token format:', error);
      return false;
    }
  }

  isTokenExpired(): boolean {
    return !this.isAuthenticated();
  }
}

export const authService = new AuthService();