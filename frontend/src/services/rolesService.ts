import axios, { AxiosResponse } from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

export interface Role {
  id: string;
  name: string;
  icon: string;
  description?: string | null;
  ip_access: string[];
  enforce_tfa: boolean;
  admin_access: boolean;
  app_access: boolean;
  created_at: string;
  updated_at: string;
}

export interface RolesResponse {
  data: Role[];
  meta: {
    page: number;
    limit: number;
    total: number;
  };
}

export interface RoleResponse {
  data: Role;
}

export interface ApiResponse<T> {
  data?: T;
  error?: string;
  message?: string;
}

class RolesService {
  private baseURL = `${API_BASE_URL}/roles`;

  async getRoles(page: number = 1, limit: number = 50): Promise<RolesResponse> {
    try {
      const response: AxiosResponse<RolesResponse> = await axios.get(
        `${this.baseURL}?page=${page}&limit=${limit}`
      );
      return response.data;
    } catch (error: any) {
      throw new Error(error.response?.data?.error || 'Failed to fetch roles');
    }
  }

  async getRole(id: string): Promise<Role> {
    try {
      const response: AxiosResponse<RoleResponse> = await axios.get(
        `${this.baseURL}/${id}`
      );
      if (response.data.data) {
        return response.data.data;
      } else {
        throw new Error('No role data received');
      }
    } catch (error: any) {
      throw new Error(error.response?.data?.error || 'Failed to fetch role');
    }
  }
}

export const rolesService = new RolesService();
