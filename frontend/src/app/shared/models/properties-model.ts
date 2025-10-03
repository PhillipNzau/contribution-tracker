export interface CreatePropertyModel {
  title?: string;
  description?: string;
  location?: string;
  price?: number;
  images?: string[];
  availability?: boolean;
}

export interface UpdatePropertyModel {
  title?: string;
  // phone?: string;
}

export interface PropertyResponseModel {
  id?: string;
  user_id?: string;
  title?: string;
  description?: string;
  location?: string;
  price?: number;
  images?: string[];
  available?: boolean;
  housekeepers?: string[];
  created_at?: string;
  updated_at?: string;
}

export interface PropertyUpdateResponseModel {
  message?: string;
  property?: PropertyResponseModel;
}
