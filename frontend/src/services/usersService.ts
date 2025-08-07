import axios, { AxiosResponse } from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  avatar?: string | null;
  language: string;
  theme: string;
  status: string;
  role_id: string;
  role_name: string;
  last_access?: string | null;
  last_page?: string | null;
  provider: string;
  external_identifier?: string | null;
  email_notifications: boolean;
  tags?: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateUserRequest {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  avatar?: string | null;
  language?: string;
  theme?: string;
  status?: string;
  role_id: string;
  email_notifications?: boolean;
}

export interface UpdateUserRequest {
  email?: string;
  password?: string;
  first_name?: string;
  last_name?: string;
  avatar?: string | null;
  language?: string;
  theme?: string;
  status?: string;
  role_id?: string;
  email_notifications?: boolean;
}

export interface UsersResponse {
  data: User[];
  meta: {
    page: number;
    limit: number;
    total: number;
  };
}

export interface UserResponse {
  data: User;
}

export interface ApiResponse<T> {
  data?: T;
  error?: string;
  message?: string;
}

class UsersService {
  private baseURL = `${API_BASE_URL}/users`;

  async getUsers(page: number = 1, limit: number = 50): Promise<UsersResponse> {
    try {
      const response: AxiosResponse<UsersResponse> = await axios.get(
        `${this.baseURL}?page=${page}&limit=${limit}`
      );
      return response.data;
    } catch (error: any) {
      throw new Error(error.response?.data?.error || 'Failed to fetch users');
    }
  }

  async getUser(id: string): Promise<User> {
    try {
      const response: AxiosResponse<UserResponse> = await axios.get(
        `${this.baseURL}/${id}`
      );
      if (response.data.data) {
        return response.data.data;
      } else {
        throw new Error('No user data received');
      }
    } catch (error: any) {
      throw new Error(error.response?.data?.error || 'Failed to fetch user');
    }
  }

  async createUser(userData: CreateUserRequest): Promise<User> {
    try {
      const response: AxiosResponse<UserResponse> = await axios.post(
        this.baseURL,
        userData
      );
      if (response.data.data) {
        return response.data.data;
      } else {
        throw new Error('No user data received');
      }
    } catch (error: any) {
      throw new Error(error.response?.data?.error || 'Failed to create user');
    }
  }

  async updateUser(id: string, userData: UpdateUserRequest): Promise<User> {
    try {
      const response: AxiosResponse<UserResponse> = await axios.patch(
        `${this.baseURL}/${id}`,
        userData
      );
      if (response.data.data) {
        return response.data.data;
      } else {
        throw new Error('No user data received');
      }
    } catch (error: any) {
      throw new Error(error.response?.data?.error || 'Failed to update user');
    }
  }

  async deleteUser(id: string): Promise<void> {
    try {
      await axios.delete(`${this.baseURL}/${id}`);
    } catch (error: any) {
      throw new Error(error.response?.data?.error || 'Failed to delete user');
    }
  }

  async getMe(): Promise<User> {
    try {
      const response: AxiosResponse<UserResponse> = await axios.get(
        `${this.baseURL}/me`
      );
      if (response.data.data) {
        return response.data.data;
      } else {
        throw new Error('No user data received');
      }
    } catch (error: any) {
      throw new Error(error.response?.data?.error || 'Failed to fetch user profile');
    }
  }

  async updateMe(userData: UpdateUserRequest): Promise<User> {
    try {
      const response: AxiosResponse<UserResponse> = await axios.patch(
        `${this.baseURL}/me`,
        userData
      );
      if (response.data.data) {
        return response.data.data;
      } else {
        throw new Error('No user data received');
      }
    } catch (error: any) {
      throw new Error(error.response?.data?.error || 'Failed to update user profile');
    }
  }
}

export const usersService = new UsersService();
