export interface RegisterUserModel {
  name?: string;
  email?: string;
  phone?: string;
  role?: string;
}

export interface LoginUserModel {
  email?: string;
}

export interface UpdateUserModel {
  name?: string;
  phone?: string;
}

export interface UpdateUserResponseModel {
  status: number;
  user: UserModel;
}

export interface UserModel {
  id: string;
  email: string;
  name: string;
  phone: string;
  role: string;
  refresh_token: string;
}
export interface LoginUserResponseModel {
  status: number;
  access_token: string;
  refresh_token: string;
  user: UserModel;
}
