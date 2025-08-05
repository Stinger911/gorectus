import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

export interface Collection {
  collection: string;
  icon?: string;
  note?: string;
  display_template?: string;
  hidden: boolean;
  singleton: boolean;
  translations?: any;
  archive_field?: string;
  archive_app_filter: boolean;
  archive_value?: string;
  unarchive_value?: string;
  sort_field?: string;
  accountability: string;
  color?: string;
  item_duplication_fields?: any;
  sort?: number;
  group?: string;
  collapse: string;
  preview_url?: string;
  versioning: boolean;
  created_at: string;
  updated_at: string;
}

export interface Field {
  id: string;
  collection: string;
  field: string;
  special?: string[];
  interface?: string;
  options?: any;
  display?: string;
  display_options?: any;
  readonly: boolean;
  hidden: boolean;
  sort?: number;
  width: string;
  translations?: any;
  note?: string;
  conditions?: any;
  required: boolean;
  group?: string;
  validation?: any;
  validation_message?: string;
  created_at?: string;  // Made optional for creation
  updated_at?: string;  // Made optional for creation
}

export interface CollectionWithFields {
  collection: Collection;
  fields: Field[];
}

export interface CreateCollectionRequest {
  collection: string;
  icon?: string;
  note?: string;
  display_template?: string;
  hidden?: boolean;
  singleton?: boolean;
  translations?: any;
  archive_field?: string;
  archive_app_filter?: boolean;
  archive_value?: string;
  unarchive_value?: string;
  sort_field?: string;
  accountability?: string;
  color?: string;
  item_duplication_fields?: any;
  sort?: number;
  group?: string;
  collapse?: string;
  preview_url?: string;
  versioning?: boolean;
  fields: Field[];
}

export interface UpdateCollectionRequest {
  icon?: string;
  note?: string;
  display_template?: string;
  hidden?: boolean;
  singleton?: boolean;
  translations?: any;
  archive_field?: string;
  archive_app_filter?: boolean;
  archive_value?: string;
  unarchive_value?: string;
  sort_field?: string;
  accountability?: string;
  color?: string;
  item_duplication_fields?: any;
  sort?: number;
  group?: string;
  collapse?: string;
  preview_url?: string;
  versioning?: boolean;
}

export interface CollectionsListResponse {
  data: Collection[];
  meta: {
    page: number;
    limit: number;
    total: number;
  };
}

export interface ApiResponse<T> {
  data: T;
  error?: string;
}

class CollectionsService {
  private baseURL = API_BASE_URL;

  private getAuthHeaders() {
    const token = localStorage.getItem('access_token');
    return {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    };
  }

  async getCollections(page: number = 1, limit: number = 50): Promise<CollectionsListResponse> {
    try {
      const response = await axios.get(`${this.baseURL}/collections`, {
        headers: this.getAuthHeaders(),
        params: { page, limit },
      });
      return response.data;
    } catch (error) {
      console.error('Error fetching collections:', error);
      throw error;
    }
  }

  async getCollection(collectionName: string): Promise<CollectionWithFields> {
    try {
      const response = await axios.get(`${this.baseURL}/collections/${collectionName}`, {
        headers: this.getAuthHeaders(),
      });
      return response.data.data;
    } catch (error) {
      console.error('Error fetching collection:', error);
      throw error;
    }
  }

  async createCollection(collectionData: CreateCollectionRequest): Promise<Collection> {
    try {
      const response = await axios.post(`${this.baseURL}/collections`, collectionData, {
        headers: this.getAuthHeaders(),
      });
      return response.data.data;
    } catch (error) {
      console.error('Error creating collection:', error);
      throw error;
    }
  }

  async updateCollection(collectionName: string, collectionData: UpdateCollectionRequest): Promise<Collection> {
    try {
      const response = await axios.patch(`${this.baseURL}/collections/${collectionName}`, collectionData, {
        headers: this.getAuthHeaders(),
      });
      return response.data.data;
    } catch (error) {
      console.error('Error updating collection:', error);
      throw error;
    }
  }

  async deleteCollection(collectionName: string): Promise<void> {
    try {
      await axios.delete(`${this.baseURL}/collections/${collectionName}`, {
        headers: this.getAuthHeaders(),
      });
    } catch (error) {
      console.error('Error deleting collection:', error);
      throw error;
    }
  }
}

const collectionsService = new CollectionsService();
export default collectionsService;
