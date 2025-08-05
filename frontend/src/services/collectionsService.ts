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

export interface FieldSchema {
  data_type: string;          // varchar, integer, boolean, text, json, uuid, timestamp, etc.
  max_length?: number;        // For varchar
  is_nullable?: boolean;      // Whether column allows NULL
  default_value?: any;        // Default value for column
  is_unique?: boolean;        // Whether column should be unique
  is_primary_key?: boolean;   // Whether column is primary key
  foreign_table?: string;     // For foreign key relationships
  foreign_column?: string;    // For foreign key relationships
}

export interface CreateFieldRequest {
  field: string;
  special?: string[];
  interface?: string;
  options?: any;
  display?: string;
  display_options?: any;
  readonly?: boolean;
  hidden?: boolean;
  sort?: number;
  width?: string;
  translations?: any;
  note?: string;
  conditions?: any;
  required?: boolean;
  group?: string;
  validation?: any;
  validation_message?: string;
  schema?: FieldSchema;  // For creating database columns
}

export interface UpdateFieldRequest {
  special?: string[];
  interface?: string;
  options?: any;
  display?: string;
  display_options?: any;
  readonly?: boolean;
  hidden?: boolean;
  sort?: number;
  width?: string;
  translations?: any;
  note?: string;
  conditions?: any;
  required?: boolean;
  group?: string;
  validation?: any;
  validation_message?: string;
  schema?: FieldSchema;  // For altering database columns
}

export interface FieldsListResponse {
  data: Field[];
  meta: {
    page: number;
    limit: number;
    total: number;
  };
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

  // Field Management Methods

  async getFields(page: number = 1, limit: number = 50, collection?: string): Promise<FieldsListResponse> {
    try {
      const params: any = { page, limit };
      if (collection) {
        params.collection = collection;
      }
      
      const response = await axios.get(`${this.baseURL}/fields`, {
        headers: this.getAuthHeaders(),
        params,
      });
      return response.data;
    } catch (error) {
      console.error('Error fetching fields:', error);
      throw error;
    }
  }

  async getFieldsByCollection(collectionName: string): Promise<Field[]> {
    try {
      const response = await axios.get(`${this.baseURL}/fields/${collectionName}`, {
        headers: this.getAuthHeaders(),
      });
      return response.data.data;
    } catch (error) {
      console.error('Error fetching fields for collection:', error);
      throw error;
    }
  }

  async getField(collectionName: string, fieldName: string): Promise<Field> {
    try {
      const response = await axios.get(`${this.baseURL}/fields/${collectionName}/${fieldName}`, {
        headers: this.getAuthHeaders(),
      });
      return response.data.data;
    } catch (error) {
      console.error('Error fetching field:', error);
      throw error;
    }
  }

  async createField(collectionName: string, fieldData: CreateFieldRequest): Promise<Field> {
    try {
      const response = await axios.post(`${this.baseURL}/fields/${collectionName}`, fieldData, {
        headers: this.getAuthHeaders(),
      });
      return response.data.data;
    } catch (error) {
      console.error('Error creating field:', error);
      throw error;
    }
  }

  async updateField(collectionName: string, fieldName: string, fieldData: UpdateFieldRequest): Promise<Field> {
    try {
      const response = await axios.patch(`${this.baseURL}/fields/${collectionName}/${fieldName}`, fieldData, {
        headers: this.getAuthHeaders(),
      });
      return response.data.data;
    } catch (error) {
      console.error('Error updating field:', error);
      throw error;
    }
  }

  async deleteField(collectionName: string, fieldName: string): Promise<void> {
    try {
      await axios.delete(`${this.baseURL}/fields/${collectionName}/${fieldName}`, {
        headers: this.getAuthHeaders(),
      });
    } catch (error) {
      console.error('Error deleting field:', error);
      throw error;
    }
  }
}

const collectionsService = new CollectionsService();
export default collectionsService;
