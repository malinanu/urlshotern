/**
 * RBAC TypeScript Types
 * Type definitions for Role-Based Access Control system
 */

export interface Role {
  id: number;
  name: string;
  display_name: string;
  description: string;
  is_system: boolean;
  created_at?: string;
}

export interface Permission {
  id: number;
  name: string;
  resource: string;
  action: string;
  description: string;
}

export interface CreateUserRequest {
  name: string;
  email: string;
  password?: string;
  phone?: string;
  account_type: 'free' | 'premium' | 'enterprise';
  role_ids: number[];
  send_invite: boolean;
  email_verified: boolean;
  is_active: boolean;
  force_password_reset?: boolean;
}

export interface CreateUserResponse {
  user: {
    id: number;
    name: string;
    email: string;
    is_active: boolean;
    is_email_verified: boolean;
    account_type: string;
    created_at: string;
  };
  roles: Role[];
  invite_url?: string;
  temp_password?: string;
}

export interface UserRoles {
  roles: Role[];
  permissions: Permission[];
}

export interface AssignRoleRequest {
  role_id: number;
}

export interface UpdateAccountTypeRequest {
  account_type: 'free' | 'premium' | 'enterprise';
}
