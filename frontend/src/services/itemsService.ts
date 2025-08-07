import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

export interface Item {
  [key: string]: any;
}

export interface ItemsListResponse {
  data: Item[];
  meta: {
    page: number;
    limit: number;
    total: number;
  };
}

class ItemsService {
  private getAuthHeaders() {
    const token = localStorage.getItem('access_token');
    return {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    };
  }

  async getItems(collectionName: string, page: number = 1, limit: number = 50): Promise<ItemsListResponse> {
    try {
      const response = await axios.get(`${API_BASE_URL}/items/${collectionName}`, {
        headers: this.getAuthHeaders(),
        params: { page, limit },
      });
      return response.data;
    } catch (error) {
      console.error(`Error fetching items for ${collectionName}:`, error);
      throw error;
    }
  }
}

const itemsService = new ItemsService();
export default itemsService;
