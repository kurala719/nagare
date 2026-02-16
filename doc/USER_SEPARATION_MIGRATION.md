# User and UserInformation Separation - Database Migration Guide

## Overview
User authentication and profile information have been separated into two distinct entities with separate database tables.

## Database Tables

### 1. `users` Table (Authentication)
Contains only authentication-related fields:
- `id` - Primary key
- `created_at`, `updated_at`, `deleted_at` - GORM timestamps
- `username` - User login name
- `password` - Hashed password
- `privileges` - User role (0=unauthorized, 1=user, 2=admin, 3=superadmin)
- `status` - Account status (0=inactive, 1=active)

### 2. `user_informations` Table (Profile)
Contains user profile/personal information:
- `id` - Primary key
- `created_at`, `updated_at`, `deleted_at` - GORM timestamps
- `user_id` - Foreign key to users.id
- `email` - Email address
- `phone` - Phone number
- `avatar` - Avatar URL
- `address` - Physical address
- `introduction` - User bio/introduction
- `nickname` - Display name

## Migration Steps

### Option 1: Fresh Database (Recommended for Development)
GORM auto-migration will create both tables automatically:
```go
db.AutoMigrate(&domain.User{}, &domain.UserInformation{})
```

### Option 2: Migrating Existing Data
If you have existing users table with all fields mixed together:

```sql
-- 1. Create new user_informations table
CREATE TABLE user_informations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    created_at DATETIME(3),
    updated_at DATETIME(3),
    deleted_at DATETIME(3),
    user_id BIGINT NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(255),
    avatar VARCHAR(255),
    address VARCHAR(255),
    introduction TEXT,
    nickname VARCHAR(255),
    INDEX idx_user_informations_deleted_at (deleted_at),
    INDEX idx_user_informations_user_id (user_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 2. Migrate existing profile data
INSERT INTO user_informations (created_at, updated_at, user_id, email, phone, avatar, address, introduction, nickname)
SELECT created_at, updated_at, id, email, phone, avatar, address, introduction, nickname
FROM users
WHERE email IS NOT NULL OR phone IS NOT NULL OR nickname IS NOT NULL;

-- 3. Remove profile columns from users table (backup first!)
ALTER TABLE users
DROP COLUMN email,
DROP COLUMN phone,
DROP COLUMN avatar,
DROP COLUMN address,
DROP COLUMN introduction,
DROP COLUMN nickname;
```

## API Endpoints

### User API (`/user/*`)
**Authentication** (Public):
- `POST /user/login` - Login
- `POST /user/register` - Register
- `POST /user/reset` - Reset password

**Management** (Admin only - privilege >= 2):
- `GET /user/` - List all users
- `GET /user/search?q=&privileges=&status=` - Search users
- `GET /user/:id` - Get user by ID
- `POST /user/` - Create user
- `PUT /user/:id` - Update user
- `DELETE /user/:id` - Delete user

### User Information API (`/user-information/*`)
**User's Own Profile** (Authenticated - privilege >= 1):
- `GET /user-information/` - Get own profile
- `POST /user-information/` - Create own profile
- `PUT /user-information/` - Update own profile
- `DELETE /user-information/` - Delete own profile

**Admin Access** (Admin only - privilege >= 2):
- `GET /user-information/user/:user_id` - Get any user's profile

## Frontend Changes

### API Files Structure
```
src/api/
├── users.js              # User authentication & management
└── userInformation.js    # User profile/information
```

### Usage Example
```javascript
// Login (users.js)
import { loginUser } from '@/api/users'
await loginUser({ username, password })

// Get profile (userInformation.js)
import { getUserInformation } from '@/api/userInformation'
const profile = await getUserInformation()

// Update profile (userInformation.js)
import { updateUserInformation } from '@/api/userInformation'
await updateUserInformation({ email, phone, nickname, ... })
```

## Benefits of This Architecture

1. **Separation of Concerns**: Authentication is isolated from profile data
2. **Performance**: Profile data is only loaded when needed
3. **Security**: Sensitive auth data is separate from public profile info
4. **Scalability**: Each entity can be extended independently
5. **Flexibility**: Different caching strategies for auth vs profile data
6. **Database Optimization**: Smaller auth table for faster login queries

## Notes

- UserInformation is optional - users can exist without profile information
- Frontend handles 404 gracefully when profile doesn't exist
- Profile creation is automatic on first save attempt
- Deleting a user cascades to delete their profile information
